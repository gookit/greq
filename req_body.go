package greq

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/gookit/goutil/netutil/httpctype"
)

// BodyProvider provides Body content for http.Request attachment.
type BodyProvider interface {
	// ContentType returns the Content-Type of the body.
	ContentType() string
	// Body returns the io.Reader body.
	Body() (io.Reader, error)
}

// bodyProvider provides the wrapped body value as a Body for reqests.
type bodyProvider struct {
	body io.Reader
}

// ContentType value
func (p bodyProvider) ContentType() string {
	return ""
}

// Body get body reader
func (p bodyProvider) Body() (io.Reader, error) {
	return p.body, nil
}

// jsonBodyProvider encodes a JSON tagged struct value as a Body for requests.
type jsonBodyProvider struct {
	payload any
}

// ContentType value
func (p jsonBodyProvider) ContentType() string {
	return httpctype.JSON
}

// Body get body reader
func (p jsonBodyProvider) Body() (io.Reader, error) {
	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).Encode(p.payload)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

// formBodyProvider encodes a url tagged struct value as Body for requests.
type formBodyProvider struct {
	// allow type: string, url.Values
	payload any
}

// ContentType value
func (p formBodyProvider) ContentType() string {
	return httpctype.Form
}

// Body get body reader
func (p formBodyProvider) Body() (io.Reader, error) {
	values, ok := p.payload.(url.Values)
	if ok {
		return strings.NewReader(values.Encode()), nil
	}

	mps, ok := p.payload.(map[string][]string)
	if ok {
		return strings.NewReader(url.Values(mps).Encode()), nil
	}

	if str, ok := p.payload.(string); ok {
		return strings.NewReader(str), nil
	}
	return nil, fmt.Errorf("formBodyProvider: invalid form data type: %T", p.payload)
}

type multipartBodyProvider struct {
	files       map[string]string
	fields      map[string]string
	contentType string
	body        *bytes.Buffer
}

// ContentType value
func (p *multipartBodyProvider) ContentType() string {
	return p.contentType
}

// Body data build
func (p *multipartBodyProvider) Body() (io.Reader, error) {
	if p.body == nil {
		if err := p.build(); err != nil {
			return nil, err
		}
	}
	return p.body, nil
}

func (p *multipartBodyProvider) build() error {
	buf := &bytes.Buffer{}
	writer := multipart.NewWriter(buf)

	for fieldName, filePath := range p.files {
		file, err := os.Open(filePath)
		if err != nil {
			return fmt.Errorf("open file %s failed: %w", filePath, err)
		}

		part, err := writer.CreateFormFile(fieldName, filepath.Base(filePath))
		if err != nil {
			file.Close()
			return fmt.Errorf("create form file %s failed: %w", fieldName, err)
		}

		_, err = io.Copy(part, file)
		file.Close()
		if err != nil {
			return fmt.Errorf("copy file %s failed: %w", filePath, err)
		}
	}

	for key, value := range p.fields {
		if err := writer.WriteField(key, value); err != nil {
			return fmt.Errorf("write field %s failed: %w", key, err)
		}
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("close multipart writer failed: %w", err)
	}

	p.contentType = writer.FormDataContentType()
	p.body = buf
	return nil
}
