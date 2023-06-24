package greq_test

import (
	"bytes"
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
	// create server
	s := testutil.NewEchoServer()
	defer s.Close()
	testBaseURL = "http://" + s.Listener.Addr().String()
	fmt.Println("Test server listen on:", testBaseURL)

	// with base url
	greq.BaseURL(testBaseURL)

	// do something
	m.Run()
}

func TestHReq_Doer(t *testing.T) {
	buf := &bytes.Buffer{}

	mid0 := greq.MiddleFunc(func(r *http.Request, next greq.HandleFunc) (*greq.Response, error) {
		dump.P("MID0++")
		w, err := next(r)
		dump.P("MID0--")
		return w, err
	})

	resp, err := greq.New(testBaseURL).
		Doer(testDoer).
		Use(mid0).
		UserAgent("custom-cli/1.0").
		Get("/get").
		Do()

	assert.NoErr(t, err)
	assert.True(t, resp.IsOK())

	err = resp.Write(buf)
	assert.NoErr(t, err)
	dump.P(buf.String())
}

func TestHReq_Send(t *testing.T) {
	resp, err := greq.New(testBaseURL).
		UserAgent("custom-cli/1.0").
		Send("GET", "/get")

	assert.NoErr(t, err)
	assert.True(t, resp.IsOK())
	assert.False(t, resp.IsFail())

	retMp := make(map[string]any)
	err = resp.Decode(&retMp)
	assert.NoErr(t, err)
	dump.P(retMp)

	assert.Contains(t, retMp, "headers")

	headers := retMp["headers"].(map[string]any)
	assert.Contains(t, headers, "User-Agent")
	assert.Eq(t, "custom-cli/1.0", headers["User-Agent"])
}

func TestHReq_GetDo(t *testing.T) {
	resp, err := greq.New(testBaseURL).
		JSONType().
		GetDo("/get")

	assert.NoErr(t, err)
	assert.True(t, resp.IsOK())
	assert.True(t, resp.IsSuccessful())

	retMp := make(map[string]any)
	err = resp.Decode(&retMp)
	assert.NoErr(t, err)
	dump.P(retMp)
}

func TestHReq_PostDo(t *testing.T) {
	resp, err := greq.New(testBaseURL).
		UserAgent(greq.AgentCURL).
		JSONType().
		PostDo("/post", `{"name": "inhere"}`)

	assert.NoErr(t, err)
	assert.True(t, resp.IsOK())
	assert.True(t, resp.IsSuccessful())

	retMp := make(map[string]any)
	err = resp.Decode(&retMp)
	assert.NoErr(t, err)
	dump.P(retMp)
}

func TestHReq_String(t *testing.T) {
	str := greq.New(testBaseURL).
		UserAgent("some-cli/1.0").
		BasicAuth("inhere", "some string").
		JSONType().
		StringBody("hi, with body").
		Post("/post", `{"name": "inhere"}`).
		String()

	fmt.Println(str)
}
