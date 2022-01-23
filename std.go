package hreq

import (
	"net/http"
	"net/url"
)

// std instance
var std = New()

// Std instance
func Std() *HReq {
	return std
}

// Reset std
func Reset() *HReq {
	std.header = make(http.Header)

	std.bodyProvider = nil
	std.queryParams = make(url.Values, 0)
	return std
}

func reset() *HReq {
	std.bodyProvider = nil
	return std
}

// BaseURL set base URL for request
func BaseURL(baseURL string) *HReq {
	return std.BaseURL(baseURL)
}

// Get sets the method to GET and sets the given pathURL, then send request and return response.
func Get(pathURL string) (*Response, error) {
	return reset().Send(pathURL, http.MethodGet)
}

// Post sets the method to POST and sets the given pathURL,
// then send request and return http response.
func Post(pathURL string, body ...interface{}) (*Response, error) {
	if len(body) > 0 {
		std.Body(body[0])
	} else {
		reset()
	}

	return std.Send(pathURL, http.MethodPost)
}

// Put sets the method to PUT and sets the given pathURL,
// then send request and return http response.
func Put(pathURL string, body ...interface{}) (*Response, error) {
	if len(body) > 0 {
		std.Body(body[0])
	} else {
		reset()
	}

	return std.Send(pathURL, http.MethodPut)
}

// Patch sets the method to PATCH and sets the given pathURL,
// then send request and return http response.
func Patch(pathURL string, body ...interface{}) (*Response, error) {
	if len(body) > 0 {
		std.Body(body[0])
	} else {
		reset()
	}

	return std.Send(pathURL, http.MethodPatch)
}

// Delete sets the method to DELETE and sets the given pathURL,
// then send request and return http response.
func Delete(pathURL string) (*Response, error) {
	reset()
	return std.Send(pathURL, http.MethodDelete)
}

// Head sets the method to HEAD and request the pathURL, then send request and return response.
func Head(pathURL string) (*Response, error) {
	return reset().Send(pathURL, http.MethodHead)
}

// Trace sets the method to TRACE and sets the given pathURL,
// then send request and return http response.
func Trace(pathURL string) (*Response, error) {
	return reset().Send(pathURL, http.MethodTrace)
}

// Options sets the method to OPTIONS and request the pathURL,
// then send request and return response.
func Options(pathURL string) (*Response, error) {
	return reset().Send(pathURL, http.MethodOptions)
}

// Connect sets the method to CONNECT and sets the given pathURL,
// then send request and return http response.
func Connect(pathURL string) (*Response, error) {
	return reset().Send(pathURL, http.MethodConnect)
}
