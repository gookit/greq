package main

import (
	"io"
	"net/http"
	"net/http/httptest"
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

func resetCmdOpts() {
	cmdOpts.method = "GET"
	cmdOpts.data = ""
	cmdOpts.headers = cflag.KVString{Sep: ":"}
	cmdOpts.formData = cflag.KVString{Sep: "="}
	cmdOpts.jsonData = cflag.KVString{Sep: "="}
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
