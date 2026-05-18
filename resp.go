package greq

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"

	"github.com/gookit/goutil/fsutil"
	"github.com/gookit/goutil/netutil/httpctype"
	"github.com/gookit/goutil/netutil/httpreq"
)

// Response is a http.Response wrapper, add some useful methods.
type Response struct {
	// raw http.Response
	*http.Response
	// CostTime for a request-response. unit: ms
	CostTime int64
	// decoder for response, default will extends from Client.respDecoder
	decoder RespDecoder
}

// NewResponse create a new Response instance
func NewResponse(resp *http.Response, decoder RespDecoder) *Response {
	return &Response{Response: resp, decoder: decoder}
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

// IsEmptyBody reports whether the response body is known to be empty.
// Returns true only when Content-Length is explicitly 0; chunked responses
// (Content-Length == -1) are NOT treated as empty because their actual size
// isn't known until the stream is consumed.
func (r *Response) IsEmptyBody() bool {
	return r.ContentLength == 0
}

// ContentType get response content type
func (r *Response) ContentType() string {
	return r.Header.Get(httpctype.Key)
}

// IsContentType check response content type is equals the given.
//
// Usage:
//
//	resp, err := greq.Post("some.host/path")
//	ok := resp.IsContentType("application/xml")
func (r *Response) IsContentType(prefix string) bool {
	val := r.Header.Get(httpctype.Key)
	return val != "" && strings.HasPrefix(val, prefix)
}

// IsJSONType check response content type is JSON
func (r *Response) IsJSONType() bool {
	return r.IsContentType(httpctype.MIMEJSON)
}

// Decode get the raw http.Response
func (r *Response) Decode(ptr any) error {
	defer r.QuietCloseBody()
	return r.decoder.Decode(r.Response, ptr)
}

// SetDecoder for response
func (r *Response) SetDecoder(decoder RespDecoder) {
	r.decoder = decoder
}

// BodyBuffer reads body to buffer. Panics on read error — kept for backward
// compatibility. Prefer BodyBufferE for new code so a transient network
// error during read doesn't crash the program.
//
// NOTICE: closes resp body.
func (r *Response) BodyBuffer() *bytes.Buffer {
	buf, err := r.BodyBufferE()
	if err != nil {
		panic(err)
	}
	return buf
}

// BodyBufferE reads body to buffer and returns any read error.
//
// NOTICE: closes resp body.
func (r *Response) BodyBufferE() (*bytes.Buffer, error) {
	buf := &bytes.Buffer{}
	if r.ContentLength > bytes.MinRead {
		buf.Grow(int(r.ContentLength) + 2)
	}

	defer r.QuietCloseBody()
	if _, err := buf.ReadFrom(r.Body); err != nil {
		return buf, err
	}
	return buf, nil
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

// BodyString reads response body as string. Panics on read error;
// prefer BodyStringE for new code.
func (r *Response) BodyString() string {
	return r.BodyBuffer().String()
}

// BodyStringE reads response body as string and returns any read error.
func (r *Response) BodyStringE() (string, error) {
	buf, err := r.BodyBufferE()
	return buf.String(), err
}

// SaveFile writes the response body to file. Returns an error (without
// writing anything) if the response status is not 2xx — otherwise a 404
// HTML page or 500 stack trace would silently get saved as the "file".
//
// Call SaveBody if you really want to save regardless of status.
func (r *Response) SaveFile(file string) (n int, err error) {
	if r.IsFail() {
		r.QuietCloseBody()
		return 0, fmt.Errorf("greq: refuse to save body on non-2xx response (status %d)", r.StatusCode)
	}
	return r.SaveBody(file)
}

// SaveBody writes the response body to file unconditionally, regardless of
// status code. Use this when you intend to capture error pages too.
func (r *Response) SaveBody(file string) (n int, err error) {
	if r.Body != nil {
		defer r.QuietCloseBody()
		n, err = fsutil.PutContents(file, r.Body)
	}
	return
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

// CloseBody close resp body
func (r *Response) CloseBody() error {
	if r.Body != nil {
		return r.Body.Close()
	}
	return nil
}

// QuietCloseBody close resp body, ignore error
func (r *Response) QuietCloseBody() {
	_ = r.CloseBody()
}
