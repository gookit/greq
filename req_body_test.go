package greq_test

import (
	"encoding/json"
	"io"
	"net/url"
	"strings"
	"testing"

	"github.com/gookit/goutil/testutil/assert"
	"github.com/gookit/greq"
)

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
