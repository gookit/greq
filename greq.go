// Package greq is a simple http client request builder, support batch requests.
package greq

import (
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"time"
)

// DefaultDoer for request.
var DefaultDoer = http.DefaultClient

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
	// Customize the Transport to have larger connection pool.
	transport := &http.Transport{
		// Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		// ForceAttemptHTTP2:     true,
		MaxIdleConns:          500,
		MaxConnsPerHost:       200,
		MaxIdleConnsPerHost:   100,
		IdleConnTimeout:       60 * time.Second,
		TLSHandshakeTimeout:   1 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	if onCreate != nil {
		onCreate(transport)
	}
	return transport
}