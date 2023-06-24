package greq

import (
	"context"
	"io"
	"net/http"
	gourl "net/url"
	"strings"

	"github.com/gookit/goutil/basefn"
	"github.com/gookit/goutil/netutil/httpctype"
	"github.com/gookit/goutil/netutil/httpheader"
	"github.com/gookit/goutil/netutil/httpreq"
	"github.com/gookit/goutil/strutil"
)

// Client is an HTTP Request builder and sender.
type Client struct {
	doer httpreq.Doer
	// core handler.
	handler HandleFunc
	middles []Middleware
	// Vars template vars for URL, Header, Query, Body
	//
	// eg: http://example.com/{name}
	Vars map[string]string

	// BeforeSend callback
	BeforeSend func(r *http.Request)
	// AfterSend callback
	AfterSend AfterSendFn

	//
	// default options for all requests
	//

	// http method eg: GET,POST
	method  string
	baseURL string
	header  http.Header
	// Query params data. allow: map[string]string, url.Values
	query gourl.Values
	// content type
	ContentType string
	// response data decoder
	respDecoder RespDecoder
}

// New create
func New(baseURL ...string) *Client {
	h := &Client{
		doer:   &http.Client{},
		method: http.MethodGet,
		header: make(http.Header),
		query:  make(gourl.Values),
		// default use JSON decoder
		respDecoder: jsonDecoder{},
	}

	if len(baseURL) > 0 {
		h.baseURL = baseURL[0]
	}
	return h
}

// Sub create an instance from current.
func (h *Client) Sub() *Client {
	// copy HeaderM pairs into new Header map
	headerCopy := make(http.Header)
	for k, v := range h.header {
		headerCopy[k] = v
	}

	return &Client{
		doer:    h.doer,
		method:  h.method,
		baseURL: h.baseURL,
		header:  headerCopy,
		// query:    append([]any{}, s.query...),
		respDecoder: h.respDecoder,
	}
}

// ------------ Config ------------

// Doer custom set http request doer.
// If a nil cli is given, the DefaultDoer will be used.
func (h *Client) Doer(doer httpreq.Doer) *Client {
	if doer != nil {
		h.doer = doer
	} else {
		h.doer = DefaultDoer
	}
	return h
}

// Client custom set http request doer
func (h *Client) Client(doer httpreq.Doer) *Client {
	return h.Doer(doer)
}

// HttpClient custom set http cli as request doer
func (h *Client) HttpClient(hClient *http.Client) *Client {
	return h.Doer(hClient)
}

// Config custom config http request doer
func (h *Client) Config(fn func(doer httpreq.Doer)) *Client {
	fn(h.doer)
	return h
}

// ConfigHClient custom config http cli.
//
// Usage:
//
//	h.ConfigHClient(func(hClient *http.Client) {
//		hClient.Timeout = 30 * time.Second
//	})
func (h *Client) ConfigHClient(fn func(hClient *http.Client)) *Client {
	if hc, ok := h.doer.(*http.Client); ok {
		fn(hc)
	} else {
		panic("the doer is not an *http.Client")
	}

	return h
}

// Use one or multi middlewares
func (h *Client) Use(middles ...Middleware) *Client {
	return h.Middlewares(middles...)
}

// Uses one or multi middlewares
func (h *Client) Uses(middles ...Middleware) *Client {
	return h.Middlewares(middles...)
}

// Middleware add one or multi middlewares
func (h *Client) Middleware(middles ...Middleware) *Client {
	return h.Middlewares(middles...)
}

// Middlewares add one or multi middlewares
func (h *Client) Middlewares(middles ...Middleware) *Client {
	h.middles = append(h.middles, middles...)
	return h
}

// WithRespDecoder for cli
func (h *Client) WithRespDecoder(respDecoder RespDecoder) *Client {
	h.respDecoder = respDecoder
	return h
}

// OnBeforeSend for cli
func (h *Client) OnBeforeSend(fn func(r *http.Request)) *Client {
	h.BeforeSend = fn
	return h
}

// ----------- Header ------------

// DefaultHeader sets the http.Header value, it will be used for all requests.
func (h *Client) DefaultHeader(key, value string) *Client {
	h.header.Set(key, value)
	return h
}

// DefaultHeaders sets all the http.Header values, it will be used for all requests.
func (h *Client) DefaultHeaders(headers http.Header) *Client {
	h.header = headers
	return h
}

// DefaultContentType set default ContentType header, it will be used for all requests.
//
// Usage:
//
//	// json type
//	h.DefaultContentType(httpctype.JSON)
//	// form type
//	h.DefaultContentType(httpctype.Form)
func (h *Client) DefaultContentType(value string) *Client {
	h.ContentType = value
	return h
}

// DefaultUserAgent with User-Agent header setting for all requests.
func (h *Client) DefaultUserAgent(value string) *Client {
	return h.DefaultHeader(httpheader.UserAgent, value)
}

// DefaultUserAuth with user auth header value for all requests.
func (h *Client) DefaultUserAuth(value string) *Client {
	return h.DefaultHeader(httpheader.UserAuth, value)
}

