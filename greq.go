// Package greq is a simple http client request builder, support batch requests.
package greq

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"net"
	"net/http"
	"time"
)

// DefaultDoer for request.
var DefaultDoer = http.DefaultClient

// BodyProvider provides Body content for http.Request attachment.
//
// Concrete implementations live in internal/bodyprovider; users may also
// implement this interface themselves and pass it via Builder.BodyProvider.
type BodyProvider interface {
	// ContentType returns the Content-Type of the body.
	ContentType() string
	// Body returns the io.Reader body.
	Body() (io.Reader, error)
}

// HandleFunc for the Middleware
type HandleFunc func(r *http.Request) (*Response, error)

// AfterSendFn callback func
type AfterSendFn func(resp *Response, err error)

// RetryChecker function type for checking if a request should be retried
type RetryChecker func(resp *Response, err error, attempt int) bool

// DefaultRetryChecker is the default retry condition checker
// It retries on:
// - Network errors (err != nil)
// - 5xx server errors
// - 429 Too Many Requests
func DefaultRetryChecker(resp *Response, err error, attempt int) bool {
	// Retry on network errors
	if err != nil {
		return true
	}

	// Retry on server errors (5xx)
	if resp.StatusCode >= 500 && resp.StatusCode < 600 {
		return true
	}

	// Retry on rate limiting (429)
	if resp.StatusCode == 429 {
		return true
	}

	return false
}

// RequestCreator interface
type RequestCreator interface {
	NewRequest(method, target string, body io.Reader) *http.Request
}

// RequestCreatorFunc func
type RequestCreatorFunc func(method, target string, body io.Reader) *http.Request

// Must return response, if error will panic
func Must(w *Response, err error) *Response {
	if err != nil {
		panic(err)
	}
	return w
}

// NewTransport create new http transport
func NewTransport(onCreate func(ht *http.Transport)) *http.Transport {
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          500,
		MaxConnsPerHost:       200,
		MaxIdleConnsPerHost:   100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	if onCreate != nil {
		onCreate(transport)
	}
	return transport
}

//
// region Middleware
// ------------------------------

// Middleware interface for cli request.
type Middleware interface {
	Handle(r *http.Request, next HandleFunc) (*Response, error)
}

// MiddleFunc implements the Middleware interface
type MiddleFunc func(r *http.Request, next HandleFunc) (*Response, error)

// Handle request
func (mf MiddleFunc) Handle(r *http.Request, next HandleFunc) (*Response, error) {
	return mf(r, next)
}

// wrap middlewares, and will wrap http.Response to Response
func (h *Client) wrapMiddlewares() {
	// set core handler
	h.handler = func(r *http.Request) (*Response, error) {
		rawResp, err := h.doer.Do(r)
		if err != nil {
			return nil, err
		}
		return NewResponse(rawResp, h.RespDecoder), nil
	}

	for _, m := range h.middles {
		h.wrapMiddleware(m)
	}
}

func (h *Client) wrapMiddleware(m Middleware) {
	next := h.handler

	// wrap handler
	h.handler = func(r *http.Request) (*Response, error) {
		return m.Handle(r, next)
	}
}

//
// region Response decoders
// ------------------------------

// RespDecoder decodes http responses into struct values.
type RespDecoder interface {
	// Decode decodes the response into the value pointed to by ptr.
	Decode(resp *http.Response, ptr any) error
}

// jsonDecoder decodes http response JSON into a JSON-tagged struct value.
type jsonDecoder struct{}

// Decode decodes the Response Body into the value pointed to by ptr.
// Caller must provide a non-nil v and close the resp.Body.
func (d jsonDecoder) Decode(resp *http.Response, ptr any) error {
	return json.NewDecoder(resp.Body).Decode(ptr)
}

// XmlDecoder decodes http response body into a XML-tagged struct value.
type XmlDecoder struct{}

// Decode decodes the Response body into the value pointed to by ptr.
func (d XmlDecoder) Decode(resp *http.Response, ptr any) error {
	return xml.NewDecoder(resp.Body).Decode(ptr)
}
