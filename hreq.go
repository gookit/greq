// Package hreq is a simple http client request builder, inspired from https://github.com/dghubble/sling
package hreq

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
)

// HandleFunc implements the Middleware interface
type HandleFunc func(r *http.Request) (*Response, error)

// RequestCreator interface
type RequestCreator interface {
	NewRequest(method, target string, body io.Reader) *http.Request
}

// RequestCreatorFunc func
type RequestCreatorFunc func(method, target string, body io.Reader) *http.Request

// HReq is an HTTP Request builder and sender.
type HReq struct {
	client httpreq.Doer
	// core handler.
	handler HandleFunc
	middles []Middleware
	// http method eg: GET,POST
	method  string
	header  http.Header
	baseURL string
	// query params data. allow: map[string]string, url.Values
	queryParams url.Values
	// body provider
	bodyProvider BodyProvider
	respDecoder  RespDecoder
	// beforeSend callback
	beforeSend func(req *http.Request)
}

// New create
func New(baseURL ...string) *HReq {
	h := &HReq{
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
func (h *HReq) New() *HReq {
	// copy Headers pairs into new Header map
	headerCopy := make(http.Header)
	for k, v := range h.header {
		headerCopy[k] = v
	}

	return &HReq{
		client:  h.client,
		method:  h.method,
		baseURL: h.baseURL,
		header:  headerCopy,
		// queryParams:    append([]interface{}{}, s.queryParams...),
		bodyProvider: h.bodyProvider,
		respDecoder:  h.respDecoder,
	}
}

// ------------ Config ------------

// Doer custom set http request doer.
// If a nil client is given, the http.DefaultClient will be used.
func (h *HReq) Doer(doer httpreq.Doer) *HReq {
	if doer != nil {
		h.client = doer
	} else {
		h.client = http.DefaultClient
	}

	return h
}

// Client custom set http request doer
func (h *HReq) Client(doer httpreq.Doer) *HReq {
	return h.Doer(doer)
}

// HttpClient custom set http client as request doer
func (h *HReq) HttpClient(hClient *http.Client) *HReq {
	return h.Doer(hClient)
}

// Config custom config http request doer
func (h *HReq) Config(fn func(doer httpreq.Doer)) *HReq {
	fn(h.client)
	return h
}

// ConfigHClient custom config http client.
//
// Usage:
// 	h.ConfigHClient(func(hClient *http.Client) {
//		hClient.Timeout = 30 * time.Second
// 	})
func (h *HReq) ConfigHClient(fn func(hClient *http.Client)) *HReq {
	if hc, ok := h.client.(*http.Client); ok {
		fn(hc)
	} else {
		panic("the doer is not an *http.Client")
	}

	return h
}

// Use one or multi middlewares
func (h *HReq) Use(middles ...Middleware) *HReq {
	return h.Middlewares(middles...)
}

// Uses one or multi middlewares
func (h *HReq) Uses(middles ...Middleware) *HReq {
	return h.Middlewares(middles...)
}

// Middleware add one or multi middlewares
func (h *HReq) Middleware(middles ...Middleware) *HReq {
	return h.Middlewares(middles...)
}

// Middlewares add one or multi middlewares
func (h *HReq) Middlewares(middles ...Middleware) *HReq {
	h.middles = append(h.middles, middles...)
	return h
}

// ------------ Method ------------

// Method set http method name.
func (h *HReq) Method(method string) *HReq {
	h.method = method
	return h
}

// Head sets the method to HEAD and request the pathURL, then send request and return response.
func (h *HReq) Head(pathURL string) (*Response, error) {
	return h.Send(pathURL, http.MethodHead)
}

// Get sets the method to GET and sets the given pathURL, then send request and return response.
func (h *HReq) Get(pathURL string) (*Response, error) {
	return h.Send(pathURL, http.MethodGet)
}

// Post sets the method to POST and sets the given pathURL,
// then send request and return http response.
func (h *HReq) Post(pathURL string) (*Response, error) {
	return h.Send(pathURL, http.MethodPost)
}

// Put sets the method to PUT and sets the given pathURL,
// then send request and return http response.
func (h *HReq) Put(pathURL string) (*Response, error) {
	return h.Send(pathURL, http.MethodPut)
}

// Patch sets the method to PATCH and sets the given pathURL,
// then send request and return http response.
func (h *HReq) Patch(pathURL string) (*Response, error) {
	return h.Send(pathURL, http.MethodPatch)
}

// Delete sets the method to DELETE and sets the given pathURL,
// then send request and return http response.
func (h *HReq) Delete(pathURL string) (*Response, error) {
	return h.Send(pathURL, http.MethodDelete)
}

// Trace sets the method to TRACE and sets the given pathURL,
// then send request and return http response.
func (h *HReq) Trace(pathURL string) (*Response, error) {
	return h.Send(pathURL, http.MethodTrace)
}

// Options sets the method to OPTIONS and request the pathURL,
// then send request and return response.
func (h *HReq) Options(pathURL string) (*Response, error) {
	return h.Send(pathURL, http.MethodOptions)
}

// Connect sets the method to CONNECT and sets the given pathURL,
// then send request and return http response.
func (h *HReq) Connect(pathURL string) (*Response, error) {
	return h.Send(pathURL, http.MethodConnect)
}

// ----------- URL, query params ------------

// BaseURL set base URL for request
func (h *HReq) BaseURL(baseURL string) *HReq {
	h.baseURL = baseURL
	return h
}

// QueryParam appends new k-v param to the query string.
func (h *HReq) QueryParam(key, value string) *HReq {
	h.queryParams.Add(key, value)

	return h
}

// QueryParams appends url.Values/map[string]string to the query string.
// The value will be encoded as url query parameters on send requests (see Send()).
func (h *HReq) QueryParams(ps interface{}) *HReq {
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
func (h *HReq) QueryValues(values url.Values) *HReq {
	return h.QueryParams(values)
}

// ----------- Header ------------

// AddHeader adds the key, value pair in Headers, appending values for existing keys
// to the key's values. Header keys are canonicalized.
func (h *HReq) AddHeader(key, value string) *HReq {
	h.header.Add(key, value)
	return h
}

// SetHeader sets the key, value pair in Headers, replacing existing values
// associated with key. Header keys are canonicalized.
func (h *HReq) SetHeader(key, value string) *HReq {
	h.header.Set(key, value)
	return h
}

// AddHeaders adds all the http.Header values, appending values for existing keys
// to the key's values. Header keys are canonicalized.
func (h *HReq) AddHeaders(headers http.Header) *HReq {
	for key, values := range headers {
		for i := range values {
			h.header.Add(key, values[i])
		}
	}
	return h
}

// SetHeaders sets all the http.Header values, replacing values for existing keys
// to the key's values. Header keys are canonicalized.
func (h *HReq) SetHeaders(headers http.Header) *HReq {
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
//	// json type
//	h.ContentType(httpctype.JSON)
//	// form type
//	h.ContentType(httpctype.Form)
func (h *HReq) ContentType(value string) *HReq {
	return h.SetHeader(httpctype.Key, value)
}

// JSONType with json Content-Type header
func (h *HReq) JSONType() *HReq {
	return h.SetHeader(httpctype.Key, httpctype.JSON)
}

// FormType with from Content-Type header
func (h *HReq) FormType() *HReq {
	return h.SetHeader(httpctype.Key, httpctype.Form)
}

// MultipartType with multipart/form-data Content-Type header
func (h *HReq) MultipartType() *HReq {
	return h.SetHeader(httpctype.Key, httpctype.FormData)
}

// UserAgent with User-Agent header setting.
func (h *HReq) UserAgent(value string) *HReq {
	return h.SetHeader("User-Agent", value)
}

// UserAuth with user auth header value.
func (h *HReq) UserAuth(value string) *HReq {
	return h.SetHeader("Authorization", value)
}

// BasicAuth sets the Authorization header to use HTTP Basic Authentication
// with the provided username and password. With HTTP Basic Authentication
// the provided username and password are not encrypted.
func (h *HReq) BasicAuth(username, password string) *HReq {
	return h.SetHeader("Authorization", httpreq.BuildBasicAuth(username, password))
}

// SetCookieString set cookie header value.
//
// Usage:
//	h.New().
//		SetCookieString("name=inhere;age=30").
//		Get("/some/api")
func (h *HReq) SetCookieString(value string) *HReq {
	// return h.AddHeader("Set-Cookie", value)
	return h.AddHeader("Cookie", value)
}

// ----------- Request Body ------------

// Body with custom body
func (h *HReq) Body(body interface{}) *HReq {
	switch typVal := body.(type) {
	case io.Reader:
		h.BodyReader(typVal)
		break
	case BodyProvider:
		h.BodyProvider(typVal)
		break
	case string:
		h.StringBody(typVal)
		break
	case []byte:
		h.BytesBody(typVal)
		break
	default:
		panic("invalid data type as body")
	}
	return h
}

// BodyReader with custom io reader body
func (h *HReq) BodyReader(r io.Reader) *HReq {
	h.bodyProvider = bodyProvider{body: r}
	return h
}

// BodyProvider with custom body provider
func (h *HReq) BodyProvider(bp BodyProvider) *HReq {
	h.bodyProvider = bp
	return h
}

// FileContentsBody read file contents as body
func (h *HReq) FileContentsBody(filepath string) *HReq {
	file, err := fsutil.OpenFile(filepath, os.O_RDONLY, fsutil.DefaultFilePerm)
	if err != nil {
		panic(err)
	}

	return h.BodyReader(file)
}

// JSONBody with JSON data body
func (h *HReq) JSONBody(jsonData interface{}) *HReq {
	h.bodyProvider = jsonBodyProvider{
		payload: jsonData,
	}
	return h
}

// FormBody with form data body
func (h *HReq) FormBody(formData interface{}) *HReq {
	h.bodyProvider = formBodyProvider{
		payload: formData,
	}
	return h
}

// BytesBody with custom string body
func (h *HReq) BytesBody(bs []byte) *HReq {
	return h.BodyReader(bytes.NewReader(bs))
}

// StringBody with custom string body
func (h *HReq) StringBody(s string) *HReq {
	return h.BodyReader(strings.NewReader(s))
}

// Multipart with custom multipart body
func (h *HReq) Multipart(key, value string) *HReq {
	// TODO
	return h
}

// ----------- Do send request ------------

// Send request and return response
func (h *HReq) Send(pathURLAndMethod ...string) (*Response, error) {
	return h.SendWithCtx(context.Background(), pathURLAndMethod...)
}

// MustSend send request and return response, will panic on error
func (h *HReq) MustSend(pathURLAndMethod ...string) *Response {
	resp, err := h.SendWithCtx(context.Background(), pathURLAndMethod...)
	if err != nil {
		panic(err)
	}

	return resp
}

// SendWithCtx request with context, then return response
func (h *HReq) SendWithCtx(ctx context.Context, pathURLAndMethod ...string) (*Response, error) {
	req, err := h.NewRequestWithCtx(ctx, pathURLAndMethod...)
	if err != nil {
		return nil, err
	}

	// do send
	return h.SendRequest(req)
}

// SendRequest send request
func (h *HReq) SendRequest(req *http.Request) (*Response, error) {
	// call before send.
	if h.beforeSend != nil {
		h.beforeSend(req)
	}

	// wrap middlewares
	h.wrapMiddlewares()

	// do send
	return h.handler(req)
}

// ----------- Build request ------------

// NewRequest build new request
func (h *HReq) NewRequest(pathURLAndMethod ...string) (*http.Request, error) {
	return h.NewRequestWithCtx(context.Background(), pathURLAndMethod...)
}

// NewRequestWithCtx build new request with context
func (h *HReq) NewRequestWithCtx(ctx context.Context, pathURLAndMethod ...string) (*http.Request, error) {
	pathURL := "/"
	if ln := len(pathURLAndMethod); ln > 0 {
		pathURL = pathURLAndMethod[0]
		if ln > 1 {
			h.method = pathURLAndMethod[1]
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

	// append query params
	if err = appendQueryParams(reqURL, h.queryParams); err != nil {
		return nil, err
	}

	var body io.Reader
	if h.bodyProvider != nil {
		body, err = h.bodyProvider.Body()
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequestWithContext(ctx, h.method, reqURL.String(), body)
	if err != nil {
		return nil, err
	}

	// copy headers
	httpreq.AddHeadersToRequest(req, h.header)

	return req, err
}

func appendQueryParams(reqURL *url.URL, uv url.Values) error {
	urlValues, err := url.ParseQuery(reqURL.RawQuery)
	if err != nil {
		return err
	}

	for key, values := range uv {
		for _, value := range values {
			urlValues.Add(key, value)
		}
	}

	// url.Values format to a sorted "url encoded" string.
	// e.g. "key=val&foo=bar"
	reqURL.RawQuery = urlValues.Encode()
	return nil
}
