package greq

import (
	"net/http"
	"net/url"
)

// std instance
var std = New()

// Std instance
func Std() *HiReq {
	return std
}

// Reset std
func Reset() *HiReq {
	std.header = make(http.Header)

	std.bodyProvider = nil
	std.queryParams = make(url.Values, 0)
	return std
}

func reset() *HiReq {
	std.bodyProvider = nil
	return std
}

// BaseURL set base URL for request
func BaseURL(baseURL string) *HiReq {
	return std.BaseURL(baseURL)
}

// GetDo sets the method to GET and sets the given pathURL, then send request and return response.
func GetDo(pathURL string) (*Response, error) {
	return reset().Send(pathURL, http.MethodGet)
}

// PostDo sets the method to POST and sets the given pathURL,
// then send request and return http response.
func PostDo(pathURL string, body ...interface{}) (*Response, error) {
	if len(body) > 0 {
		std.Body(body[0])
	} else {
		reset()
	}

	return std.Send(pathURL, http.MethodPost)
}

// PutDo sets the method to PUT and sets the given pathURL,
// then send request and return http response.
func PutDo(pathURL string, body ...interface{}) (*Response, error) {
	if len(body) > 0 {
		std.Body(body[0])
	} else {
		reset()
	}

	return std.Send(pathURL, http.MethodPut)
}

// PatchDo sets the method to PATCH and sets the given pathURL,
// then send request and return http response.
func PatchDo(pathURL string, body ...interface{}) (*Response, error) {
	if len(body) > 0 {
		std.Body(body[0])
	} else {
		reset()
	}

	return std.Send(pathURL, http.MethodPatch)
}

// DeleteDo sets the method to DELETE and sets the given pathURL,
// then send request and return http response.
func DeleteDo(pathURL string) (*Response, error) {
	return reset().Send(pathURL, http.MethodDelete)
}

// Head sets the method to HEAD and request the pathURL, then send request and return response.
func Head(pathURL string) (*Response, error) {
	return reset().Send(pathURL, http.MethodHead)
}

// TraceDo sets the method to TRACE and sets the given pathURL,
// then send request and return http response.
func TraceDo(pathURL string) (*Response, error) {
	return reset().Send(pathURL, http.MethodTrace)
}

// OptionsDo sets the method to OPTIONS and request the pathURL,
// then send request and return response.
func OptionsDo(pathURL string) (*Response, error) {
	return reset().Send(pathURL, http.MethodOptions)
}

// ConnectDo sets the method to CONNECT and sets the given pathURL,
// then send request and return http response.
func ConnectDo(pathURL string) (*Response, error) {
	return reset().Send(pathURL, http.MethodConnect)
}
