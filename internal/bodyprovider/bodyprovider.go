// Package bodyprovider implements the concrete BodyProvider variants used by
// the greq client (reader, JSON, form, multipart). The greq.BodyProvider
// interface is structurally satisfied by every type in this package.
package bodyprovider

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

// Reader wraps an io.Reader and exposes it as a body without a Content-Type.
type Reader struct {
	body io.Reader
}

// NewReader returns a Reader provider.
func NewReader(r io.Reader) Reader { return Reader{body: r} }

// ContentType returns the Content-Type — empty for raw reader bodies.
func (p Reader) ContentType() string { return "" }

// Body returns the underlying reader.
func (p Reader) Body() (io.Reader, error) { return p.body, nil }

// JSON encodes a JSON tagged struct value as request body.
type JSON struct {
	payload any
}

// NewJSON returns a JSON provider.
func NewJSON(payload any) JSON { return JSON{payload: payload} }

// ContentType returns application/json.
func (p JSON) ContentType() string { return httpctype.JSON }

// Body marshals the payload into a JSON reader.
func (p JSON) Body() (io.Reader, error) {
	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(p.payload); err != nil {
		return nil, err
	}
	return buf, nil
}

// Form encodes a url tagged struct value (or string/url.Values/map[string][]string) as form body.
type Form struct {
	payload any
}

// NewForm returns a Form provider.
func NewForm(payload any) Form { return Form{payload: payload} }

// ContentType returns application/x-www-form-urlencoded.
func (p Form) ContentType() string { return httpctype.Form }

// Body encodes the payload into a form-urlencoded reader.
func (p Form) Body() (io.Reader, error) {
	if values, ok := p.payload.(url.Values); ok {
		return strings.NewReader(values.Encode()), nil
	}
	if mps, ok := p.payload.(map[string][]string); ok {
		return strings.NewReader(url.Values(mps).Encode()), nil
	}
	if str, ok := p.payload.(string); ok {
		return strings.NewReader(str), nil
	}
	return nil, fmt.Errorf("formBodyProvider: invalid form data type: %T", p.payload)
}

// Multipart builds a multipart/form-data body from files and fields.
// Uses a pointer receiver because Body() lazily materializes the buffer.
type Multipart struct {
	files       map[string]string
	fields      map[string]string
	contentType string
	body        *bytes.Buffer
}

// NewMultipart returns a Multipart provider. Either map may be nil.
func NewMultipart(files, fields map[string]string) *Multipart {
	return &Multipart{files: files, fields: fields}
}

// ContentType returns the multipart Content-Type including the boundary.
// Returns empty string until Body() has been called at least once.
func (p *Multipart) ContentType() string { return p.contentType }

// Body lazily builds the multipart body on first call.
func (p *Multipart) Body() (io.Reader, error) {
	if p.body == nil {
		if err := p.build(); err != nil {
			return nil, err
		}
	}
	return p.body, nil
}

func (p *Multipart) build() error {
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
