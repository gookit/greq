package greq

import (
	"context"
	"fmt"
	"io"
	"net/http"
	// gourl "net/url"
	"strings"
	"time"

	"github.com/gookit/goutil/netutil/httpctype"
	"github.com/gookit/goutil/netutil/httpheader"
	"github.com/gookit/goutil/netutil/httpreq"
	"github.com/gookit/goutil/strutil"
	"github.com/gookit/greq/ext/httpfile"
)

// Client is an HTTP Request builder and sender.
type Client struct {
	doer httpreq.Doer
	// core handler.
	handler HandleFunc
	middles []Middleware

	//
	// default options for all requests
	//

	// Method default http method. default is GET
	Method string
	// Header default http header. default is nil
	Header http.Header
	// defalut content type
	ContentType string
	// BaseURL default base URL. default is ""
	BaseURL string
	// Timeout default timeout(ms) for each request. default 10s
	//
	//  - 0: not limit
	Timeout int
	// RespDecoder response data decoder.
	//  - use for create Response instance. default is JSON decoder
	RespDecoder RespDecoder

	// ReqVars template vars for request: URL, Header, Query, Body
	//
	// eg: http://example.com/${name}
	ReqVars map[string]string
	// BeforeSend callback on each request, can return error to deny request.
	BeforeSend func(r *http.Request) error
	// AfterSend callback on each request, can use for record request and response
	AfterSend AfterSendFn

	//
	// default retry config
	//

	// MaxRetries max retry times. default is 0 (not retry)
	MaxRetries int
	// RetryDelay retry delay time (ms). default is 0 (no delay)
	RetryDelay int
	// RetryChecker retry condition checker. default is nil (not retry)
	RetryChecker RetryChecker
}

// NewClient create a new http request client. alias of New()
func NewClient(baseURL ...string) *Client { return New(baseURL...) }

// New create a new http request client.
func New(baseURL ...string) *Client {
	timeoutMs := 10000 // default 10s
	timeout := 10 * time.Second
	h := &Client{
		doer: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				MaxIdleConns:        200,
				MaxIdleConnsPerHost: 50,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		Timeout: timeoutMs,
		Method: http.MethodGet,
		Header: make(http.Header),
		// default use JSON decoder
		RespDecoder: jsonDecoder{},
	}

	if len(baseURL) > 0 {
		h.BaseURL = baseURL[0]
	}
	return h
}

// Sub create an instance from current. will inherit all options
func (h *Client) Sub() *Client {
	// copy HeaderM pairs into new Header map
	headerCopy := make(http.Header)
	for k, v := range h.Header {
		headerCopy[k] = v
	}

	return &Client{
		doer:    h.doer,
		Method:  h.Method,
		BaseURL: h.BaseURL,
		Header:  headerCopy,
		// query:    append([]any{}, s.query...),
		RespDecoder: h.RespDecoder,
	}
}

//
// region Config
// ----------------------------

// WithRetryConfig set retry configuration
func (h *Client) WithRetryConfig(maxRetries, retryDelay int, checker RetryChecker) *Client {
	h.MaxRetries = maxRetries
	h.RetryDelay = retryDelay
	h.RetryChecker = checker
	return h
}

// WithMaxRetries set max retry times
func (h *Client) WithMaxRetries(maxRetries int) *Client {
	h.MaxRetries = maxRetries
	return h
}

// WithRetryDelay set retry delay time in milliseconds
func (h *Client) WithRetryDelay(retryDelay int) *Client {
	h.RetryDelay = retryDelay
	return h
}

// WithRetryChecker set custom retry checker function
func (h *Client) WithRetryChecker(checker RetryChecker) *Client {
	h.RetryChecker = checker
	return h
}

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
func (h *Client) Client(doer httpreq.Doer) *Client { return h.Doer(doer) }

// HttpClient custom set http cli as request doer
func (h *Client) HttpClient(hClient *http.Client) *Client { return h.Doer(hClient) }

