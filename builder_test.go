package greq_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/gookit/goutil/testutil/assert"
	"github.com/gookit/greq"
)

func TestBuilder_NewBuilder(t *testing.T) {
	b := greq.NewBuilder()
	assert.NotNil(t, b)
	assert.NotNil(t, b.Options)
	assert.NotNil(t, b.Header)
	assert.NotNil(t, b.Query)
}

func TestBuilder_BuilderWithClient(t *testing.T) {
	client := greq.New("https://api.example.com")
	b := greq.BuilderWithClient(client)
	assert.NotNil(t, b)
	// Builder does not expose a Client() method, so we can't directly test it
	// We can test that the builder was created with the client by using it
}

func TestBuilder_WithMethod(t *testing.T) {
	b := greq.NewBuilder().WithMethod("POST")
	assert.Equal(t, "POST", b.Method)
}

func TestBuilder_PathURL(t *testing.T) {
	b := greq.NewBuilder().PathURL("/api/users")
	// pathURL is a private field, so we can't directly test it
	// Instead, we'll test through the Builder's behavior
	assert.NotNil(t, b)
	// We could test the actual URL through the Build method, but that requires more setup
}

func TestBuilder_QueryParams(t *testing.T) {
	// Test with map[string]string
	b := greq.NewBuilder().QueryParams(map[string]string{
		"name": "inhere",
		"age":  "30",
	})
	assert.Equal(t, "inhere", b.Query.Get("name"))
	assert.Equal(t, "30", b.Query.Get("age"))

	// Test with url.Values
	params := make(map[string][]string)
	params["city"] = []string{"Beijing"}
	b.QueryParams(params)
	assert.Equal(t, "Beijing", b.Query.Get("city"))
}

func TestBuilder_AddQuery(t *testing.T) {
	b := greq.NewBuilder().AddQuery("key", "value")
	assert.Equal(t, "value", b.Query.Get("key"))

	b.AddQuery("key", 123)
	values := b.Query["key"]
	assert.Len(t, values, 2)
	assert.Contains(t, values, "value")
	assert.Contains(t, values, "123")
}

func TestBuilder_HeaderOperations(t *testing.T) {
	b := greq.NewBuilder()

	// Test AddHeader
	b.AddHeader("X-Custom", "value1")
	assert.Equal(t, "value1", b.Header.Get("X-Custom"))

	// Test SetHeader
	b.SetHeader("X-Custom", "value2")
	assert.Equal(t, "value2", b.Header.Get("X-Custom"))

	// Test AddHeaders
	headers := make(http.Header)
	headers.Add("X-Test1", "val1")
	headers.Add("X-Test2", "val2")
	b.AddHeaders(headers)
	assert.Equal(t, "val1", b.Header.Get("X-Test1"))
	assert.Equal(t, "val2", b.Header.Get("X-Test2"))

	// Test SetHeaders
	newHeaders := make(http.Header)
	newHeaders.Set("X-Test1", "new-val1")
	newHeaders.Set("X-Test2", "new-val2")
	b.SetHeaders(newHeaders)
	assert.Equal(t, "new-val1", b.Header.Get("X-Test1"))
	assert.Equal(t, "new-val2", b.Header.Get("X-Test2"))
	b.RemoveHeaders("X-Test1", "X-Test2")
	assert.Equal(t, "", b.Header.Get("X-Test1"))
	assert.Equal(t, "", b.Header.Get("X-Test2"))

	// Test AddHeaderMap
	b.AddHeaderMap(map[string]string{
		"X-Map1": "map-val1",
		"X-Map2": "map-val2",
	})
	assert.Equal(t, "map-val1", b.Header.Get("X-Map1"))
	assert.Equal(t, "map-val2", b.Header.Get("X-Map2"))

	// Test SetHeaderMap
	b.SetHeaderMap(map[string]string{
		"X-Map1": "new-map-val1",
		"X-Map2": "",
	})
	assert.Equal(t, "new-map-val1", b.Header.Get("X-Map1"))
	assert.Equal(t, "", b.Header.Get("X-Map2"))
}

func TestBuilder_UserAgent(t *testing.T) {
	b := greq.NewBuilder().UserAgent("test-agent/1.0")
	assert.Equal(t, "test-agent/1.0", b.HeaderM["User-Agent"])
}

