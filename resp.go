package hreq

import (
	"bytes"
	"net/http"
	"strings"

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

// IsSuccessful check response status code is in 200 - 300
func (r *Response) IsSuccessful() bool {
	return httpreq.IsSuccessful(r.StatusCode)
}

// IsFail check response status code != 200
func (r *Response) IsFail() bool {
	return !httpreq.IsOK(r.StatusCode)
}

// Decode get the raw http.Response
func (r *Response) Decode(ptr interface{}) error {
	return r.decoder.Decode(r.Response, ptr)
}

// String convert Response to string
func (r *Response) String() string {
	buf := &bytes.Buffer{}
	buf.WriteString(r.Proto)
	buf.WriteByte(' ')
	buf.WriteString(r.Status)
	buf.WriteByte('\n')

	for key, values := range r.Header {
		buf.WriteString(key)
		buf.WriteString(": ")
		buf.WriteString(strings.Join(values, ";"))
		buf.WriteByte('\n')
	}

	if r.Body != nil {
		buf.WriteByte('\n')
		_, _ = buf.ReadFrom(r.Body)
	}

	return buf.String()
}

// Result get the raw http.Response
func (r *Response) Result() *http.Response {
	return r.Response
}
