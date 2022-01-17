package hreq_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gookit/goutil/dump"
	"github.com/gookit/goutil/fsutil"
	"github.com/gookit/goutil/netutil/httpreq"
	"github.com/gookit/hreq"
	"github.com/stretchr/testify/assert"
)

var final = HandleFunc(func(req *http.Request) (*http.Response, error) {
	tw := httptest.NewRecorder()
	dump.P("CORE++")

	_, err := tw.WriteString(req.RequestURI + " > ")
	if err != nil {
		return nil, err
	}

	_, err = tw.Write(fsutil.MustReadReader(req.Body))
	dump.P("CORE--")

	return tw.Result(), err
})

type MiddleFunc func(r *http.Request, next HandleFunc) (*http.Response, error)

type HandleFunc func(r *http.Request) (*http.Response, error)

func TestBuild_Middleware(t *testing.T) {
	var middles []MiddleFunc

	mid2 := func(r *http.Request, next HandleFunc) (*http.Response, error) {
		dump.P("MID2++")
		w, err := next(r)
		dump.P("MID2--")
		return w, err
	}

	mid3 := func(msg string) MiddleFunc {
		return func(r *http.Request, next HandleFunc) (*http.Response, error) {
			dump.P("MID3++")
			w, err := next(r)
			dump.P("MID3--")
			return w, err
		}
	}

	middles = append(middles, mid2, mid3("test"))

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	do := final
	for _, m := range middles {
		do = func(r *http.Request) (*http.Response, error) {
			return m(r, do)
		}
	}

	resp, err := do(req)
	dump.P(err)
	dump.P(resp)
}

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

	resp, err := hreq.New(testBaseURL).
		Doer(testDoer).
		Use(&mid1{}).
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
		Use(mid0).
		Get("/get")

	assert.NoError(t, err)
	sc := resp.StatusCode
	assert.True(t, httpreq.IsOK(sc))

}