func TestBuilder_BasicAuth(t *testing.T) {
	b := greq.NewBuilder().BasicAuth("user", "pass")
	authHeader := b.Header.Get("Authorization")
	assert.NotEmpty(t, authHeader)
	assert.Contains(t, authHeader, "Basic")
}

func TestBuilder_ContentType(t *testing.T) {
	b := greq.NewBuilder()

	// Test WithContentType
	b.WithContentType("application/custom")
	assert.Equal(t, "application/custom", b.ContentType)

	// Test JSONType
	b.JSONType()
	assert.Equal(t, "application/json; charset=utf-8", b.ContentType)

	// Test XMLType
	b.XMLType()
	assert.Equal(t, "application/xml; charset=utf-8", b.ContentType)

	// Test FormType
	b.FormType()
	assert.Equal(t, "application/x-www-form-urlencoded; charset=utf-8", b.ContentType)

	// Test MultipartType
	b.MultipartType()
	assert.Equal(t, "multipart/form-data", b.ContentType)
}

func TestBuilder_BodyOperations(t *testing.T) {
	b := greq.NewBuilder()

	// Test StringBody
	b.StringBody("test body")
	assert.NotNil(t, b.Provider)
	reader, err := b.Provider.Body()
	assert.NoErr(t, err)
	body, _ := io.ReadAll(reader)
	assert.Equal(t, "test body", string(body))

	// Test BytesBody
	b.BytesBody([]byte("bytes body"))
	reader, err = b.Provider.Body()
	assert.NoErr(t, err)
	body, _ = io.ReadAll(reader)
	assert.Equal(t, "bytes body", string(body))

	// Test JSONBody
	data := map[string]string{"key": "value"}
	b.JSONBody(data)
	assert.NotNil(t, b.Provider)
	reader, err = b.Provider.Body()
	assert.NoErr(t, err)
	// Check if it's valid JSON
	var result map[string]string
	err = json.NewDecoder(reader).Decode(&result)
	assert.NoErr(t, err)
	assert.Equal(t, "value", result["key"])

	// Test FormBody
	formData := map[string][]string{
		"name": {"inhere"},
		"age":  {"30"},
	}
	b.FormBody(formData)
	reader, err = b.Provider.Body()
	assert.NoErr(t, err)
	body, _ = io.ReadAll(reader)
	assert.Contains(t, string(body), "name=inhere")
	assert.Contains(t, string(body), "age=30")
}

func TestBuilder_Build(t *testing.T) {
	b := greq.NewBuilder().
		WithMethod("GET").
		PathURL("/test")

	req, err := b.Build("GET", "/test")
	assert.NoErr(t, err)
	assert.NotNil(t, req)
	assert.Equal(t, "GET", req.Method)
	assert.Equal(t, "/test", req.URL.Path)
}

func TestBuilder_HTTPMethods(t *testing.T) {
	b := greq.NewBuilder()

	// Test Get
	b.Get("/get")
	assert.Equal(t, "GET", b.Method)
	// Note: pathURL is not directly accessible, we test through other means

	// Test Post
	b.Post("/post", "data")
	assert.Equal(t, "POST", b.Method)
	assert.Equal(t, "data", b.Body)

	// Test Put
	b.Put("/put", "data")
	assert.Equal(t, "PUT", b.Method)
	assert.Equal(t, "data", b.Body)

	// Test Delete
	b.Delete("/delete")
	assert.Equal(t, "DELETE", b.Method)

	// Test Patch
	b.Patch("/patch", "data")
	assert.Equal(t, "PATCH", b.Method)
	assert.Equal(t, "data", b.Body)
}

func TestBuilder_String(t *testing.T) {
	b := greq.NewBuilder().
		WithMethod("GET").
		PathURL("/test").
		AddHeader("X-Custom", "value")

	str := b.String()
	assert.NotEmpty(t, str)
	assert.Contains(t, str, "GET")
	assert.Contains(t, str, "/test")
	assert.Contains(t, str, "X-Custom")
}
func TestBodyProvider_Interface(t *testing.T) {
	// Test that BodyProvider interface is properly defined
	// Using the actual constructor functions from builder.go
	b := greq.NewBuilder()
	b.BodyReader(strings.NewReader("test"))
	assert.NotNil(t, b.Provider)
	assert.Equal(t, "", b.Provider.ContentType())

	b.JSONBody(map[string]string{"key": "value"})
	assert.NotNil(t, b.Provider)
	assert.Equal(t, "application/json; charset=utf-8", b.Provider.ContentType())

	b.FormBody(url.Values{"key": []string{"value"}})
	assert.NotNil(t, b.Provider)
	assert.Equal(t, "application/x-www-form-urlencoded; charset=utf-8", b.Provider.ContentType())
}

