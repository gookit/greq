package hreq_test

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/gookit/goutil/dump"
	"github.com/gookit/goutil/netutil/httpreq"
	"github.com/gookit/hreq"
	"github.com/stretchr/testify/assert"
)

type mid1 struct {
}

func (m mid1) Handle(r *http.Request, next hreq.HandleFunc) (*hreq.Response, error) {
	dump.P("MID1++")
	w, err := next(r)
	dump.P("MID1--")
	return w, err
}

func TestHReq_Use_Middleware(t *testing.T) {
	buf := &bytes.Buffer{}

	mid2 := func(r *http.Request, next hreq.HandleFunc) (*hreq.Response, error) {
		dump.P("MID2++")
		w, err := next(r)
		dump.P("MID2--")
		return w, err
	}

	resp, err := hreq.New(testBaseURL).
		Doer(testDoer).
		Uses(&mid1{}, hreq.MiddleFunc(mid2)).
		Get("/get")

	assert.NoError(t, err)
	sc := resp.StatusCode
	assert.True(t, httpreq.IsOK(sc))

	err = resp.Write(buf)
	assert.NoError(t, err)

	dump.P(buf.String())
}

func TestHReq_Use_MiddleFunc(t *testing.T) {
	mid0 := hreq.MiddleFunc(func(r *http.Request, next hreq.HandleFunc) (*hreq.Response, error) {
		dump.P("MID0++")
		w, err := next(r)
		dump.P("MID0--")
		return w, err
	})

	resp, err := hreq.New(testBaseURL).
		Doer(testDoer).
		Middleware(mid0).
		Get("/get")

	assert.NoError(t, err)
	sc := resp.StatusCode
	assert.True(t, httpreq.IsOK(sc))

}
