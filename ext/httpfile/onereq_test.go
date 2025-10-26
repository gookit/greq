package httpfile_test

import (
	"testing"

	"github.com/gookit/goutil/testutil/assert"
	"github.com/gookit/greq/ext/httpfile"
)

func TestParseOneRequest(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    *httpfile.HTTPRequest
		wantErr bool
	}{
		{
			name: "simple GET request",
			content: "GET https://example.com/api",
			want: &httpfile.HTTPRequest{
				Method: "GET",
				URL:    "https://example.com/api",
				Headers: make(map[string]string),
				Comments: []string{},
			},
			wantErr: false,
		},
		{
			name: "POST request with headers and body",
			content: `POST https://example.com/api
Content-Type: application/json
Authorization: Bearer token

{"name": "test", "value": 123}`,
			want: &httpfile.HTTPRequest{
				Method: "POST",
				URL:    "https://example.com/api",
				Headers: map[string]string{
					"Content-Type":  "application/json",
					"Authorization": "Bearer token",
				},
				Body: `{"name": "test", "value": 123}`,
				Comments: []string{},
			},
			wantErr: false,
		},
		{
			name: "request with comments",
			content: `# This is a comment
### Request Name
GET https://example.com/api
# Another comment
X-Custom-Header: custom-value`,
			want: &httpfile.HTTPRequest{
				Name: "Request Name",
				Method: "GET",
				URL:    "https://example.com/api",
				Headers: map[string]string{
					"X-Custom-Header": "custom-value",
				},
				Comments: []string{"# This is a comment", "# Another comment"},
			},
			wantErr: false,
		},
		{
			name: "request with body containing empty lines",
			content: `POST https://example.com/api
Content-Type: text/plain

Line 1

Line 2`,
			want: &httpfile.HTTPRequest{
				Method: "POST",
				URL:    "https://example.com/api",
				Headers: map[string]string{
					"Content-Type": "text/plain",
				},
				Body: "Line 1\n\nLine 2",
				Comments: []string{},
			},
			wantErr: false,
		},
		{
			name:    "invalid request - missing method and URL",
			content: "Content-Type: application/json",
			wantErr: true,
		},
		{
			name:    "empty content",
			content: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := httpfile.ParseOneRequest(tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseOneRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				assert.Equal(t, tt.want.Method, got.Method)
				assert.Equal(t, tt.want.URL, got.URL)
				assert.Equal(t, tt.want.Name, got.Name)
				assert.Equal(t, tt.want.Body, got.Body)
				assert.Equal(t, tt.want.Headers, got.Headers)
				assert.Equal(t, tt.want.Comments, got.Comments)
			}
		})
	}
}