// DefaultBasicAuth sets the Authorization header to use HTTP Basic Authentication
// with the provided username and password. With HTTP Basic Authentication
// the provided username and password are not encrypted.
func (h *Client) DefaultBasicAuth(username, password string) *Client {
	return h.DefaultHeader(httpheader.UserAuth, httpreq.BuildBasicAuth(username, password))
}

// ------------ Method ------------

// BaseURL set default base URL for all request
func (h *Client) BaseURL(baseURL string) *Client {
	h.baseURL = baseURL
	return h
}

// DefaultMethod set default method name. it will be used when the Get()/Post() method is empty.
func (h *Client) DefaultMethod(method string) *Client {
	h.method = method
	return h
}

//
// ------------ REST requests ------------
//

// Head sets the method to HEAD and request the pathURL, then send request and return response.
func (h *Client) Head(pathURL string) *Builder {
	return newBuilder(h, http.MethodHead, pathURL)
}

// HeadDo sets the method to HEAD and request the pathURL,
// then send request and return response.
func (h *Client) HeadDo(pathURL string, withOpt ...*Options) (*Response, error) {
	opt := basefn.FirstOr(withOpt, &Options{})
	return h.SendWithOpt(pathURL, opt.WithMethod(http.MethodHead))
}

// Get sets the method to GET and sets the given pathURL
func (h *Client) Get(pathURL string) *Builder {
	return newBuilder(h, http.MethodGet, pathURL)
}

// GetDo sets the method to GET and sets the given pathURL,
// then send request and return response.
func (h *Client) GetDo(pathURL string, optFns ...OptionFn) (*Response, error) {
	return h.SendWithOpt(pathURL, NewOpt2(optFns, http.MethodGet))
}

// Post sets the method to POST and sets the given pathURL
func (h *Client) Post(pathURL string) *Builder {
	return newBuilder(h, http.MethodPost, pathURL)
}

// PostDo sets the method to POST and sets the given pathURL,
// then send request and return http response.
func (h *Client) PostDo(pathURL string, optFns ...OptionFn) (*Response, error) {
	return h.SendWithOpt(pathURL, NewOpt2(optFns, http.MethodPost))
}

// Put sets the method to PUT and sets the given pathURL
func (h *Client) Put(pathURL string) *Builder {
	return newBuilder(h, http.MethodPut, pathURL)
}

// PutDo sets the method to PUT and sets the given pathURL,
// then send request and return http response.
func (h *Client) PutDo(pathURL string, optFns ...OptionFn) (*Response, error) {
	return h.SendWithOpt(pathURL, NewOpt2(optFns, http.MethodPut))
}

// Patch sets the method to PATCH and sets the given pathURL
func (h *Client) Patch(pathURL string) *Builder {
	return newBuilder(h, http.MethodPatch, pathURL)
}

// PatchDo sets the method to PATCH and sets the given pathURL,
// then send request and return http response.
func (h *Client) PatchDo(pathURL string, optFns ...OptionFn) (*Response, error) {
	return h.SendWithOpt(pathURL, NewOpt2(optFns, http.MethodPatch))
}

// Delete sets the method to DELETE and sets the given pathURL
func (h *Client) Delete(pathURL string) *Builder {
	return newBuilder(h, http.MethodDelete, pathURL)
}

// DeleteDo sets the method to DELETE and sets the given pathURL,
// then send request and return http response.
func (h *Client) DeleteDo(pathURL string, optFns ...OptionFn) (*Response, error) {
	return h.SendWithOpt(pathURL, NewOpt2(optFns, http.MethodDelete))
}

// ----------- URL, Query params ------------

// JSONType with json Content-Type header
func (h *Client) JSONType() *Builder {
	return BuilderWithClient(h).JSONType()
}

// FormType with from Content-Type header
func (h *Client) FormType() *Builder {
	return BuilderWithClient(h).FormType()
}

// UserAgent with User-Agent header setting for all requests.
func (h *Client) UserAgent(value string) *Builder {
	return BuilderWithClient(h).UserAgent(value)
}

// UserAuth with user auth header value for all requests.
func (h *Client) UserAuth(value string) *Builder {
	return BuilderWithClient(h).UserAuth(value)
}

// BasicAuth sets the Authorization header to use HTTP Basic Authentication
// with the provided username and password.
//
// With HTTP Basic Authentication the provided username and password are not encrypted.
func (h *Client) BasicAuth(username, password string) *Builder {
	return BuilderWithClient(h).BasicAuth(username, password)
}

//
//
// ----------- URL, Query params ------------
//
//

// QueryParams appends url.Values/map[string]string to the Query string.
// The value will be encoded as url Query parameters on send requests (see Send()).
func (h *Client) QueryParams(ps any) *Builder {
	return BuilderWithClient(h).QueryParams(ps)
}

// ----------- Request Body ------------

// Body with custom any type body
func (h *Client) Body(body any) *Builder {
	return BuilderWithClient(h).AnyBody(body)
}

