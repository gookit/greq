package greq

import (
	"context"
	"net/http"
	gourl "net/url"
	"time"

	"github.com/gookit/goutil/netutil/httpreq"
)

// Options for a request build
type Options struct {
	// url or path for current request
	pathURL string

	// Method for request
	Method string
	// ContentType header
	ContentType string
	// Headers for request
	Header http.Header
	// HeaderM map string data.
	HeaderM map[string]string

	// Query params data.
	Query  gourl.Values
	QueryM map[string]any

	// Data for request, will be encoded to query string or req body.
	//
	// type allow: string, []byte, io.Reader, io.ReadCloser, url.Values, map[string]string
	Data any
	// Body data for request, only for POST, PUT, PATCH
	//
	// type allow: string, []byte, io.Reader, io.ReadCloser, url.Values, map[string]string
	Body any
	// Provider body data provider, can with custom content-type
	Provider BodyProvider

	// EncodeJSON req body
	EncodeJSON bool
	// Timeout unit: ms
	Timeout int
	// TCancelFn will auto set it on Timeout > 0
	TCancelFn context.CancelFunc
	// Context for request
	Context context.Context
	// Logger for request
	Logger httpreq.ReqLogger
}

// OptionFn for config request options
type OptionFn func(opt *Options)

// NewOpt for request
func NewOpt(fns ...OptionFn) *Options {
	return NewOpt2(fns, "")
}

// NewOpt2 for request
func NewOpt2(fns []OptionFn, method string) *Options {
	opt := &Options{
		Method:  method,
		Header:  make(http.Header),
		HeaderM: make(map[string]string),
		Query:   make(gourl.Values),
	}
	for _, fn := range fns {
		fn(opt)
	}
	return opt
}

func orCreate(opt *Options) *Options {
	if opt == nil {
		opt = &Options{}
	}
	return opt
}

func ensureOpt(opt *Options) *Options {
	if opt == nil {
		opt = &Options{}
	}
	if opt.Context == nil {
		opt.Context = context.Background()
	}

	if opt.Timeout > 0 && opt.TCancelFn == nil {
		opt.Context, opt.TCancelFn = context.WithTimeout(
			opt.Context,
			time.Duration(opt.Timeout)*time.Millisecond,
		)
	}
	return opt
}

//
// ----------- built in OptionFn ------------
//

// WithMethod set method
func WithMethod(method string) OptionFn {
	return func(opt *Options) {
		if method != "" {
			opt.Method = method
		}
	}
}

// WithContentType set content-type
func WithContentType(contentType string) OptionFn {
	return func(opt *Options) {
		opt.ContentType = contentType
	}
}

// WithUserAgent set user-agent header
func WithUserAgent(userAgent string) OptionFn {
	return func(opt *Options) {
		opt.Header.Set("User-Agent", userAgent)
	}
}
