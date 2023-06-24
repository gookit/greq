package greq

import (
	"bytes"
	"io"
	"net/http"
	gourl "net/url"
	"os"
	"strings"

	"github.com/gookit/goutil/basefn"
	"github.com/gookit/goutil/fsutil"
	"github.com/gookit/goutil/netutil/httpctype"
	"github.com/gookit/goutil/netutil/httpheader"
	"github.com/gookit/goutil/netutil/httpreq"
	"github.com/gookit/goutil/strutil"
)

const (
	HeaderUAgent = "User-Agent"
	HeaderAuth   = "Authorization"

	AgentCURL = "CURL/7.64.1 greq/1.0.2"
)

// Builder is a http request builder.
type Builder struct {
	*Options
	// client for send request. if not set, will use default client.
	cli *Client
}

// NewBuilder for request
func NewBuilder(fns ...OptionFn) *Builder {
	opt := &Options{
		// Method: http.MethodGet,
		Header:  make(http.Header),
		HeaderM: make(map[string]string),
		Query:   make(gourl.Values),
	}
	for _, fn := range fns {
		fn(opt)
	}
	return &Builder{Options: opt}
}

// BuilderWithClient create a new builder with client
func BuilderWithClient(c *Client, optFns ...OptionFn) *Builder {
	return NewBuilder(optFns...).WithClient(c)
}

func newBuilder(c *Client, method, pathURL string) *Builder {
	b := NewBuilder()
	b.cli = c
	b.Method = method
	b.pathURL = pathURL
	return b
}

// WithClient set cli to builder
func (b *Builder) WithClient(c *Client) *Builder {
	b.cli = c
	return b
}

// WithOptionFn set option fns to builder
func (b *Builder) WithOptionFn(fns ...OptionFn) *Builder {
	return b.WithOptionFns(fns)
}

// WithOptionFns set option fns to builder
func (b *Builder) WithOptionFns(fns []OptionFn) *Builder {
	for _, fn := range fns {
		fn(b.Options)
	}
	return b
}

// WithMethod set request method name.
func (b *Builder) WithMethod(method string) *Builder {
	b.Method = method
	return b
}

// PathURL set path URL for current request
func (b *Builder) PathURL(pathURL string) *Builder {
	b.pathURL = pathURL
	return b
}

//
//
// ----------- URL, Query params ------------
//
//

// AddQuery appends new k-v param to the Query string.
func (b *Builder) AddQuery(key string, value any) *Builder {
	b.Query.Add(key, strutil.SafeString(value))
	return b
}

// QueryParams appends url.Values/map[string]string to the Query string.
// The value will be encoded as url Query parameters on send requests (see Send()).
func (b *Builder) QueryParams(ps any) *Builder {
	if ps != nil {
		for key, values := range httpreq.ToQueryValues(ps) {
			for _, value := range values {
				b.Query.Add(key, value)
			}
		}
	}

	return b
}

// QueryValues appends url.Values to the Query string.
// The value will be encoded as url Query parameters on new requests (see Send()).
func (b *Builder) QueryValues(values gourl.Values) *Builder {
	return b.QueryParams(values)
}

// WithQuerySMap appends map[string]string to the Query string.
func (b *Builder) WithQuerySMap(smp map[string]string) *Builder {
	return b.QueryParams(smp)
}

//
//
// ----------- HeaderM ------------
//
//

// AddHeader adds the key, value pair in HeaderM, appending values for existing keys
// to the key's values. Header keys are canonicalized.
func (b *Builder) AddHeader(key, value string) *Builder {
	b.Header.Add(key, value)
	return b
}

// SetHeader sets the key, value pair in HeaderM, replacing existing values
// associated with key. Header keys are canonicalized.
func (b *Builder) SetHeader(key, value string) *Builder {
	b.Header.Set(key, value)
	return b
}

// AddHeaders adds all the http.Header values, appending values for existing keys
// to the key's values. Header keys are canonicalized.
func (b *Builder) AddHeaders(headers http.Header) *Builder {
	for key, values := range headers {
		for i := range values {
			b.Header.Add(key, values[i])
		}
	}
	return b
}

// AddHeaderMap sets all the http.Header values, appending values for existing keys
// to the key's values.
func (b *Builder) AddHeaderMap(headers map[string]string) *Builder {
	for key, value := range headers {
		b.Header.Add(key, value)
	}
	return b
}

