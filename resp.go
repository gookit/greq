package hreq

import (
	"net/http"

	"github.com/gookit/goutil/netutil/httpreq"
)

// Response is a http.Response wrapper
type Response struct {
	*http.Response
	decoder RespDecoder
}

// IsOK check response status code is 200
func (r *Response) IsOK() bool {
	return httpreq.IsOK(r.StatusCode)
}

// Decode get the raw http.Response
func (r *Response) Decode(ptr interface{}) error {
	return r.decoder.Decode(r.Response, ptr)
}

// Result get the raw http.Response
func (r *Response) Result() *http.Response {
	return r.Response
}
