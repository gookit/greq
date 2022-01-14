package hreq

import "net/http"

// Middleware interface for client request.
type Middleware interface {
	Handle(r *http.Request, next NextFunc) (*http.Response, error)
}

// NextFunc implements the Middleware interface
type NextFunc func(r *http.Request) (*http.Response, error)

// MiddleFunc implements the Middleware interface
type MiddleFunc func(r *http.Request, next NextFunc) (*http.Response, error)

// Handle request
func (mf MiddleFunc) Handle(r *http.Request, next NextFunc) (*http.Response, error) {
	return mf(r, next)
}