// SetHeaders sets all the http.Header values, replacing values for existing keys
// to the key's values. Header keys are canonicalized.
func (b *Builder) SetHeaders(headers http.Header) *Builder {
	for key, values := range headers {
		for i := range values {
			if i == 0 {
				b.Header.Set(key, values[i])
			} else {
				b.Header.Add(key, values[i])
			}
		}
	}
	return b
}

// SetHeaderMap sets all the http.Header values, replacing values for existing keys
// to the key's values.
func (b *Builder) SetHeaderMap(headers map[string]string) *Builder {
	for key, value := range headers {
		b.Header.Set(key, value)
	}
	return b
}

// UserAgent set User-Agent header
func (b *Builder) UserAgent(ua string) *Builder {
	b.HeaderM[httpheader.UserAgent] = ua
	return b
}

// UserAuth with user auth header value.
func (b *Builder) UserAuth(value string) *Builder {
	return b.SetHeader(httpheader.UserAuth, value)
}

// BasicAuth sets the Authorization header to use HTTP Basic Authentication
// with the provided username and password.
//
// With HTTP Basic Authentication the provided username and password are not encrypted.
func (b *Builder) BasicAuth(username, password string) *Builder {
	return b.SetHeader(httpheader.UserAuth, httpreq.BuildBasicAuth(username, password))
}

//
//
// ----------- Cookies ------------
//
//

// WithCookies to request
func (b *Builder) WithCookies(hcs ...*http.Cookie) *Builder {
	var sb strings.Builder
	for i, hc := range hcs {
		if i > 0 {
			sb.WriteByte(';')
		}
		sb.WriteString(hc.String())
	}

	return b.WithCookieString(sb.String())
}

// WithCookieString set cookie header value.
//
// Usage:
//
//	h.NewOpt().
//		WithCookieString("name=inhere;age=30").
//		Do("/some/api", "GET")
func (b *Builder) WithCookieString(value string) *Builder {
	// return h.AddHeader("Set-Cookie", value) // response
	return b.AddHeader("Cookie", value)
}

//
//
// ----------- Content Type ------------
//
//

// WithContentType with custom ContentType header
func (b *Builder) WithContentType(value string) *Builder {
	b.ContentType = value
	return b
}

// WithType with custom ContentType header
func (b *Builder) WithType(value string) *Builder {
	b.ContentType = value
	return b
}

// XMLType with xml Content-Type header
func (b *Builder) XMLType() *Builder {
	b.ContentType = httpctype.XML
	return b
}

// JSONType with json Content-Type header
func (b *Builder) JSONType() *Builder {
	b.ContentType = httpctype.JSON
	return b
}

// FormType with from Content-Type header
func (b *Builder) FormType() *Builder {
	b.ContentType = httpctype.Form
	return b
}

// MultipartType with multipart/form-data Content-Type header
func (b *Builder) MultipartType() *Builder {
	b.ContentType = httpctype.FormData
	return b
}

//
//
// ----------- Request Body ------------
//
//

// WithBody with custom any type body
func (b *Builder) WithBody(body any) *Builder {
	return b.AnyBody(body)
}

// AnyBody with custom any type body
func (b *Builder) AnyBody(body any) *Builder {
	if bp, ok := body.(BodyProvider); ok {
		return b.BodyProvider(bp)
	}

	b.Body = body
	return b
}

// BodyProvider with custom body provider
func (b *Builder) BodyProvider(bp BodyProvider) *Builder {
	b.Provider = bp
	return b
}

// BodyReader with custom io reader body
func (b *Builder) BodyReader(r io.Reader) *Builder {
	b.Provider = bodyProvider{body: r}
	return b
}

// FileContentsBody read file contents as body
func (b *Builder) FileContentsBody(filePath string) *Builder {
	file, err := os.OpenFile(filePath, os.O_RDONLY, fsutil.DefaultFilePerm)
	if err != nil {
		panic(err)
	}
	return b.BodyReader(file)
}

// JSONBody with JSON data body
func (b *Builder) JSONBody(jsonData any) *Builder {
	b.Provider = jsonBodyProvider{
		payload: jsonData,
	}
	return b
}

