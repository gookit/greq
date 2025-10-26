// Package batch provides utilities for batch processing HTTP requests
//  - supports concurrent execution with configurable concurrency and timeout
//  - provides methods for adding requests and executing them in batches
package batch

import (
	"strconv"
	"time"

	"github.com/gookit/greq"
)

// Request represents a single request in a batch
type Request struct {
	// ID is an optional identifier for the request
	ID string
	// Method is the HTTP method (GET, POST, etc.)
	Method string
	// URL is the full request URL
	URL string
	// Body is the request body (optional)
	Body any
	// Options are the request options
	Options []greq.OptionFn
}

// Result represents the result of a single batch request
type Result struct {
	// ID is the request identifier
	ID string
	// Request is the original request
	Request *Request
	// Response is the HTTP response (nil if error occurred)
	Response *greq.Response
	// Error is any error that occurred during the request
	Error error
	// Duration is the time taken to complete the request
	Duration time.Duration
}

// Results is a map of request ID to Result
type Results map[string]*Result

//
// region Quickly request
// -----------------------------------

// ExecuteAll executes all requests using the standard client and returns results
func ExecuteAll(requests []*Request, optFns ...ProcessOptionFn) Results {
	bp := NewProcessor(optFns...)
	for _, req := range requests {
		bp.AddRequest(req)
	}
	return bp.ExecuteAll()
}

// ExecuteAny executes requests using the standard client and returns first successful result
func ExecuteAny(requests []*Request, optFns ...ProcessOptionFn) *Result {
	bp := NewProcessor(optFns...)
	for _, req := range requests {
		bp.AddRequest(req)
	}
	return bp.ExecuteAny()
}

// GetAll executes multiple GET requests and waits for all to complete
//  - each request will be assigned an ID as "id_<index>"
func GetAll(urls []string, optFns ...greq.OptionFn) Results {
	bp := NewProcessor()
	for i, url := range urls {
		bp.AddGet("id_"+strconv.Itoa(i), url, optFns...)
	}
	return bp.ExecuteAll()
}

// GetAny executes multiple GET requests and returns first successful result
//  - each request will be assigned an ID as "id_<index>"
func GetAny(urls []string, optFns ...greq.OptionFn) *Result {
	bp := NewProcessor()
	for i, url := range urls {
		bp.AddGet("id_"+strconv.Itoa(i), url, optFns...)
	}
	return bp.ExecuteAny()
}

// PostAll executes multiple POST requests and waits for all to complete
//  - each request will be assigned an ID as "id_<index>"
func PostAll(urls []string, bodies []any, optFns ...greq.OptionFn) Results {
	if len(urls) != len(bodies) {
		panic("urls and bodies must have the same length")
	}

	bp := NewProcessor()
	for i, url := range urls {
		bp.AddPost("id_"+strconv.Itoa(i), url, bodies[i], optFns...)
	}
	return bp.ExecuteAll()
}
