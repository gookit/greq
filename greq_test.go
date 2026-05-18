package greq_test

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gookit/goutil/dump"
	"github.com/gookit/goutil/fsutil"
	"github.com/gookit/goutil/netutil/httpreq"
	"github.com/gookit/goutil/testutil"
	"github.com/gookit/goutil/testutil/assert"
	"github.com/gookit/greq"
)

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

var testBaseURL string

func TestMain(m *testing.M) {
	// create mock server
	s := testutil.NewEchoServer()
	defer s.Close()
	testBaseURL = s.HTTPHost()
	s.PrintHttpHost()

	// with base url
	greq.BaseURL(testBaseURL)

	// do something
	m.Run()
}

func TestMust(t *testing.T) {
	// Test that Must() panics when an error is returned
	assert.Panics(t, func() {
		greq.Must(greq.GetDo("https://invalid-url"))
	})

	assert.NotPanics(t, func() {
		greq.Must(greq.GetDo(testBaseURL + "/get"))
	})
}


// DefaultRetryChecker 测试默认重试检查器
func TestDefaultRetryChecker(t *testing.T) {
	// 测试网络错误
	assert.True(t, greq.DefaultRetryChecker(nil, fmt.Errorf("network error"), 0))

	// 测试5xx服务器错误
	resp500 := &greq.Response{Response: &http.Response{StatusCode: 500}}
	assert.True(t, greq.DefaultRetryChecker(resp500, nil, 0))

	// 测试429限流错误
	resp429 := &greq.Response{Response: &http.Response{StatusCode: 429}}
	assert.True(t, greq.DefaultRetryChecker(resp429, nil, 0))

	// 测试2xx成功响应
	resp200 := &greq.Response{Response: &http.Response{StatusCode: 200}}
	assert.False(t, greq.DefaultRetryChecker(resp200, nil, 0))

	// 测试4xx客户端错误
	resp404 := &greq.Response{Response: &http.Response{StatusCode: 404}}
	assert.False(t, greq.DefaultRetryChecker(resp404, nil, 0))
}

//
// Middleware tests
//

type mid1 struct{}

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

	assert.NoErr(t, err)
	assert.True(t, resp.IsOK())

	err = resp.Write(buf)
	assert.NoErr(t, err)

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

	assert.NoErr(t, err)
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

	assert.NoErr(t, err)
	assert.True(t, resp.IsOK())

	fmt.Println(buf.String())
}

//
// Response decoder tests (jsonDecoder is exercised via integration tests in client_test.go)
//

func TestXmlDecoder_Decode(t *testing.T) {
	type Person struct {
		Name string `xml:"name"`
		Age  int    `xml:"age"`
		City string `xml:"city"`
	}

	xmlData, _ := xml.Marshal(Person{Name: "inhere", Age: 30, City: "Beijing"})
	resp := &http.Response{Body: io.NopCloser(bytes.NewReader(xmlData))}

	var got Person
	err := (greq.XmlDecoder{}).Decode(resp, &got)
	assert.NoErr(t, err)
	assert.Eq(t, "inhere", got.Name)
	assert.Eq(t, 30, got.Age)
	assert.Eq(t, "Beijing", got.City)
}

func TestXmlDecoder_Decode_Error(t *testing.T) {
	resp := &http.Response{Body: io.NopCloser(strings.NewReader("<invalid xml>"))}
	var got map[string]any
	err := (greq.XmlDecoder{}).Decode(resp, &got)
	assert.Err(t, err)
}