// Config custom config for client.
//
// Usage:
//
//	h.Config(func(h *Client) {
//		h.Method = http.MethodPost
//	})
func (h *Client) Config(fn func(h *Client)) *Client {
	fn(h)
	return h
}

// ConfigDoer custom config http request doer
func (h *Client) ConfigDoer(fn func(doer httpreq.Doer)) *Client {
	fn(h.doer)
	return h
}

// ConfigHClient custom config http client.
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

// SetMaxIdleConns Set the maximum number of idle connections.
func (h *Client) SetMaxIdleConns(maxIdleConns, maxIdleConnsPerHost int) *Client {
	if hc, ok := h.doer.(*http.Client); ok {
		transport := hc.Transport.(*http.Transport)
		transport.MaxIdleConns = maxIdleConns
		transport.MaxIdleConnsPerHost = maxIdleConnsPerHost
	}
	return h
}

// DefaultTimeout set default timeout in milliseconds for requests. default is 0 (infinite)
func (h *Client) DefaultTimeout(timeoutMs int) *Client {
	h.Timeout = timeoutMs
	if hc, ok := h.doer.(*http.Client); ok {
		hc.Timeout = time.Duration(timeoutMs) * time.Millisecond
	}
	return h
}

// Use one or multi middlewares
func (h *Client) Use(middles ...Middleware) *Client { return h.Middlewares(middles...) }

// Uses one or multi middlewares
func (h *Client) Uses(middles ...Middleware) *Client { return h.Middlewares(middles...) }

// Middleware add one or multi middlewares
func (h *Client) Middleware(middles ...Middleware) *Client { return h.Middlewares(middles...) }

// Middlewares add one or multi middlewares
func (h *Client) Middlewares(middles ...Middleware) *Client {
	h.middles = append(h.middles, middles...)
	return h
}

// WithRespDecoder for cli
func (h *Client) WithRespDecoder(respDecoder RespDecoder) *Client {
	h.RespDecoder = respDecoder
	return h
}

// OnBeforeSend for cli
func (h *Client) OnBeforeSend(fn func(r *http.Request) error) *Client {
	h.BeforeSend = fn
	return h
}

// ----------- Header ------------

// DefaultHeader sets the http.Header value, it will be used for all requests.
func (h *Client) DefaultHeader(key, value string) *Client {
	h.Header.Set(key, value)
	return h
}

