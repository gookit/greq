// Package greq is a simple http client request builder, inspired from https://github.com/dghubble/sling
package greq

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/gookit/goutil/fsutil"
	"github.com/gookit/goutil/netutil/httpctype"
	"github.com/gookit/goutil/netutil/httpreq"
	"github.com/gookit/goutil/strutil"
)

// DefaultDoer for request.
var DefaultDoer = http.DefaultClient

// HandleFunc for the Middleware
type HandleFunc func(r *http.Request) (*Response, error)

// RequestCreator interface
type RequestCreator interface {
	NewRequest(method, target string, body io.Reader) *http.Request
}

// RequestCreatorFunc func
type RequestCreatorFunc func(method, target string, body io.Reader) *http.Request

// HiReq is an HTTP Request builder and sender.
type HiReq struct {
	client httpreq.Doer
	// core handler.
	handler HandleFunc
	middles []Middleware
	// http method eg: GET,POST
	method  string
	header  http.Header
	baseURL string
	// pathURL only valid in one request
	pathURL string
	// query params data. allow: map[string]string, url.Values
	queryParams url.Values
	// body provider
	bodyProvider BodyProvider
	respDecoder  RespDecoder
	// BeforeSend callback
	BeforeSend func(r *http.Request)
}

// New create
func New(baseURL ...string) *HiReq {
	h := &HiReq{
		client: http.DefaultClient,
		method: http.MethodGet,
		header: make(http.Header),
		// default use JSON decoder
		respDecoder: jsonDecoder{},
		queryParams: make(url.Values, 0),
	}

	if len(baseURL) > 0 {
		h.baseURL = baseURL[0]
	}
	return h
}

// New create an instance from current.
func (h *HiReq) New() *HiReq {
	// copy Headers pairs into new Header map
	headerCopy := make(http.Header)
	for k, v := range h.header {
		headerCopy[k] = v
	}

	return &HiReq{
		client:  h.client,
		method:  h.method,
		baseURL: h.baseURL,
		header:  headerCopy,
		// queryParams:    append([]any{}, s.queryParams...),
		bodyProvider: h.bodyProvider,
		respDecoder:  h.respDecoder,
	}
}

// ------------ Config ------------

// Doer custom set http request doer.
// If a nil client is given, the DefaultDoer will be used.
func (h *HiReq) Doer(doer httpreq.Doer) *HiReq {
	if doer != nil {
		h.client = doer
	} else {
		h.client = DefaultDoer
	}

	return h
}

// Client custom set http request doer
func (h *HiReq) Client(doer httpreq.Doer) *HiReq {
	return h.Doer(doer)
}

// HttpClient custom set http client as request doer
func (h *HiReq) HttpClient(hClient *http.Client) *HiReq {
	return h.Doer(hClient)
}

// Config custom config http request doer
func (h *HiReq) Config(fn func(doer httpreq.Doer)) *HiReq {
	fn(h.client)
	return h
}

// ConfigHClient custom config http client.
//
// Usage:
//
//	h.ConfigHClient(func(hClient *http.Client) {
//		hClient.Timeout = 30 * time.Second
//	})
func (h *HiReq) ConfigHClient(fn func(hClient *http.Client)) *HiReq {
	if hc, ok := h.client.(*http.Client); ok {
		fn(hc)
	} else {
		panic("the doer is not an *http.Client")
	}

	return h
}

// Use one or multi middlewares
func (h *HiReq) Use(middles ...Middleware) *HiReq {
	return h.Middlewares(middles...)
}

// Uses one or multi middlewares
func (h *HiReq) Uses(middles ...Middleware) *HiReq {
	return h.Middlewares(middles...)
}

// Middleware add one or multi middlewares
func (h *HiReq) Middleware(middles ...Middleware) *HiReq {
	return h.Middlewares(middles...)
}

// Middlewares add one or multi middlewares
func (h *HiReq) Middlewares(middles ...Middleware) *HiReq {
	h.middles = append(h.middles, middles...)
	return h
}

