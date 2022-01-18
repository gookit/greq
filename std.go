package hreq

import "net/http"

// std instance
var std = New()

// Std instance
func Std() *HReq {
	return std
}

// BaseURL set base URL for request
func BaseURL(baseURL string) *HReq {
	return std.BaseURL(baseURL)
}

// Get sets the method to GET and sets the given pathURL, then send request and return response.
func Get(pathURL string) (*Response, error) {
	return std.Method(http.MethodGet).Send(pathURL)
}

// Post sets the method to POST and sets the given pathURL,
// then send request and return http response.
func Post(pathURL string) (*Response, error) {
	return std.Method(http.MethodPost).Send(pathURL)
}

// Put sets the method to PUT and sets the given pathURL,
// then send request and return http response.
func Put(pathURL string) (*Response, error) {
	return std.Method(http.MethodPut).Send(pathURL)
}

// Patch sets the method to PATCH and sets the given pathURL,
// then send request and return http response.
func Patch(pathURL string) (*Response, error) {
	return std.Method(http.MethodPatch).Send(pathURL)
}

// Delete sets the method to DELETE and sets the given pathURL,
// then send request and return http response.
func Delete(pathURL string) (*Response, error) {
	return std.Method(http.MethodDelete).Send(pathURL)
}

// Head sets the method to HEAD and request the pathURL, then send request and return response.
func Head(pathURL string) (*Response, error) {
	return std.Method(http.MethodHead).Send(pathURL)
}

// Trace sets the method to TRACE and sets the given pathURL,
// then send request and return http response.
func Trace(pathURL string) (*Response, error) {
	return std.Method(http.MethodTrace).Send(pathURL)
}

// Options sets the method to OPTIONS and request the pathURL,
// then send request and return response.
func Options(pathURL string) (*Response, error) {
	return std.Method(http.MethodOptions).Send(pathURL)
}

// Connect sets the method to CONNECT and sets the given pathURL,
// then send request and return http response.
func Connect(pathURL string) (*Response, error) {
	return std.Method(http.MethodConnect).Send(pathURL)
}
