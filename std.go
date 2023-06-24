package greq

import (
	"net/http"
	"net/url"

	"github.com/gookit/goutil"
)

// std instance
var std = New()

// Std instance
func Std() *Client {
	return std
}

// Reset std default settings
func Reset() *Client {
	std.header = make(http.Header)
	std.query = make(url.Values, 0)
	return std
}

// BaseURL set base URL for request
func BaseURL(baseURL string) *Client {
	return std.BaseURL(baseURL)
}

// GetDo sets the method to GET and sets the given pathURL, then send request and return response.
func GetDo(pathURL string, optFns ...OptionFn) (*Response, error) {
	return std.SendWithOpt(pathURL, NewOpt2(optFns, http.MethodGet))
}

// PostDo sets the method to POST and sets the given pathURL,
// then send request and return http response.
func PostDo(pathURL string, data any, optFns ...OptionFn) (*Response, error) {
	opt := NewOpt2(optFns, http.MethodPost)
	opt.Body = data

	return std.SendWithOpt(pathURL, opt)
}

// PutDo sets the method to PUT and sets the given pathURL,
// then send request and return http response.
func PutDo(pathURL string, data any, optFns ...OptionFn) (*Response, error) {
	opt := NewOpt2(optFns, http.MethodPut)
	opt.Body = data

	return std.SendWithOpt(pathURL, opt)
}

// PatchDo sets the method to PATCH and sets the given pathURL,
// then send request and return http response.
func PatchDo(pathURL string, data any, optFns ...OptionFn) (*Response, error) {
	opt := NewOpt2(optFns, http.MethodPatch)
	opt.Body = data

	return std.SendWithOpt(pathURL, opt)
}

// DeleteDo sets the method to DELETE and sets the given pathURL,
// then send request and return http response.
func DeleteDo(pathURL string, optFns ...OptionFn) (*Response, error) {
	return std.SendWithOpt(pathURL, NewOpt2(optFns, http.MethodDelete))
}

// HeadDo sets the method to HEAD and request the pathURL, then send request and return response.
func HeadDo(pathURL string, optFns ...OptionFn) (*Response, error) {
	return std.SendWithOpt(pathURL, NewOpt2(optFns, http.MethodHead))
}

// TraceDo sets the method to TRACE and sets the given pathURL,
// then send request and return http response.
func TraceDo(pathURL string, optFns ...OptionFn) (*Response, error) {
	return std.SendWithOpt(pathURL, NewOpt2(optFns, http.MethodTrace))
}

// OptionsDo sets the method to OPTIONS and request the pathURL,
// then send request and return response.
func OptionsDo(pathURL string, optFns ...OptionFn) (*Response, error) {
	return std.SendWithOpt(pathURL, NewOpt2(optFns, http.MethodOptions))
}

// ConnectDo sets the method to CONNECT and sets the given pathURL,
// then send request and return http response.
func ConnectDo(pathURL string, optFns ...OptionFn) (*Response, error) {
	return std.SendWithOpt(pathURL, NewOpt2(optFns, http.MethodConnect))
}

// SendDo sets the method to POST and sets the given pathURL,
// then send request and return http response.
func SendDo(method, pathURL string, optFns ...OptionFn) (*Response, error) {
	return std.SendWithOpt(pathURL, NewOpt2(optFns, method))
}

// MustDo sets the method to POST and sets the given pathURL,
// then send request and return http response.
func MustDo(method, pathURL string, optFns ...OptionFn) *Response {
	resp, err := std.SendWithOpt(pathURL, NewOpt2(optFns, method))
	if err != nil {
		goutil.Panicf("send request error: %s", err.Error())
	}
	return resp
}