// WithRespDecoder for client
func (h *HiReq) WithRespDecoder(respDecoder RespDecoder) *HiReq {
	h.respDecoder = respDecoder
	return h
}

// OnBeforeSend for client
func (h *HiReq) OnBeforeSend(fn func(r *http.Request)) *HiReq {
	h.BeforeSend = fn
	return h
}

// ------------ Method ------------

// Method set http method name.
func (h *HiReq) Method(method string) *HiReq {
	h.method = method
	return h
}

// Head sets the method to HEAD and request the pathURL, then send request and return response.
func (h *HiReq) Head(pathURL string) *HiReq {
	return h.Method(http.MethodHead).PathURL(pathURL)
}

// HeadDo sets the method to HEAD and request the pathURL,
// then send request and return response.
func (h *HiReq) HeadDo(pathURL string) (*Response, error) {
	return h.Send(pathURL, http.MethodHead)
}

// Get sets the method to GET and sets the given pathURL
func (h *HiReq) Get(pathURL string) *HiReq {
	return h.Method(http.MethodGet).PathURL(pathURL)
}

// GetDo sets the method to GET and sets the given pathURL,
// then send request and return response.
func (h *HiReq) GetDo(pathURL string) (*Response, error) {
	return h.Send(pathURL, http.MethodGet)
}

// Post sets the method to POST and sets the given pathURL
func (h *HiReq) Post(pathURL string) *HiReq {
	return h.Method(http.MethodPost).PathURL(pathURL)
}

// PostDo sets the method to POST and sets the given pathURL,
// then send request and return http response.
func (h *HiReq) PostDo(pathURL string) (*Response, error) {
	return h.Send(pathURL, http.MethodPost)
}

// Put sets the method to PUT and sets the given pathURL
func (h *HiReq) Put(pathURL string) *HiReq {
	return h.Method(http.MethodPut).PathURL(pathURL)
}

// PutDo sets the method to PUT and sets the given pathURL,
// then send request and return http response.
func (h *HiReq) PutDo(pathURL string) (*Response, error) {
	return h.Send(pathURL, http.MethodPut)
}

// Patch sets the method to PATCH and sets the given pathURL
func (h *HiReq) Patch(pathURL string) *HiReq {
	return h.Method(http.MethodPatch).PathURL(pathURL)
}

// PatchDo sets the method to PATCH and sets the given pathURL,
// then send request and return http response.
func (h *HiReq) PatchDo(pathURL string) (*Response, error) {
	return h.Send(pathURL, http.MethodPatch)
}

// Delete sets the method to DELETE and sets the given pathURL
func (h *HiReq) Delete(pathURL string) *HiReq {
	return h.Method(http.MethodDelete).PathURL(pathURL)
}

// DeleteDo sets the method to DELETE and sets the given pathURL,
// then send request and return http response.
func (h *HiReq) DeleteDo(pathURL string) (*Response, error) {
	return h.Send(pathURL, http.MethodDelete)
}

// Trace sets the method to TRACE and sets the given pathURL,
// then send request and return http response.
func (h *HiReq) Trace(pathURL string) *HiReq {
	return h.Method(http.MethodTrace).PathURL(pathURL)
}

// TraceDo sets the method to TRACE and sets the given pathURL,
// then send request and return http response.
func (h *HiReq) TraceDo(pathURL string) (*Response, error) {
	return h.Send(pathURL, http.MethodTrace)
}

// Options sets the method to OPTIONS and request the pathURL,
// then send request and return response.
func (h *HiReq) Options(pathURL string) *HiReq {
	return h.Method(http.MethodOptions).PathURL(pathURL)
}

// OptionsDo sets the method to OPTIONS and request the pathURL,
// then send request and return response.
func (h *HiReq) OptionsDo(pathURL string) (*Response, error) {
	return h.Send(pathURL, http.MethodOptions)
}

