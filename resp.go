package hireq

import (
	"bytes"
	"net/http"
	"strings"

	"github.com/gookit/goutil/netutil/httpctype"
	"github.com/gookit/goutil/netutil/httpreq"
)

// Response is a http.Response wrapper
type Response struct {
	*http.Response
	// decoder for response, default will extends from HiReq.respDecoder
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

// IsEmptyBody check response body is empty
func (r *Response) IsEmptyBody() bool {
	return r.ContentLength <= 0
}

// ContentType get response content type
func (r *Response) ContentType() string {
	return r.Header.Get(httpctype.Key)
}

// IsContentType check response content type is equals the given.
//
// Usage:
//	resp, err := hireq.Post("some.host/path")
//  ok := resp.IsContentType("application/xml")
//
func (r *Response) IsContentType(prefix string) bool {
	val := r.Header.Get(httpctype.Key)
	return val != "" && strings.HasPrefix(val, prefix)
}

// IsJSONType check response content type is JSON
func (r *Response) IsJSONType() bool {
	return r.IsContentType(httpctype.MIMEJSON)
}

// Decode get the raw http.Response
func (r *Response) Decode(ptr interface{}) error {
	defer r.QuietCloseBody()
	return r.decoder.Decode(r.Response, ptr)
}

// SetDecoder for response
func (r *Response) SetDecoder(decoder RespDecoder) {
	r.decoder = decoder
}

// BodyBuffer read body to buffer.
//
// NOTICE: must close resp body.
func (r *Response) BodyBuffer() *bytes.Buffer {
	buf := &bytes.Buffer{}
	// prof: Allocate memory in advance
	if r.ContentLength > bytes.MinRead {
		buf.Grow(int(r.ContentLength) + 2)
	}

	// NOTICE: must close resp body.
	defer r.QuietCloseBody()
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		panic(err)
	}

	return buf
}

// HeaderString convert response headers to string
func (r *Response) HeaderString() string {
	buf := &bytes.Buffer{}
	for key, values := range r.Header {
		buf.WriteString(key)
		buf.WriteString(": ")
		buf.WriteString(strings.Join(values, ";"))
		buf.WriteByte('\n')
	}

	return buf.String()
}

// BodyString convert response body to string
func (r *Response) BodyString() string {
	return r.BodyBuffer().String()
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

	if !r.IsEmptyBody() {
		buf.WriteByte('\n')
		_, _ = buf.ReadFrom(r.Body)
	}

	return buf.String()
}

// Result get the raw http.Response
func (r *Response) Result() *http.Response {
	return r.Response
}

// CloseBody close resp body
func (r *Response) CloseBody() error {
	return r.Body.Close()
}

// QuietCloseBody close resp body, ignore error
func (r *Response) QuietCloseBody() {
	_ = r.Body.Close()
}
