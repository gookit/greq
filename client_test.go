package greq_test

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

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

// TestClient_Retry_Config 测试重试配置
func TestClient_Retry_Config(t *testing.T) {
	client := greq.New()

	// 测试设置最大重试次数
	client.WithMaxRetries(3)
	assert.Eq(t, 3, client.MaxRetries)

	// 测试设置重试延迟
	client.WithRetryDelay(100)
	assert.Eq(t, 100, client.RetryDelay)

	// 测试设置自定义重试检查器
	customChecker := func(resp *greq.Response, err error, attempt int) bool {
		return false
	}
	client.WithRetryChecker(customChecker)
	assert.NotNil(t, client.RetryChecker)

	// 测试设置完整配置
	client.WithMaxRetries(5).WithRetryDelay(200).WithRetryChecker(customChecker)
	assert.Eq(t, 5, client.MaxRetries)
	assert.Eq(t, 200, client.RetryDelay)
	assert.NotNil(t, client.RetryChecker)
}

// TestClient_Retry_ServerError 测试服务器错误重试
func TestClient_Retry_ServerError(t *testing.T) {
	attemptCount := 0

	// 创建测试服务器，模拟服务器错误
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		if attemptCount <= 2 {
			// 前两次返回500错误
			w.WriteHeader(500)
			w.Write([]byte("server error"))
		} else {
			// 第三次返回成功
			w.WriteHeader(200)
			w.Write([]byte("success"))
		}
	}))
	defer ts.Close()

	// 创建客户端并设置重试配置
	client := greq.New().WithMaxRetries(3).WithRetryDelay(10)

	// 发送请求
	resp, err := client.GetDo(ts.URL)

	// 验证结果
	assert.NoErr(t, err)
	assert.NotNil(t, resp)
	assert.Eq(t, 200, resp.StatusCode)
	assert.Eq(t, 3, attemptCount) // 应该重试了2次，总共3次请求
}

// TestClient_Retry_MaxAttempts 测试最大重试次数
func TestClient_Retry_MaxAttempts(t *testing.T) {
	attemptCount := 0

	// 创建测试服务器，始终返回500错误
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		w.WriteHeader(500)
		w.Write([]byte("server error"))
	}))
	defer ts.Close()

	// 创建客户端并设置重试配置
	client := greq.New().WithMaxRetries(2).WithRetryDelay(10)

	// 发送请求
	resp, err := client.GetDo(ts.URL)

	// 验证结果
	assert.NoErr(t, err) // 即使达到最大重试次数，只要连接成功就不会返回错误
	assert.NotNil(t, resp)
	assert.Eq(t, 500, resp.StatusCode)
	assert.Eq(t, 3, attemptCount) // 应该重试了2次，总共3次请求
}

// TestClient_Retry_CustomChecker 测试自定义重试检查器
func TestClient_Retry_CustomChecker(t *testing.T) {
	attemptCount := 0

	// 创建测试服务器
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		if attemptCount == 1 {
			// 第一次返回404错误
			w.WriteHeader(404)
			w.Write([]byte("not found"))
		} else {
			// 后续返回成功
			w.WriteHeader(200)
			w.Write([]byte("success"))
		}
	}))
	defer ts.Close()

	// 创建自定义重试检查器，只对404错误重试
	customChecker := func(resp *greq.Response, err error, attempt int) bool {
		if resp != nil && resp.StatusCode == 404 {
			return true
		}
		return false
	}

	// 创建客户端并设置重试配置
	client := greq.New().WithMaxRetries(2).WithRetryDelay(10).
	WithRetryChecker(customChecker)

	// 发送请求
	resp, err := client.GetDo(ts.URL)

	// 验证结果
	assert.NoErr(t, err)
	assert.NotNil(t, resp)
	assert.Eq(t, 200, resp.StatusCode)
	assert.Eq(t, 2, attemptCount) // 应该重试了1次，总共2次请求
}

// TestClient_Retry_OptionLevel 测试选项级别的重试配置
func TestClient_Retry_OptionLevel(t *testing.T) {
	attemptCount := 0

	// 创建测试服务器，模拟服务器错误
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		if attemptCount <= 2 {
			// 前两次返回500错误
			w.WriteHeader(500)
			w.Write([]byte("server error"))
		} else {
			// 第三次返回成功
			w.WriteHeader(200)
			w.Write([]byte("success"))
		}
	}))
	defer ts.Close()

	// 创建客户端（不设置重试配置）
	client := greq.New()

	// 使用选项级别的重试配置发送请求
	resp, err := client.GetDo(ts.URL, greq.WithMaxRetries(3), greq.WithRetryDelay(10))

	// 验证结果
	assert.NoErr(t, err)
	assert.NotNil(t, resp)
	assert.Eq(t, 200, resp.StatusCode)
	assert.Eq(t, 3, attemptCount) // 应该重试了2次，总共3次请求

	// 验证客户端级别的配置没有被修改
	assert.Eq(t, 0, client.MaxRetries)
	assert.Eq(t, 0, client.RetryDelay)
}

// TestClient_Retry_NoRetry 测试不需要重试的情况
func TestClient_Retry_NoRetry(t *testing.T) {
	attemptCount := 0

	// 创建测试服务器，始终返回200成功
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		w.WriteHeader(200)
		w.Write([]byte("success"))
	}))
	defer ts.Close()

	// 创建客户端并设置重试配置
	client := greq.New().WithMaxRetries(3).WithRetryDelay(10)

	// 发送请求
	resp, err := client.GetDo(ts.URL)

	// 验证结果
	assert.NoErr(t, err)
	assert.NotNil(t, resp)
	assert.Eq(t, 200, resp.StatusCode)
	assert.Eq(t, 1, attemptCount) // 不应该重试，只有1次请求
}

// TestClient_Retry_Delay 测试重试延迟
func TestClient_Retry_Delay(t *testing.T) {
	attemptCount := 0
	lastTime := time.Now()

	// 创建测试服务器
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		if attemptCount <= 2 {
			// 记录时间间隔
			if attemptCount > 1 {
				duration := time.Since(lastTime)
				assert.True(t, duration >= 50*time.Millisecond, "重试延迟应该至少50ms")
			}
			lastTime = time.Now()

			// 返回500错误
			w.WriteHeader(500)
			w.Write([]byte("server error"))
		} else {
			// 第三次返回成功
			w.WriteHeader(200)
			w.Write([]byte("success"))
		}
	}))
	defer ts.Close()

	// 创建客户端并设置重试配置（延迟50ms）
	client := greq.New().WithMaxRetries(3).WithRetryDelay(50)

	// 发送请求
	resp, err := client.GetDo(ts.URL)

	// 验证结果
	assert.NoErr(t, err)
	assert.NotNil(t, resp)
	assert.Eq(t, 200, resp.StatusCode)
	assert.Eq(t, 3, attemptCount) // 应该重试了2次，总共3次请求
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
	_, err = client.Download(testBaseURL + "/json", savePath)
	assert.NoErr(t, err)

	// 验证文件内容
	content, err := os.ReadFile(savePath)
	assert.NoErr(t, err)
	assert.StrContains(t, string(content), `"/json"`)

	// 测试下载失败（404）
	ts404 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts404.Close()

	savePath404 := filepath.Join(tempDir, "not_found.json")
	_, err = client.Download(ts404.URL, savePath404)
	assert.Err(t, err)
	assert.Contains(t, err.Error(), "Download failed, status code: 404")
}