// Connect sets the method to CONNECT and sets the given pathURL,
// then send request and return http response.
func (h *HiReq) Connect(pathURL string) (*Response, error) {
	return h.Send(pathURL, http.MethodConnect)
}

// ConnectDo sets the method to CONNECT and sets the given pathURL,
// then send request and return http response.
func (h *HiReq) ConnectDo(pathURL string) (*Response, error) {
	return h.Send(pathURL, http.MethodConnect)
}

// ----------- URL, query params ------------

// BaseURL set base URL for all request
func (h *HiReq) BaseURL(baseURL string) *HiReq {
	h.baseURL = baseURL
	return h
}

// PathURL set path URL for current request
func (h *HiReq) PathURL(pathURL string) *HiReq {
	h.pathURL = pathURL
	return h
}

// QueryParam appends new k-v param to the query string.
func (h *HiReq) QueryParam(key string, value any) *HiReq {
	h.queryParams.Add(key, strutil.MustString(value))
	return h
}

// QueryParams appends url.Values/map[string]string to the query string.
// The value will be encoded as url query parameters on send requests (see Send()).
func (h *HiReq) QueryParams(ps any) *HiReq {
	if ps != nil {
		queryValues := httpreq.ToQueryValues(ps)

		for key, values := range queryValues {
			for _, value := range values {
				h.queryParams.Add(key, value)
			}
		}
	}

	return h
}

// QueryValues appends url.Values to the query string.
// The value will be encoded as url query parameters on new requests (see Send()).
func (h *HiReq) QueryValues(values url.Values) *HiReq {
	return h.QueryParams(values)
}

// ----------- Header ------------

// AddHeader adds the key, value pair in Headers, appending values for existing keys
// to the key's values. Header keys are canonicalized.
func (h *HiReq) AddHeader(key, value string) *HiReq {
	h.header.Add(key, value)
	return h
}

// SetHeader sets the key, value pair in Headers, replacing existing values
// associated with key. Header keys are canonicalized.
func (h *HiReq) SetHeader(key, value string) *HiReq {
	h.header.Set(key, value)
	return h
}

// AddHeaders adds all the http.Header values, appending values for existing keys
// to the key's values. Header keys are canonicalized.
func (h *HiReq) AddHeaders(headers http.Header) *HiReq {
	for key, values := range headers {
		for i := range values {
			h.header.Add(key, values[i])
		}
	}
	return h
}

// SetHeaders sets all the http.Header values, replacing values for existing keys
// to the key's values. Header keys are canonicalized.
func (h *HiReq) SetHeaders(headers http.Header) *HiReq {
	for key, values := range headers {
		for i := range values {
			if i == 0 {
				h.header.Set(key, values[i])
			} else {
				h.header.Add(key, values[i])
			}
		}
	}
	return h
}

// ContentType with custom ContentType header
//
// Usage:
//
//	// json type
//	h.ContentType(httpctype.JSON)
//	// form type
//	h.ContentType(httpctype.Form)
func (h *HiReq) ContentType(value string) *HiReq {
	return h.SetHeader(httpctype.Key, value)
}

// XMLType with xml Content-Type header
func (h *HiReq) XMLType() *HiReq {
	return h.SetHeader(httpctype.Key, httpctype.XML)
}

// JSONType with json Content-Type header
func (h *HiReq) JSONType() *HiReq {
	return h.SetHeader(httpctype.Key, httpctype.JSON)
}

// FormType with from Content-Type header
func (h *HiReq) FormType() *HiReq {
	return h.SetHeader(httpctype.Key, httpctype.Form)
}

// MultipartType with multipart/form-data Content-Type header
func (h *HiReq) MultipartType() *HiReq {
	return h.SetHeader(httpctype.Key, httpctype.FormData)
}

// UserAgent with User-Agent header setting.
func (h *HiReq) UserAgent(value string) *HiReq {
	return h.SetHeader("User-Agent", value)
}

