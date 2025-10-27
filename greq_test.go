package greq_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
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
