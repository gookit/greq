package hireq

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
)

// RespDecoder decodes http responses into struct values.
type RespDecoder interface {
	// Decode decodes the response into the value pointed to by ptr.
	Decode(resp *http.Response, ptr interface{}) error
}

// jsonDecoder decodes http response JSON into a JSON-tagged struct value.
type jsonDecoder struct {
}

// Decode decodes the Response Body into the value pointed to by ptr.
// Caller must provide a non-nil v and close the resp.Body.
func (d jsonDecoder) Decode(resp *http.Response, ptr interface{}) error {
	return json.NewDecoder(resp.Body).Decode(ptr)
}

// xmlDecoder decodes http response JSON into a XML-tagged struct value.
type xmlDecoder struct {
}

// Decode decodes the Response Body into the value pointed to by ptr.
func (d xmlDecoder) Decode(resp *http.Response, ptr interface{}) error {
	return xml.NewDecoder(resp.Body).Decode(ptr)
}
