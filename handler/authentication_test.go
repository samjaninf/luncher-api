package handler_test

import (
	"net/http"

	"github.com/Lunchr/luncher-api/db"
	. "github.com/Lunchr/luncher-api/handler"
	"github.com/Lunchr/luncher-api/router"
	"github.com/Lunchr/luncher-api/session"
	"gopkg.in/mgo.v2/bson"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Authentication Handlers", func() {
	Describe("Logout", func() {
		var (
			usersCollection db.Users
			sessionManager  session.Manager
			handler         router.Handler
		)

		BeforeEach(func() {
			usersCollection = &mockUsers{}
		})

		JustBeforeEach(func() {
			handler = Logout(sessionManager, usersCollection)
		})

		ExpectUserToBeLoggedIn(func() *router.HandlerError {
			return handler(responseRecorder, request)
		}, func(mgr session.Manager, users db.Users) {
			sessionManager = mgr
			usersCollection = users
		})

		Context("with user logged in", func() {
			BeforeEach(func() {
				sessionManager = &mockSessionManager{isSet: true, id: "correctSession"}
			})

			It("should redirect to root", func() {
				handler(responseRecorder, request)
				ExpectLocationToBeMockedURL(responseRecorder, "/")
			})
		})
	})
})

func ExpectUserToBeLoggedIn(handler func() *router.HandlerError, setDependencies func(session.Manager, db.Users)) {
	Describe("it expects the user to be logged in", func() {
		Context("with no session set", func() {
			BeforeEach(func() {
				setDependencies(&mockSessionManager{}, nil)
			})

			It("should be forbidden", func() {
				err := handler()
				Expect(err.Code).To(Equal(http.StatusUnauthorized))
			})
		})

		Context("with session set, but no matching user in DB", func() {
			BeforeEach(func() {
				setDependencies(&mockSessionManager{isSet: true}, mockUsers{})
			})

			It("should be forbidden", func() {
				err := handler()
				Expect(err.Code).To(Equal(http.StatusUnauthorized))
			})
		})
	})
}

func (m mockUsers) UnsetSessionID(id bson.ObjectId) error {
	Expect(id).To(Equal(objectID))
	return nil
}
