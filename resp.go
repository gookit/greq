package hreq

import (
	"net/http"

	"github.com/gookit/goutil/netutil/httpreq"
)

// Response is a http.Response wrapper
type Response struct {
	*http.Response
}

// IsOK check response status code is 200
func (r *Response) IsOK() bool {
	return httpreq.IsOK(r.StatusCode)
}

// Result get the raw http.Response
func (r *Response) Result() *http.Response {
	return r.Response
}
