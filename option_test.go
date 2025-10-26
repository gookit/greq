package greq_test

import (
	"net/http"
	"testing"

	"github.com/gookit/goutil/testutil/assert"
	"github.com/gookit/greq"
)

func TestOptions_Struct(t *testing.T) {
	opt := greq.NewOpt()
	assert.NotNil(t, opt)
	assert.NotNil(t, opt.Header)
	assert.NotNil(t, opt.HeaderM)
	assert.NotNil(t, opt.Query)
}

func TestNewOpt(t *testing.T) {
	opt := greq.NewOpt()
	assert.NotNil(t, opt)
	assert.NotNil(t, opt.Header)
	assert.NotNil(t, opt.Query)
}

func TestNewOpt2(t *testing.T) {
	opt := greq.NewOpt2(nil, "POST")
	assert.NotNil(t, opt)
	assert.Equal(t, "POST", opt.Method)

	// Test with option functions
	opt = greq.NewOpt2([]greq.OptionFn{
		greq.WithMethod("GET"),
		greq.WithContentType("application/json"),
	}, "POST")

	assert.Equal(t, "GET", opt.Method) // Should be overridden by option function
	assert.Equal(t, "application/json", opt.ContentType)
}

func TestWithMethod(t *testing.T) {
	fn := greq.WithMethod("POST")
	opt := &greq.Options{}
	fn(opt)
	assert.Equal(t, "POST", opt.Method)

	// Test with empty method (should not change)
	fn = greq.WithMethod("")
	opt = &greq.Options{Method: "GET"}
	fn(opt)
	assert.Equal(t, "GET", opt.Method) // Should remain unchanged
}

func TestWithContentType(t *testing.T) {
	fn := greq.WithContentType("application/json")
	opt := &greq.Options{}
	fn(opt)
	assert.Equal(t, "application/json", opt.ContentType)
}

func TestWithUserAgent(t *testing.T) {
	fn := greq.WithUserAgent("test-agent/1.0")
	opt := &greq.Options{
		Header: make(http.Header),
	}
	fn(opt)
	assert.Equal(t, "test-agent/1.0", opt.Header.Get("User-Agent"))
}

func TestWithHeader(t *testing.T) {
	fn := greq.WithHeader("X-Custom", "value")
	opt := &greq.Options{
		Header: make(http.Header),
	}
	fn(opt)
	assert.Equal(t, "value", opt.Header.Get("X-Custom"))
}

func TestWithBody(t *testing.T) {
	fn := greq.WithBody("test body")
	opt := &greq.Options{}
	fn(opt)
	assert.Equal(t, "test body", opt.Body)
}

func TestWithData(t *testing.T) {
	fn := greq.WithData("test data")
	opt := &greq.Options{}
	fn(opt)
	assert.Equal(t, "test data", opt.Data)
}

func TestWithTimeout(t *testing.T) {
	// Test that timeout is set correctly
	fn := greq.WithTimeout(5000)
	opt := &greq.Options{}
	fn(opt)
	assert.Equal(t, 5000, opt.Timeout)
}