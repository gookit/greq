package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gookit/goutil/cflag"
	"github.com/gookit/goutil/maputil"
	"github.com/gookit/goutil/netutil/httpctype"
	"github.com/gookit/goutil/x/assert"
)

func TestBuildJSONBody(t *testing.T) {
	body, err := buildJSONBody(maputil.SMap{
		"name": "inhere",
		"age":  "18",
	})

	assert.NoErr(t, err)
	assert.Eq(t, `{"age":"18","name":"inhere"}`, string(body))
}

func TestBuildJSONBody_Empty(t *testing.T) {
	body, err := buildJSONBody(nil)

	assert.NoErr(t, err)
	assert.Eq(t, 0, len(body))
}

func TestHandleNormalRequest_JSONFields(t *testing.T) {
	var reqMethod, reqContentType, reqBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		assert.NoErr(t, err)

		reqMethod = r.Method
		reqContentType = r.Header.Get("Content-Type")
		reqBody = string(body)
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	resetCmdOpts()
	cmdOpts.silent = true
	cmdOpts.output = filepath.Join(t.TempDir(), "resp.txt")
	cmdOpts.jsonData.Set("name=inhere")
	cmdOpts.jsonData.Set("age=18")

	err := handleNormalRequest(server.URL)

	assert.NoErr(t, err)
	assert.Eq(t, http.MethodPost, reqMethod)
	assert.Eq(t, httpctype.JSON, reqContentType)
	assert.Eq(t, `{"age":"18","name":"inhere"}`, reqBody)
}

func TestHandleNormalRequest_JSONTypeWithRawData(t *testing.T) {
	var reqContentType, reqBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		assert.NoErr(t, err)

		reqContentType = r.Header.Get("Content-Type")
		reqBody = string(body)
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	resetCmdOpts()
	cmdOpts.silent = true
	cmdOpts.output = filepath.Join(t.TempDir(), "resp.txt")
	cmdOpts.jsonType = true
	cmdOpts.data = `{"name":"inhere"}`

	err := handleNormalRequest(server.URL)

	assert.NoErr(t, err)
	assert.Eq(t, httpctype.JSON, reqContentType)
	assert.Eq(t, `{"name":"inhere"}`, reqBody)
}

func TestHandleNormalRequest_HeaderContentType(t *testing.T) {
	var reqContentType string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqContentType = r.Header.Get("Content-Type")
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	resetCmdOpts()
	cmdOpts.silent = true
	cmdOpts.output = filepath.Join(t.TempDir(), "resp.txt")
	cmdOpts.data = "hello"
	cmdOpts.headers.Set("Content-Type: text/plain")

	err := handleNormalRequest(server.URL)

	assert.NoErr(t, err)
	assert.Eq(t, "text/plain", reqContentType)
}

func TestHandleNormalRequest_CustomHeader(t *testing.T) {
	var reqToken string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqToken = r.Header.Get("X-Token")
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	resetCmdOpts()
	cmdOpts.silent = true
	cmdOpts.output = filepath.Join(t.TempDir(), "resp.txt")
	cmdOpts.headers.Set("X-Token: abc")

	err := handleNormalRequest(server.URL)

	assert.NoErr(t, err)
	assert.Eq(t, "abc", reqToken)
}

func TestHandleNormalRequest_DataFromFile(t *testing.T) {
	var reqMethod, reqBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		assert.NoErr(t, err)

		reqMethod = r.Method
		reqBody = string(body)
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	bodyFile := filepath.Join(t.TempDir(), "body.json")
	err := os.WriteFile(bodyFile, []byte(`{"name":"inhere"}`), 0644)
	assert.NoErr(t, err)

	resetCmdOpts()
	cmdOpts.silent = true
	cmdOpts.output = filepath.Join(t.TempDir(), "resp.txt")
	cmdOpts.data = "@" + bodyFile

	err = handleNormalRequest(server.URL)

	assert.NoErr(t, err)
	assert.Eq(t, http.MethodPost, reqMethod)
	assert.Eq(t, `{"name":"inhere"}`, reqBody)
}

func TestHandleNormalRequest_UploadFileWithFormData(t *testing.T) {
	var reqMethod, reqContentType, reqToken, fileBody, formName string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqMethod = r.Method
		reqContentType = r.Header.Get("Content-Type")
		reqToken = r.Header.Get("X-Token")
		err := r.ParseMultipartForm(10 << 20)
		assert.NoErr(t, err)

		file, _, err := r.FormFile("file")
		assert.NoErr(t, err)
		defer file.Close()

		body, err := io.ReadAll(file)
		assert.NoErr(t, err)
		fileBody = string(body)
		formName = r.FormValue("name")
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	uploadFile := filepath.Join(t.TempDir(), "upload.txt")
	err := os.WriteFile(uploadFile, []byte("hello upload"), 0644)
	assert.NoErr(t, err)

	resetCmdOpts()
	cmdOpts.silent = true
	cmdOpts.output = filepath.Join(t.TempDir(), "resp.txt")
	cmdOpts.uploadFiles.Set("file=" + uploadFile)
	cmdOpts.formData.Set("name=inhere")
	cmdOpts.headers.Set("X-Token: abc")

	err = handleNormalRequest(server.URL)

	assert.NoErr(t, err)
	assert.Eq(t, http.MethodPost, reqMethod)
	assert.StrContains(t, reqContentType, "multipart/form-data")
	assert.Eq(t, "abc", reqToken)
	assert.Eq(t, "hello upload", fileBody)
	assert.Eq(t, "inhere", formName)
}

func TestHandleNormalRequest_UploadFileWithPutMethod(t *testing.T) {
	var reqMethod string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqMethod = r.Method
		err := r.ParseMultipartForm(10 << 20)
		assert.NoErr(t, err)
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	uploadFile := filepath.Join(t.TempDir(), "upload.txt")
	err := os.WriteFile(uploadFile, []byte("hello upload"), 0644)
	assert.NoErr(t, err)

	resetCmdOpts()
	cmdOpts.method = http.MethodPut
	cmdOpts.silent = true
	cmdOpts.output = filepath.Join(t.TempDir(), "resp.txt")
	cmdOpts.uploadFiles.Set("file=" + uploadFile)

	err = handleNormalRequest(server.URL)

	assert.NoErr(t, err)
	assert.Eq(t, http.MethodPut, reqMethod)
}

func resetCmdOpts() {
	cmdOpts.method = "GET"
	cmdOpts.data = ""
	cmdOpts.headers = cflag.KVString{Sep: ":"}
	cmdOpts.formData = cflag.KVString{Sep: "="}
	cmdOpts.jsonData = cflag.KVString{Sep: "="}
	cmdOpts.uploadFiles = cflag.KVString{Sep: "="}
	cmdOpts.timeout = 30
	cmdOpts.output = ""
	cmdOpts.raw = ""
	cmdOpts.httpVars = cflag.KVString{Sep: "="}
	cmdOpts.down = false
	cmdOpts.verbose = false
	cmdOpts.silent = false
	cmdOpts.follow = false
	cmdOpts.insecure = false
	cmdOpts.jsonType = false
	cmdOpts.agent = ""
	cmdOpts.headOnly = false
}
