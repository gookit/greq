// Package greq is a simple http cli request builder, inspired from https://github.com/dghubble/sling
package greq

import (
	"io"
	"net/http"
)

// DefaultDoer for request.
var DefaultDoer = http.DefaultClient

// HandleFunc for the Middleware
type HandleFunc func(r *http.Request) (*Response, error)

// AfterSendFn callback func
type AfterSendFn func(resp *Response, err error)

// RequestCreator interface
type RequestCreator interface {
	NewRequest(method, target string, body io.Reader) *http.Request
}

// RequestCreatorFunc func
type RequestCreatorFunc func(method, target string, body io.Reader) *http.Request

// Must return response, if error will panic
func Must(w *Response, err error) *Response {
	if err != nil {
		panic(err)
	}
	return w
}