// AnyBody with custom any type body
func (h *Client) AnyBody(body any) *Builder {
	return BuilderWithClient(h).AnyBody(body)
}

// BodyReader with custom io reader body
func (h *Client) BodyReader(r io.Reader) *Builder {
	return BuilderWithClient(h).BodyReader(r)
}

// BodyProvider with custom body provider
func (h *Client) BodyProvider(bp BodyProvider) *Builder {
	b := BuilderWithClient(h)
	b.Provider = bp
	return b
}

// ----------- Do send request ------------

// Do send request and return response
func (h *Client) Do(method, url string, optFns ...OptionFn) (*Response, error) {
	return h.SendWithOption(method, url, optFns...)
}

// Send request and return response, alias of Do()
func (h *Client) Send(method, url string, optFns ...OptionFn) (*Response, error) {
	return h.SendWithOption(method, url, optFns...)
}

// MustSend send request and return response, will panic on error
func (h *Client) MustSend(method, url string, optFns ...OptionFn) *Response {
	resp, err := h.SendWithOption(method, url, optFns...)
	if err != nil {
		panic(err)
	}
	return resp
}

// SendRaw http request text.
//
// Format:
//
//	POST https://example.com/path?name=inhere
//	Content-Type: application/json
//	Accept: */*
//
//	<content>
func (h *Client) SendRaw(raw string, varMp map[string]string) (*Response, error) {
	method := "GET"
	reqUrl := "TODO"

	var body io.Reader
	req, err := http.NewRequest(method, reqUrl, body)
	if err != nil {
		return nil, err
	}

	return h.SendRequest(req)
}

// DoWithOption request with options, then return response
func (h *Client) DoWithOption(method, url string, optFns ...OptionFn) (*Response, error) {
	return h.SendWithOption(method, url, optFns...)
}

// SendWithOption request with options, then return response
func (h *Client) SendWithOption(method, url string, optFns ...OptionFn) (*Response, error) {
	return h.SendWithOpt(url, NewOpt2(optFns, method))
}

// SendWithOpt send request with option, then return response
func (h *Client) SendWithOpt(pathURL string, opt *Options) (*Response, error) {
	// build request
	req, err := h.NewRequestWithOptions(pathURL, opt)
	if err != nil {
		return nil, err
	}

	// do send
	return h.SendRequest(req)
}

// SendRequest send request
func (h *Client) SendRequest(req *http.Request) (*Response, error) {
	// wrap middlewares
	h.wrapMiddlewares()

	// call before send.
	if h.BeforeSend != nil {
		h.BeforeSend(req)
	}

	// do send by core handler
	resp, err := h.handler(req)

	// call after send.
	if h.AfterSend != nil {
		h.AfterSend(resp, err)
	}
	return resp, err
}

// ----------- Build request ------------

// NewRequest build new request
func (h *Client) NewRequest(method, url string, optFns ...OptionFn) (*http.Request, error) {
	return h.NewRequestWithOptions(url, NewOpt2(optFns, method))
}

// NewRequestWithOptions build new request with Options
func (h *Client) NewRequestWithOptions(url string, opt *Options) (*http.Request, error) {
	fullURL := url

	if len(h.baseURL) > 0 {
		if !strings.HasPrefix(url, "http") {
			fullURL = h.baseURL + url
		} else if len(url) == 0 {
			fullURL = h.baseURL
		}
	}

	opt = orCreate(opt)
	ctx := opt.Context
	if ctx == nil {
		ctx = context.Background()
	}

	// append Query params
	qm := MergeURLValues(h.query, opt.Query)
	if len(qm) > 0 {
		fullURL = httpreq.AppendQueryToURLString(fullURL, qm)
	}

	// make body
	var err error
	var body io.Reader
	if opt.Provider != nil {
		body, err = opt.Provider.Body()
		if err != nil {
			return nil, err
		}

		if bpTyp := opt.Provider.ContentType(); bpTyp != "" {
			opt.ContentType = bpTyp
		}
	}

	cType := strutil.OrElse(opt.ContentType, h.ContentType)
	method := strings.ToUpper(strutil.OrElse(opt.Method, h.method))

	if opt.Data != nil {
		if httpreq.IsNoBodyMethod(method) {
			body = nil
			fullURL = httpreq.AppendQueryToURLString(fullURL, httpreq.MakeQuery(opt.Data))
		} else if body == nil {
			cType := strutil.OrElse(opt.HeaderM[httpctype.Key], cType)
			body = httpreq.MakeBody(opt.Data, cType)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, body)
	if err != nil {
		return nil, err
	}

	// copy and set headers
	SetHeaders(req, h.header, opt.Header)
	if len(opt.HeaderM) > 0 {
		SetHeaderMap(req, opt.HeaderM)
	}
	if len(cType) > 0 {
		req.Header.Set(httpheader.ContentType, cType)
	}

	return req, err
}

// String request to string.
func (h *Client) String() string {
	r, err := h.NewRequest("", "")
	if err != nil {
		return ""
	}
	return httpreq.RequestToString(r)
}
