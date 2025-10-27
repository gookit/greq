package greq_test

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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

func TestClient_Download(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "greq_download_test")
	assert.NoErr(t, err)
	defer os.RemoveAll(tempDir)

	// 创建客户端
	client := greq.New()

	// 测试下载成功
	savePath := filepath.Join(tempDir, "test_down.json")
	err = client.Download(testBaseURL + "/json", savePath)
	assert.NoErr(t, err)

	// 验证文件内容
	content, err := os.ReadFile(savePath)
	assert.NoErr(t, err)
	assert.Equal(t, `{"message": "test content"}`, string(content))

	// 测试下载失败（404）
	ts404 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts404.Close()

	savePath404 := filepath.Join(tempDir, "not_found.json")
	err = client.Download(ts404.URL, savePath404)
	assert.Err(t, err)
	assert.Contains(t, err.Error(), "下载失败，状态码: 404")
}