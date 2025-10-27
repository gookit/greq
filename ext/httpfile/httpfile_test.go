package httpfile_test

import (
	"testing"

	"github.com/gookit/goutil/testutil/assert"
	"github.com/gookit/greq/ext/httpfile"
)

func TestParseFileContent(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    []*httpfile.HTTPRequest
		wantErr bool
	}{
		{
			name: "single request",
			content: `GET https://example.com/api
X-Custom-Header: custom-value

Request body`,
			want: []*httpfile.HTTPRequest{
				{
					Method: "GET",
					URL:    "https://example.com/api",
					Headers: map[string]string{
						"X-Custom-Header": "custom-value",
					},
					Body: "Request body",
					Comments: []string{},
				},
			},
			wantErr: false,
		},
		{
			name: "multiple requests",
			content: `### First Request
GET https://example.com/api1
X-Header1: value1

Body1

### Second Request
POST https://example.com/api2
Content-Type: application/json

{"key": "value"}`,
			want: []*httpfile.HTTPRequest{
				{
					Name: "First Request",
					Method: "GET",
					URL:    "https://example.com/api1",
					Headers: map[string]string{
						"X-Header1": "value1",
					},
					Body: "Body1",
					Comments: []string{},
				},
				{
					Name: "Second Request",
					Method: "POST",
					URL:    "https://example.com/api2",
					Headers: map[string]string{
						"Content-Type": "application/json",
					},
					Body: `{"key": "value"}`,
					Comments: []string{},
				},
			},
			wantErr: false,
		},
		{
			name: "requests with comments",
			content: `# This is a comment
### Request with Comments
GET https://example.com/api
# Header comment
X-Header: value

# Body comment
Request body`,
			want: []*httpfile.HTTPRequest{
				{
					Name: "Request with Comments",
					Method: "GET",
					URL:    "https://example.com/api",
					Headers: map[string]string{
						"X-Header": "value",
					},
					Body: "Request body",
					Comments: []string{"# This is a comment", "# Header comment", "# Body comment"},
				},
			},
			wantErr: false,
		},
		{
			name: "empty content",
			content: "",
			want:    []*httpfile.HTTPRequest{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hf, err := httpfile.ParseFileContent(tt.content)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, hf)
				assert.Equal(t, len(tt.want), len(hf.Requests))
				for i, wantReq := range tt.want {
					assert.Equal(t, wantReq.Method, hf.Requests[i].Method)
					assert.Equal(t, wantReq.URL, hf.Requests[i].URL)
					assert.Equal(t, wantReq.Name, hf.Requests[i].Name)
					assert.Equal(t, wantReq.Body, hf.Requests[i].Body)
					assert.Equal(t, wantReq.Headers, hf.Requests[i].Headers)
					assert.Equal(t, wantReq.Comments, hf.Requests[i].Comments)
				}
			}
		})
	}
}

func TestHTTPFile_Parse(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		contents string
		want     []*httpfile.HTTPRequest
		wantErr  bool
	}{
		{
			name:     "parse with contents",
			contents: `GET https://example.com/api
X-Header: value

Body`,
			want: []*httpfile.HTTPRequest{
				{
					Method: "GET",
					URL:    "https://example.com/api",
					Headers: map[string]string{
						"X-Header": "value",
					},
					Body: "Body",
					Comments: []string{},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hf := &httpfile.HTTPFile{
				FilePath: tt.filePath,
				Contents: tt.contents,
			}
			err := hf.Parse()
			if (err != nil) != tt.wantErr {
				t.Errorf("HTTPFile.Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				assert.Equal(t, len(tt.want), len(hf.Requests))
				for i, wantReq := range tt.want {
					assert.Equal(t, wantReq.Method, hf.Requests[i].Method)
					assert.Equal(t, wantReq.URL, hf.Requests[i].URL)
					assert.Equal(t, wantReq.Name, hf.Requests[i].Name)
					assert.Equal(t, wantReq.Body, hf.Requests[i].Body)
					assert.Equal(t, wantReq.Headers, hf.Requests[i].Headers)
				}
			}
		})
	}
}

func TestParseHTTPFile(t *testing.T) {
	hf, err := httpfile.ParseHTTPFile("testdata/test-req.http")
	assert.NoError(t, err)
	assert.NotEmpty(t, hf.Requests)

	reqs := hf.SearchName("request")
	assert.NotEmpty(t, reqs)
	assert.Equal(t, 2, len(reqs))

	req := hf.SearchOne("test", "request1")
	assert.NotEmpty(t, req)
	assert.Equal(t, "test request1", req.Name)
}