// DefaultHeaders sets all the http.Header values, it will be used for all requests.
func (h *Client) DefaultHeaders(headers http.Header) *Client {
	h.Header = headers
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

// WithBaseURL set default base URL for all request
func (h *Client) WithBaseURL(baseURL string) *Client {
	h.BaseURL = baseURL
	return h
}

// DefaultMethod set default method name. it will be used when the Get()/Post() method is empty.
func (h *Client) DefaultMethod(method string) *Client {
	h.Method = strings.ToUpper(method)
	return h
}

// Builder create a new builder with current client.
func (h *Client) Builder(optFns ...OptionFn) *Builder { return BuilderWithClient(h, optFns...) }

//
// region REST Methods
// ------------------------------
//

// Head sets the method to HEAD and request the pathURL, then send request and return response.
func (h *Client) Head(pathURL string) *Builder { return newBuilder(h, http.MethodHead, pathURL) }

// HeadDo sets the method to HEAD and request the pathURL,
// then send request and return response.
func (h *Client) HeadDo(pathURL string, optFns ...OptionFn) (*Response, error) {
	return h.SendWithOpt(pathURL, NewOpt2(optFns, http.MethodHead))
}

// Get sets the method to GET and sets the given pathURL
func (h *Client) Get(pathURL string) *Builder { return newBuilder(h, http.MethodGet, pathURL) }

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

// Download remote file from url and save to savePath.
func (h *Client) Download(url, savePath string, optFns ...OptionFn) (int, error) {
	resp, err := h.Send(http.MethodGet, url, optFns...)
	if err != nil {
		return 0, err
	}

	if resp.IsFail() {
		resp.QuietCloseBody()
		return 0, fmt.Errorf("download failed, status code: %d", resp.StatusCode)
	}
	return resp.SaveFile(savePath)
}

//
// region Headers
// ----------------------------

// WithContentType with custom Content-Type header
func (h *Client) WithContentType(value string) *Builder {
	return BuilderWithClient(h).WithContentType(value)
}

// JSONType with json Content-Type header
func (h *Client) JSONType() *Builder { return BuilderWithClient(h).JSONType() }

// FormType with from Content-Type header
func (h *Client) FormType() *Builder { return BuilderWithClient(h).FormType() }

// UserAgent with User-Agent header setting for all requests.
func (h *Client) UserAgent(value string) *Builder { return BuilderWithClient(h).UserAgent(value) }

// UserAuth with user auth header value for all requests.
func (h *Client) UserAuth(value string) *Builder { return BuilderWithClient(h).UserAuth(value) }

// BasicAuth sets the Authorization header to use HTTP Basic Authentication
// with the provided username and password.
//
// With HTTP Basic Authentication the provided username and password are not encrypted.
func (h *Client) BasicAuth(username, password string) *Builder {
	return BuilderWithClient(h).BasicAuth(username, password)
}

//
// region URL and Query
// ---------------------------

// QueryParams appends url.Values/map[string]string to the Query string.
// The value will be encoded as url Query parameters on send requests (see Send()).
func (h *Client) QueryParams(ps any) *Builder {
	return BuilderWithClient(h).QueryParams(ps)
}

//
// region Set Body
// -----------------------------

// Body with custom any type body
func (h *Client) Body(body any) *Builder { return BuilderWithClient(h).AnyBody(body) }

// AnyBody with custom any type body
func (h *Client) AnyBody(body any) *Builder { return BuilderWithClient(h).AnyBody(body) }

// BodyReader with custom io reader body
func (h *Client) BodyReader(r io.Reader) *Builder { return BuilderWithClient(h).BodyReader(r) }

// BodyProvider with custom body provider
func (h *Client) BodyProvider(bp BodyProvider) *Builder { return BuilderWithClient(h).BodyProvider(bp) }

//
// region Send request
// ------------------------------

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

// SendRaw http request text. like IDE .http file contents
//
// Format:
//
//	POST https://example.com/path?name=inhere
//	Content-Type: application/json
//	Accept: */*
//
//	<content>
func (h *Client) SendRaw(raw string, varMp map[string]string) (*Response, error) {
	rawReq, err := httpfile.ParseRequest(raw)
	if err != nil {
		return nil, err
	}
	rawReq.ApplyVars(varMp)

	var body = strings.NewReader(rawReq.Body)
	fullURL := h.buildFullURL(rawReq.URL)

	req, err := http.NewRequest(rawReq.Method, fullURL, body)
	if err != nil {
		return nil, err
	}

	// apply headers
	if len(rawReq.Headers) > 0 {
		httpreq.SetHeaderMap(req, rawReq.Headers)
	}
	// apply default content type
	if len(h.ContentType) > 0 && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", h.ContentType)
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

	// do send with request-level retry config if set
	if opt.MaxRetries > 0 || opt.RetryDelay > 0 || opt.RetryChecker != nil {
		return h.sendRequestWithRetryConfig(req, opt)
	}

	// do send with client-level retry config
	return h.SendRequest(req)
}

// sendRequestWithRetryConfig send request with request-level retry configuration
func (h *Client) sendRequestWithRetryConfig(req *http.Request, opt *Options) (*Response, error) {
	// save original client retry config
	originalMaxRetries := h.MaxRetries
	originalRetryDelay := h.RetryDelay
	originalRetryChecker := h.RetryChecker

	// apply request-level retry config
	if opt.MaxRetries > 0 {
		h.MaxRetries = opt.MaxRetries
	}
	if opt.RetryDelay > 0 {
		h.RetryDelay = opt.RetryDelay
	}
	if opt.RetryChecker != nil {
		h.RetryChecker = opt.RetryChecker
	}

	// send request with retry
	resp, err := h.SendRequest(req)

	// restore original client retry config
	h.MaxRetries = originalMaxRetries
	h.RetryDelay = originalRetryDelay
	h.RetryChecker = originalRetryChecker

	return resp, err
}

// SendRequest send request
func (h *Client) SendRequest(req *http.Request) (*Response, error) {
	return h.sendRequestWithRetry(req, 0)
}

// sendRequestWithRetry send request with retry logic
func (h *Client) sendRequestWithRetry(req *http.Request, attempt int) (*Response, error) {
	start := time.Now()

	// wrap middlewares, and will wrap http.Response to Response
	h.wrapMiddlewares()

	// call before send.
	if h.BeforeSend != nil {
		if err := h.BeforeSend(req); err != nil {
			return nil, fmt.Errorf("before send check failed: %w", err)
		}
	}

	// do send by core handler
	resp, err := h.handler(req)
	if resp != nil {
		// set cost time
		resp.CostTime = time.Since(start).Milliseconds()
	}

	// call after send.
	if h.AfterSend != nil {
		h.AfterSend(resp, err)
	}

	// check if retry is needed
	if h.shouldRetry(resp, err, attempt) {
		return h.retryRequest(req, attempt)
	}

	return resp, err
}

// shouldRetry check if request should be retried
func (h *Client) shouldRetry(resp *Response, err error, attempt int) bool {
	// check max retries
	if h.MaxRetries <= 0 || attempt >= h.MaxRetries {
		return false
	}

	// use custom retry checker if set, otherwise use default
	checker := h.RetryChecker
	if checker == nil {
		checker = DefaultRetryChecker
	}

	return checker(resp, err, attempt)
}

// retryRequest perform retry with delay
func (h *Client) retryRequest(req *http.Request, attempt int) (*Response, error) {
	// apply retry delay if set
	if h.RetryDelay > 0 {
		time.Sleep(time.Duration(h.RetryDelay) * time.Millisecond)
	}

	// increment attempt count and retry
	return h.sendRequestWithRetry(req, attempt+1)
}

//
// region Build request
// -----------------------------------

// NewRequest build new request
func (h *Client) NewRequest(method, url string, optFns ...OptionFn) (*http.Request, error) {
	return h.NewRequestWithOptions(url, NewOpt2(optFns, method))
}

func (h *Client) buildFullURL(url string) string {
	fullURL := url
	if len(h.BaseURL) > 0 {
		if !strings.HasPrefix(url, "http") {
			fullURL = h.BaseURL + url
		} else if len(url) == 0 {
			fullURL = h.BaseURL
		}
	}
	return fullURL
}

// NewRequestWithOptions build new request with Options
func (h *Client) NewRequestWithOptions(url string, opt *Options) (*http.Request, error) {
	fullURL := h.buildFullURL(url)

	opt = orCreate(opt)
	ctx := opt.Context
	if ctx == nil {
		ctx = context.Background()
	}

	// append Query params
	qm := opt.Query
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

	cType := strutil.Valid(opt.ContentType, opt.HeaderM[httpctype.Key], h.ContentType)
	method := strings.ToUpper(strutil.OrElse(opt.Method, h.Method))
	allowBody := httpreq.IsNoBodyMethod(method) == false

	// check opt.Data
	if opt.Data != nil {
		if !allowBody {
			body = nil
			fullURL = httpreq.AppendQueryToURLString(fullURL, httpreq.MakeQuery(opt.Data))
		} else if body == nil {
			body = httpreq.MakeBody(opt.Data, cType)
		}
	}

	// check opt.Body
	if allowBody && body == nil && opt.Body != nil {
		body = httpreq.MakeBody(opt.Body, cType)
	}

	// create request
	req, err := http.NewRequestWithContext(ctx, method, fullURL, body)
	if err != nil {
		return nil, err
	}

	// copy and set headers
	httpreq.SetHeaders(req, h.Header, opt.Header)
	if len(opt.HeaderM) > 0 {
		httpreq.SetHeaderMap(req, opt.HeaderM)
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