// FormBody with form data body
func (b *Builder) FormBody(formData any) *Builder {
	b.Provider = formBodyProvider{
		payload: formData,
	}
	return b
}

// BytesBody with custom string body
func (b *Builder) BytesBody(bs []byte) *Builder {
	return b.BodyReader(bytes.NewReader(bs))
}

// StringBody with custom string body
func (b *Builder) StringBody(s string) *Builder {
	return b.BodyReader(strings.NewReader(s))
}

// Multipart with custom multipart body
func (b *Builder) Multipart(key, value string) *Builder {
	// TODO
	return b
}

//
//
// ----------- Build Request ------------
//
//

// Build request
func (b *Builder) Build(method, pathURL string) (*http.Request, error) {
	b.Method = method
	cli := basefn.OrValue(b.cli == nil, std, b.cli)
	return cli.NewRequestWithOptions(pathURL, b.Options)
}

// String request to string.
func (b *Builder) String() string {
	r, err := b.Build(b.Method, b.pathURL)
	if err != nil {
		return ""
	}
	return httpreq.RequestToString(r)
}

//
//
// ----------- Send Request ------------
//
//

// Get sets the method to GET and sets the given pathURL
func (b *Builder) Get(pathURL string, optFns ...OptionFn) *Builder {
	b.pathURL = pathURL
	b.Method = http.MethodGet
	b.WithOptionFns(optFns)
	return b
}

// GetDo sets the method to GET and sets the given pathURL,
// then send request and return response.
func (b *Builder) GetDo(pathURL string, optFns ...OptionFn) (*Response, error) {
	return b.Send(http.MethodGet, pathURL, optFns...)
}

// Post sets the method to POST and sets the given pathURL
func (b *Builder) Post(pathURL string, data any, optFns ...OptionFn) *Builder {
	b.pathURL = pathURL
	b.Method = http.MethodPost
	b.Body = data
	b.WithOptionFns(optFns)
	return b
}

// PostDo sets the method to POST and sets the given pathURL,
// then send request and return http response.
func (b *Builder) PostDo(pathURL string, data any, optFns ...OptionFn) (*Response, error) {
	return b.Post(pathURL, data, optFns...).Do()
}

// Put sets the method to PUT and sets the given pathURL
func (b *Builder) Put(pathURL string, data any, optFns ...OptionFn) *Builder {
	b.Body = data
	b.pathURL = pathURL
	b.Method = http.MethodPut
	b.WithOptionFns(optFns)
	return b
}

// PutDo sets the method to PUT and sets the given pathURL, then send request and return response.
func (b *Builder) PutDo(pathURL string, data any, optFns ...OptionFn) (*Response, error) {
	return b.Put(pathURL, data, optFns...).Do()
}

// Patch sets the method to PATCH and sets the given pathURL
func (b *Builder) Patch(pathURL string, data any, optFns ...OptionFn) *Builder {
	b.Body = data
	b.pathURL = pathURL
	b.Method = http.MethodPatch
	b.WithOptionFns(optFns)
	return b
}

// PatchDo sets the method to PATCH and sets the given pathURL, then send request and return response.
func (b *Builder) PatchDo(pathURL string, data any, optFns ...OptionFn) (*Response, error) {
	return b.Patch(pathURL, data, optFns...).Do()
}

// Delete sets the method to DELETE and sets the given pathURL
func (b *Builder) Delete(pathURL string, optFns ...OptionFn) *Builder {
	b.pathURL = pathURL
	b.Method = http.MethodDelete
	b.WithOptionFns(optFns)
	return b
}

// DeleteDo sets the method to DELETE and sets the given pathURL, then send request and return response.
func (b *Builder) DeleteDo(pathURL string, optFns ...OptionFn) (*Response, error) {
	return b.Send(http.MethodDelete, pathURL, optFns...)
}

// Do send request and return response.
func (b *Builder) Do(optFns ...OptionFn) (*Response, error) {
	return b.Send(b.Method, b.pathURL, optFns...)
}

// Send request and return response, alias of Do()
func (b *Builder) Send(method, url string, optFns ...OptionFn) (*Response, error) {
	if len(optFns) > 0 {
		b.WithOptionFns(optFns)
	}

	b.Method = method
	cli := basefn.OrValue(b.cli == nil, std, b.cli)
	return cli.SendWithOpt(url, b.Options)
}