func TestBodyProvider_ContentType(t *testing.T) {
	// Test bodyProvider ContentType
	b := greq.NewBuilder().BodyReader(strings.NewReader("test"))
	assert.Equal(t, "", b.Provider.ContentType())

	ir, err := b.Provider.Body()
	assert.NoErr(t, err)
	assert.NotNil(t, ir)
	bs, err := io.ReadAll(ir)
	assert.NoErr(t, err)
	assert.Equal(t, "test", string(bs))
}

func TestBodyProvider_Body(t *testing.T) {
	reader := strings.NewReader("test body")
	b := greq.NewBuilder().BodyReader(reader)

	bodyReader, err := b.Provider.Body()
	assert.NoErr(t, err)
	assert.NotNil(t, bodyReader)

	body, _ := io.ReadAll(bodyReader)
	assert.Equal(t, "test body", string(body))
}

func TestJsonBodyProvider_ContentType(t *testing.T) {
	b := greq.NewBuilder().JSONBody(map[string]string{"key": "value"})
	assert.Equal(t, "application/json; charset=utf-8", b.Provider.ContentType())
}

func TestJsonBodyProvider_Body(t *testing.T) {
	data := map[string]string{
		"name": "inhere",
		"age":  "30",
	}

	b := greq.NewBuilder().JSONBody(data)
	reader, err := b.Provider.Body()
	assert.NoErr(t, err)
	assert.NotNil(t, reader)

	// Read and decode JSON
	body, _ := io.ReadAll(reader)
	var result map[string]string
	err = json.Unmarshal(body, &result)
	assert.NoErr(t, err)
	assert.Equal(t, "inhere", result["name"])
	assert.Equal(t, "30", result["age"])
}

func TestFormBodyProvider_ContentType(t *testing.T) {
	b := greq.NewBuilder().FormBody(url.Values{})
	assert.Equal(t, "application/x-www-form-urlencoded; charset=utf-8", b.Provider.ContentType())
}

func TestFormBodyProvider_Body_WithUrlValues(t *testing.T) {
	values := make(url.Values)
	values.Set("name", "inhere")
	values.Set("age", "30")

	b := greq.NewBuilder().FormBody(values)
	reader, err := b.Provider.Body()
	assert.NoErr(t, err)
	assert.NotNil(t, reader)

	body, _ := io.ReadAll(reader)
	bodyStr := string(body)
	assert.Contains(t, bodyStr, "name=inhere")
	assert.Contains(t, bodyStr, "age=30")
}

func TestFormBodyProvider_Body_WithMapStringSlice(t *testing.T) {
	data := make(map[string][]string)
	data["name"] = []string{"inhere"}
	data["age"] = []string{"30"}

	b := greq.NewBuilder().FormBody(data)
	reader, err := b.Provider.Body()
	assert.NoErr(t, err)
	assert.NotNil(t, reader)

	body, _ := io.ReadAll(reader)
	bodyStr := string(body)
	assert.Contains(t, bodyStr, "name=inhere")
	assert.Contains(t, bodyStr, "age=30")
}

func TestFormBodyProvider_Body_WithString(t *testing.T) {
	data := "name=inhere&age=30"

	b := greq.NewBuilder().FormBody(data)
	reader, err := b.Provider.Body()
	assert.NoErr(t, err)
	assert.NotNil(t, reader)

	body, _ := io.ReadAll(reader)
	assert.Equal(t, data, string(body))
}

func TestFormBodyProvider_WithInvalidData(t *testing.T) {
	// We can't directly create formBodyProvider since it's not exported
	// Instead, we'll test the behavior through the Builder's FormBody method
	// which should handle invalid data gracefully
	b := greq.NewBuilder().FormBody(12345) // Passing an invalid type (int)

	reader, err := b.Provider.Body()
	assert.Err(t, err)
	assert.Nil(t, reader)
	assert.ErrMsg(t, err, "formBodyProvider: invalid form data type: int")
}
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