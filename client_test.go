package greq_test

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"

	"github.com/gookit/goutil/dump"
	"github.com/gookit/goutil/testutil"
	"github.com/gookit/goutil/testutil/assert"
	"github.com/gookit/greq"
)

func TestClient_Doer(t *testing.T) {
	buf := &bytes.Buffer{}

	mid0 := greq.MiddleFunc(func(r *http.Request, next greq.HandleFunc) (*greq.Response, error) {
		dump.P("MID0++")
		w, err := next(r)
		dump.P("MID0--")
		return w, err
	})

	resp, err := greq.NewClient(testBaseURL).
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

func TestClient_Send(t *testing.T) {
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

func TestClient_GetDo(t *testing.T) {
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

func TestClient_PostDo(t *testing.T) {
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

func TestClient_SendRaw(t *testing.T) {
	resp, err := greq.New(testBaseURL).
		SendRaw(`POST /post HTTP/1.1
Host: example.com
Content-Type: application/json
Accept: */*

{"name": "inhere", "age": ${age}}`, map[string]string{"age": "25"})

	assert.NoErr(t, err)
	assert.True(t, resp.IsOK())
	assert.True(t, resp.IsSuccessful())

	resData := testutil.ParseRespToReply(resp.Response)
	assert.NotEmpty(t, resData.Body)
	assert.Eq(t, "application/json", resData.ContentType())
	jsonData := resData.JSON.(map[string]any)
	assert.Eq(t, "inhere", jsonData["name"])
	assert.Eq(t, float64(25), jsonData["age"])
	dump.P(resData)
}

func TestClient_String(t *testing.T) {
	str := greq.New(testBaseURL).
		UserAgent("some-cli/1.0").
		BasicAuth("inhere", "some string").
		JSONType().
		StringBody("hi, with body").
		Post("/post", `{"name": "inhere"}`).
		String()

	fmt.Println(str)
}
