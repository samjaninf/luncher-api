package handler

import (
	"net/http"

	"github.com/deiwin/facebook"
	"github.com/deiwin/luncher-api/db"
	. "github.com/deiwin/luncher-api/router"
	"github.com/deiwin/luncher-api/session"
	"golang.org/x/oauth2"
)

type Facebook interface {
	// Login returns a handler that redirects the user to Facebook to log in
	Login() Handler
	// Redirected returns a handler that receives the user and page tokens for the
	// user who has just logged in through Facebook. Updates the user and page
	// access tokens in the DB
	Redirected() Handler
}

type fbook struct {
	auth            facebook.Authenticator
	sessionManager  session.Manager
	usersCollection db.Users
}

func NewFacebook(fbAuth facebook.Authenticator, sessMgr session.Manager, usersCollection db.Users) Facebook {
	return fbook{fbAuth, sessMgr, usersCollection}
}

func (fb fbook) Login() Handler {
	return func(w http.ResponseWriter, r *http.Request) *HandlerError {
		session := fb.sessionManager.GetOrInit(w, r)
		redirectURL := fb.auth.AuthURL(session)
		http.Redirect(w, r, redirectURL, http.StatusSeeOther)
		return nil
	}
}

func (fb fbook) Redirected() Handler {
	return func(w http.ResponseWriter, r *http.Request) *HandlerError {
		session := fb.sessionManager.GetOrInit(w, r)
		tok, err := fb.auth.Token(session, r)
		if err != nil {
			if err == facebook.ErrMissingState {
				return &HandlerError{err, "Expecting a 'state' value", http.StatusBadRequest}
			} else if err == facebook.ErrInvalidState {
				return &HandlerError{err, "Invalid 'state' value", http.StatusForbidden}
			} else if err == facebook.ErrMissingCode {
				return &HandlerError{err, "Expecting a 'code' value", http.StatusBadRequest}
			}
			return &HandlerError{err, "Failed to connect to Facebook", http.StatusInternalServerError}
		}
		userID, err := fb.getUserID(tok)
		if err != nil {
			return &HandlerError{err, "Failed to get the user information from Facebook", http.StatusInternalServerError}
		}
		pageID, err := fb.getPageID(userID)
		if err != nil {
			return &HandlerError{err, "Failed to find a Facebook page related to this user", http.StatusInternalServerError}
		}
		pageAccessToken, err := fb.auth.PageAccessToken(tok, pageID)
		if err != nil {
			if err == facebook.ErrNoSuchPage {
				return &HandlerError{err, "Access denied by Facebook to the managed page", http.StatusForbidden}
			}
			return &HandlerError{err, "Failed to get access to the Facebook page", http.StatusInternalServerError}
		}
		err = fb.storeAccessTokensInDB(userID, tok, pageAccessToken, session)
		if err != nil {
			return &HandlerError{err, "Failed to persist Facebook login information", http.StatusInternalServerError}
		}
		http.Redirect(w, r, "/#/admin", http.StatusSeeOther)
		return nil
	}
}

func (fb fbook) storeAccessTokensInDB(userID string, tok *oauth2.Token, pageAccessToken, sessionID string) error {
	err := fb.usersCollection.SetAccessToken(userID, *tok)
	if err != nil {
		return err
	}
	err = fb.usersCollection.SetPageAccessToken(userID, pageAccessToken)
	if err != nil {
		return err
	}
	return fb.usersCollection.SetSessionID(userID, sessionID)
}

func (fb fbook) getUserID(tok *oauth2.Token) (string, error) {
	api := fb.auth.APIConnection(tok)
	user, err := api.Me()
	if err != nil {
		return "", err
	}
	return user.ID, nil
}

func (fb fbook) getPageID(userID string) (string, error) {
	userInDB, err := fb.usersCollection.GetFbID(userID)
	if err != nil {
		return "", err
	}
	return userInDB.FacebookPageID, nil
}
