package greq_test

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gookit/goutil/dump"
	"github.com/gookit/goutil/netutil/httpreq"
	"github.com/gookit/greq"
	"github.com/stretchr/testify/assert"
)

type mid1 struct {
}

func (m mid1) Handle(r *http.Request, next greq.HandleFunc) (*greq.Response, error) {
	dump.P("MID1++")
	w, err := next(r)
	dump.P("MID1--")
	return w, err
}

func TestHReq_Use_Middleware(t *testing.T) {
	buf := &bytes.Buffer{}

	mid2 := func(r *http.Request, next greq.HandleFunc) (*greq.Response, error) {
		dump.P("MID2++")
		w, err := next(r)
		dump.P("MID2--")
		return w, err
	}

	resp, err := greq.New(testBaseURL).
		Doer(testDoer).
		Uses(&mid1{}, greq.MiddleFunc(mid2)).
		GetDo("/get")

	assert.NoError(t, err)
	assert.True(t, resp.IsOK())

	err = resp.Write(buf)
	assert.NoError(t, err)

	dump.P(buf.String())
}

func TestHReq_Use_MiddleFunc(t *testing.T) {
	mid0 := greq.MiddleFunc(func(r *http.Request, next greq.HandleFunc) (*greq.Response, error) {
		dump.P("MID0++")
		w, err := next(r)
		dump.P("MID0--")
		return w, err
	})

	resp, err := greq.New(testBaseURL).
		Doer(testDoer).
		Middleware(mid0).
		PutDo("/get")

	assert.NoError(t, err)
	sc := resp.StatusCode
	assert.True(t, httpreq.IsOK(sc))

}

func TestHReq_Use_Multi_MiddleFunc(t *testing.T) {
	buf := &bytes.Buffer{}
	mid0 := greq.MiddleFunc(func(r *http.Request, next greq.HandleFunc) (*greq.Response, error) {
		buf.WriteString("MID0>>")
		w, err := next(r)
		buf.WriteString(">>MID0")
		return w, err
	})

	mid1 := greq.MiddleFunc(func(r *http.Request, next greq.HandleFunc) (*greq.Response, error) {
		buf.WriteString("MID1>>")
		w, err := next(r)
		buf.WriteString(">>MID1")
		return w, err
	})

	mid2 := greq.MiddleFunc(func(r *http.Request, next greq.HandleFunc) (*greq.Response, error) {
		buf.WriteString("MID2>>")
		w, err := next(r)
		buf.WriteString(">>MID2")
		return w, err
	})

	resp, err := greq.New(testBaseURL).
		Doer(httpreq.DoerFunc(func(req *http.Request) (*http.Response, error) {
			tw := httptest.NewRecorder()
			buf.WriteString("(CORE)")
			return tw.Result(), nil
		})).
		Middleware(mid0, mid1, mid2).
		PatchDo("/get")

	assert.NoError(t, err)
	assert.True(t, resp.IsOK())

	fmt.Println(buf.String())
}
