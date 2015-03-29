package router

import (
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type Handler func(http.ResponseWriter, *http.Request) *HandlerError

type HandlerWithParams func(http.ResponseWriter, *http.Request, httprouter.Params) *HandlerError

type HandlerError struct {
	Err     error
	Message string
	Code    int
}

func (r Router) GET(path string, handler Handler) {
	r.Handler("GET", path, handleErrors(handler))
}

func (r Router) POST(path string, handler Handler) {
	r.Handler("POST", path, handleErrors(handler))
}

func (r Router) PUTWithParams(path string, handler HandlerWithParams) {
	r.PUT(path, handleErrorsWithParams(handler))
}

// Router is a wrapper around julienschmidt/httprouter that implements error
// handling specific to this application.
type Router struct {
	*httprouter.Router
	prefix string
}

// NewWithPrefix creates a Router that adds an option to use a common prefix for
// all the paths.
func NewWithPrefix(prefix string) *Router {
	return &Router{
		Router: httprouter.New(),
		prefix: prefix,
	}
}

func (r Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if r.prefix == "" {
		r.ServeHTTP(w, req)
		return
	}
	http.StripPrefix(r.prefix, r).ServeHTTP(w, req)
}

func handleErrors(h Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if e := h(w, r); e != nil {
			handleError(w, e)
		}
	})
}

func handleErrorsWithParams(h HandlerWithParams) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		if e := h(w, r, ps); e != nil {
			handleError(w, e)
		}
	}
}

func handleError(w http.ResponseWriter, e *HandlerError) {
	log.Print(e.Err)
	http.Error(w, e.Message, e.Code)
}

func (e *HandlerError) Error() string {
	return e.Err.Error()
}