// UserAuth with user auth header value.
func (h *HiReq) UserAuth(value string) *HiReq {
	return h.SetHeader("Authorization", value)
}

// BasicAuth sets the Authorization header to use HTTP Basic Authentication
// with the provided username and password. With HTTP Basic Authentication
// the provided username and password are not encrypted.
func (h *HiReq) BasicAuth(username, password string) *HiReq {
	return h.SetHeader("Authorization", httpreq.BuildBasicAuth(username, password))
}

// SetCookies to request
func (h *HiReq) SetCookies(hcs ...*http.Cookie) *HiReq {
	var b strings.Builder
	for i, hc := range hcs {
		if i > 0 {
			b.WriteByte(';')
		}
		b.WriteString(hc.String())
	}

	return h.SetCookieString(b.String())
}

// SetCookieString set cookie header value.
//
// Usage:
//
//	h.New().
//		SetCookieString("name=inhere;age=30").
//		GetDo("/some/api")
func (h *HiReq) SetCookieString(value string) *HiReq {
	// return h.AddHeader("Set-Cookie", value)
	return h.AddHeader("Cookie", value)
}

// ----------- Request Body ------------

// Body with custom any type body
func (h *HiReq) Body(body any) *HiReq {
	return h.AnyBody(body)
}

// AnyBody with custom any type body
func (h *HiReq) AnyBody(body any) *HiReq {
	switch typVal := body.(type) {
	case io.Reader:
		h.BodyReader(typVal)
	case BodyProvider:
		h.BodyProvider(typVal)
	case string:
		h.StringBody(typVal)
	case []byte:
		h.BytesBody(typVal)
	default:
		panic("invalid data type as body")
	}
	return h
}

// BodyReader with custom io reader body
func (h *HiReq) BodyReader(r io.Reader) *HiReq {
	h.bodyProvider = bodyProvider{body: r}
	return h
}

// BodyProvider with custom body provider
func (h *HiReq) BodyProvider(bp BodyProvider) *HiReq {
	h.bodyProvider = bp
	return h
}

// FileContentsBody read file contents as body
func (h *HiReq) FileContentsBody(filePath string) *HiReq {
	file, err := os.OpenFile(filePath, os.O_RDONLY, fsutil.DefaultFilePerm)
	if err != nil {
		panic(err)
	}
	return h.BodyReader(file)
}

// JSONBody with JSON data body
func (h *HiReq) JSONBody(jsonData any) *HiReq {
	h.bodyProvider = jsonBodyProvider{
		payload: jsonData,
	}
	return h
}

// FormBody with form data body
func (h *HiReq) FormBody(formData any) *HiReq {
	h.bodyProvider = formBodyProvider{
		payload: formData,
	}
	return h
}

// BytesBody with custom string body
func (h *HiReq) BytesBody(bs []byte) *HiReq {
	return h.BodyReader(bytes.NewReader(bs))
}

// StringBody with custom string body
func (h *HiReq) StringBody(s string) *HiReq {
	return h.BodyReader(strings.NewReader(s))
}

// Multipart with custom multipart body
func (h *HiReq) Multipart(key, value string) *HiReq {
	// TODO
	return h
}

// ----------- Do send request ------------

// Do send request and return response
func (h *HiReq) Do(pathURLAndMethod ...string) (*Response, error) {
	return h.SendWithCtx(context.Background(), pathURLAndMethod...)
}

// Send request and return response
func (h *HiReq) Send(pathURLAndMethod ...string) (*Response, error) {
	return h.SendWithCtx(context.Background(), pathURLAndMethod...)
}

// MustSend send request and return response, will panic on error
func (h *HiReq) MustSend(pathURLAndMethod ...string) *Response {
	resp, err := h.SendWithCtx(context.Background(), pathURLAndMethod...)
	if err != nil {
		panic(err)
	}

	return resp
}

