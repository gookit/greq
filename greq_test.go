package greq_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gookit/goutil/dump"
	"github.com/gookit/goutil/fsutil"
	"github.com/gookit/goutil/netutil/httpreq"
	"github.com/gookit/goutil/testutil"
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
