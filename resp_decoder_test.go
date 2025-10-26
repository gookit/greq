package greq

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/gookit/goutil/testutil/assert"
)

func TestJsonDecoder_Decode(t *testing.T) {
	// Create test JSON data
	testData := map[string]any{
		"name": "inhere",
		"age":  30,
		"city": "Beijing",
	}

	jsonStr, _ := json.Marshal(testData)
	reader := bytes.NewReader(jsonStr)

	// Create mock http.Response
	resp := &http.Response{
		Body: io.NopCloser(reader),
	}

	// Test decoding
	decoder := jsonDecoder{}
	var result map[string]any
	err := decoder.Decode(resp, &result)

	assert.NoErr(t, err)
	assert.Equal(t, "inhere", result["name"])
	assert.Equal(t, 30.0, result["age"]) // JSON numbers are float64 by default
	assert.Equal(t, "Beijing", result["city"])

}

func TestJsonDecoder_Decode_Error(t *testing.T) {
	// Create invalid JSON data
	reader := strings.NewReader("{invalid json}")

	// Create mock http.Response
	resp := &http.Response{
		Body: io.NopCloser(reader),
	}

	// Test decoding with invalid JSON
	decoder := jsonDecoder{}  // This won't work as jsonDecoder is private
	var result map[string]any
	err := decoder.Decode(resp, &result)

	assert.Err(t, err)
	assert.NotNil(t, err)

}

func TestXmlDecoder_Decode(t *testing.T) {
	// Create test XML data
	type Person struct {
		Name string `xml:"name"`
		Age  int    `xml:"age"`
		City string `xml:"city"`
	}

	testData := Person{
		Name: "inhere",
		Age:  30,
		City: "Beijing",
	}

	xmlData, _ := xml.Marshal(testData)
	reader := bytes.NewReader(xmlData)

	// Create mock http.Response
	resp := &http.Response{
		Body: io.NopCloser(reader),
	}

	// Test decoding
	decoder := XmlDecoder{}
	var result Person
	err := decoder.Decode(resp, &result)

	assert.NoErr(t, err)
	assert.Equal(t, "inhere", result.Name)
	assert.Equal(t, 30, result.Age)
	assert.Equal(t, "Beijing", result.City)
}

func TestXmlDecoder_Decode_Error(t *testing.T) {
	// Create invalid XML data
	reader := strings.NewReader("<invalid xml>")

	// Create mock http.Response
	resp := &http.Response{
		Body: io.NopCloser(reader),
	}

	// Test decoding with invalid XML
	decoder := XmlDecoder{}
	var result map[string]interface{}
	err := decoder.Decode(resp, &result)

	assert.Err(t, err)
	assert.NotNil(t, err)
}