// SendRaw http request text.
func (h *HiReq) SendRaw(raw string) (*Response, error) {
	method := "GET"
	reqUrl := "TODO"

	var body io.Reader

	req, err := http.NewRequest(method, reqUrl, body)
	if err != nil {
		return nil, err
	}

	return h.SendRequest(req)
}

// ReqOption type
type ReqOption = httpreq.ReqOption

// SendWithOpt send request with option, then return response
func (h *HiReq) SendWithOpt(pathURL string, opt *ReqOption) (*Response, error) {
	// ensure option
	opt = ensureOpt(opt)

	// build request
	req, err := h.NewRequestWithCtx(opt.Context, pathURL, opt.Method)
	if err != nil {
		return nil, err
	}

	// set headers
	if len(opt.HeaderMap) > 0 {
		httpreq.AddHeaderMap(req, opt.HeaderMap)
	}
	if len(opt.ContentType) > 0 {
		req.Header.Set("Content-Type", opt.ContentType)
	}

	// do send
	return h.SendRequest(req)
}

// DoWithCtx request with context, then return response
func (h *HiReq) DoWithCtx(ctx context.Context, pathURLAndMethod ...string) (*Response, error) {
	return h.SendWithCtx(ctx, pathURLAndMethod...)
}

// SendWithCtx request with context, then return response
func (h *HiReq) SendWithCtx(ctx context.Context, pathURLAndMethod ...string) (*Response, error) {
	req, err := h.NewRequestWithCtx(ctx, pathURLAndMethod...)
	if err != nil {
		return nil, err
	}

	// do send
	return h.SendRequest(req)
}

// SendRequest send request
func (h *HiReq) SendRequest(req *http.Request) (*Response, error) {
	// call before send.
	if h.BeforeSend != nil {
		h.BeforeSend(req)
	}

	// wrap middlewares
	h.wrapMiddlewares()

	// do send
	return h.handler(req)
}

// ----------- Build request ------------

// Build new request
func (h *HiReq) Build(pathURLAndMethod ...string) (*http.Request, error) {
	return h.NewRequestWithCtx(context.Background(), pathURLAndMethod...)
}

// NewRequest build new request
func (h *HiReq) NewRequest(pathURLAndMethod ...string) (*http.Request, error) {
	return h.NewRequestWithCtx(context.Background(), pathURLAndMethod...)
}

// NewRequestWithCtx build new request with context
func (h *HiReq) NewRequestWithCtx(ctx context.Context, pathURLAndMethod ...string) (*http.Request, error) {
	method := h.method
	pathURL := h.pathURL
	if ln := len(pathURLAndMethod); ln > 0 {
		pathURL = pathURLAndMethod[0]
		if ln > 1 && len(pathURLAndMethod[1]) > 0 {
			method = pathURLAndMethod[1]
		}
	}

	fullURL := pathURL
	if len(h.baseURL) > 0 {
		if !strings.HasPrefix(pathURL, "http") {
			fullURL = h.baseURL + pathURL
		} else if len(pathURL) == 0 {
			fullURL = h.baseURL
		}
	}

	reqURL, err := url.Parse(fullURL)
	if err != nil {
		return nil, err
	}
	if reqURL.Scheme == "" {
		reqURL.Scheme = "https"
	}

	// append query params
	if err = httpreq.AppendQueryToURL(reqURL, h.queryParams); err != nil {
		return nil, err
	}

	var body io.Reader
	if h.bodyProvider != nil {
		body, err = h.bodyProvider.Body()
		if err != nil {
			return nil, err
		}

		if cType := h.bodyProvider.ContentType(); cType != "" {
			h.ContentType(cType)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, reqURL.String(), body)
	if err != nil {
		return nil, err
	}

	// copy headers
	httpreq.AddHeaders(req, h.header)

	// reset after request build.
	h.pathURL = ""
	return req, err
}

// String request to string.
func (h *HiReq) String() string {
	r, err := h.Build()
	if err != nil {
		return ""
	}
	return httpreq.RequestToString(r)
}
