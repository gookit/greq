package greq_test

import (
	"encoding/json"
	"io"
	"net/http"
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