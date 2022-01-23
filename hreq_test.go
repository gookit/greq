package hreq_test

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gookit/goutil/dump"
	"github.com/gookit/goutil/fsutil"
	"github.com/gookit/goutil/netutil/httpreq"
	"github.com/gookit/hreq"
	"github.com/stretchr/testify/assert"
)

var testBaseURL = "https://httpbin.org"

var testDoer = httpreq.DoerFunc(func(req *http.Request) (*http.Response, error) {
	tw := httptest.NewRecorder()
	dump.P("CORE++")

	_, err := tw.WriteString(req.RequestURI + " > ")
	if err != nil {
		return nil, err
	}

	if req.Body != nil {
		_, err = tw.Write(fsutil.MustReadReader(req.Body))
	}

	dump.P("CORE--")
	return tw.Result(), err
})

func TestHReq_Doer(t *testing.T) {
	buf := &bytes.Buffer{}

	mid0 := hreq.MiddleFunc(func(r *http.Request, next hreq.HandleFunc) (*hreq.Response, error) {
		dump.P("MID0++")
		w, err := next(r)
		dump.P("MID0--")
		return w, err
	})

	resp, err := hreq.New(testBaseURL).
		Doer(testDoer).
		Use(mid0).
		UserAgent("custom-client/1.0").
		Get("/get")

	assert.NoError(t, err)
	assert.True(t, resp.IsOK())

	err = resp.Write(buf)
	assert.NoError(t, err)
	dump.P(buf.String())
}

func TestHReq_Send(t *testing.T) {
	resp, err := hreq.New(testBaseURL).
		UserAgent("custom-client/1.0").
		Send("/get")

	assert.NoError(t, err)
	assert.True(t, resp.IsOK())
	assert.False(t, resp.IsFail())

	retMp := make(map[string]interface{})
	err = resp.Decode(&retMp)
	assert.NoError(t, err)
	dump.P(retMp)

	assert.Contains(t, retMp, "headers")

	headers := retMp["headers"].(map[string]interface{})
	assert.Contains(t, headers, "User-Agent")
	assert.Equal(t, "custom-client/1.0", headers["User-Agent"])
}

func TestHReq_Get(t *testing.T) {
	resp, err := hreq.New(testBaseURL).
		JSONType().
		Get("/get")

	assert.NoError(t, err)
	assert.True(t, resp.IsOK())
	assert.True(t, resp.IsSuccessful())

	retMp := make(map[string]interface{})
	err = resp.Decode(&retMp)
	assert.NoError(t, err)
	dump.P(retMp)
}

func TestHReq_Post(t *testing.T) {
	resp, err := hreq.New(testBaseURL).
		JSONType().
		Post("/post")

	assert.NoError(t, err)
	assert.True(t, resp.IsOK())
	assert.True(t, resp.IsSuccessful())

	retMp := make(map[string]interface{})
	err = resp.Decode(&retMp)
	assert.NoError(t, err)
	dump.P(retMp)
}

func TestHReq_String(t *testing.T) {
	str := hreq.New(testBaseURL+"/post").
		UserAgent("some-client/1.0").
		BasicAuth("inhere", "some string").
		JSONType().
		Body("hi, with body").
		Method("POST").
		String()

	fmt.Println(str)
}
