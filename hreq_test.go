package hreq_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gookit/goutil/dump"
	"github.com/gookit/goutil/jsonutil"
	"github.com/gookit/goutil/netutil/httpreq"
	"github.com/gookit/hreq"
	"github.com/stretchr/testify/assert"
)

var testBaseURL = "https://httpbin.org"

func TestHReq_Send(t *testing.T) {
	tw := httptest.NewRecorder()
	buf := &bytes.Buffer{}

	doer := httpreq.HttpDoerFunc(func(req *http.Request) (*http.Response, error) {
		dump.P("CORE+")
		_, err := tw.WriteString("TEST")
		dump.P("CORE-")

		return tw.Result(), err
	})

	mid0 := hreq.MiddleFunc(func(r *http.Request, next hreq.NextFunc) (*http.Response, error) {
		dump.P("MID0+")
		w, err := next(r)
		dump.P("MID0-")
		return w, err
	})

	resp, err := hreq.New(testBaseURL).
		Doer(doer).
		Use(mid0).
		Get("/get")

	assert.NoError(t, err)
	sc := resp.StatusCode
	assert.True(t, httpreq.IsOK(sc))

	err = resp.Write(buf)
	assert.NoError(t, err)

	dump.P(buf.String())
}

func TestHReq_Get(t *testing.T) {
	resp, err := hreq.New(testBaseURL).
		JSONType().
		Get("/get")

	assert.NoError(t, err)
	sc := resp.StatusCode
	assert.True(t, httpreq.IsOK(sc))
	assert.True(t, httpreq.IsSuccessful(sc))

	retMp := make(map[string]interface{})
	err = jsonutil.DecodeReader(resp.Body, &retMp)
	assert.NoError(t, err)
	dump.P(retMp)
}
