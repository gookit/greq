package hreq

import "net/http"

// Response is a http.Response wrapper
type Response struct {
	result *http.Response
}

// Result get the raw http.Response
func (r *Response) Result() *http.Response {
	return r.result
}
