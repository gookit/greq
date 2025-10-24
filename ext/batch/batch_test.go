package batch_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/gookit/goutil/testutil"
	"github.com/gookit/goutil/testutil/assert"
	"github.com/gookit/greq"
	"github.com/gookit/greq/ext/batch"
)

var testApiURL string

func TestMain(m *testing.M) {
	// create mock server
	s := testutil.NewEchoServer()
	defer s.Close()
	testApiURL = s.HTTPHost()
	s.PrintHttpHost()

	// do testing
	m.Run()
}

func Example() {
	// Execute all requests
	bp := batch.NewProcessor(
		batch.WithMaxConcurrency(5),
		batch.WithBatchTimeout(10*time.Second),
	)

	bp.AddGet("req1", testApiURL+"/list")
	bp.AddPost("req2", testApiURL+"/submit", map[string]string{"key": "value"})

	results := bp.ExecuteAll()
	fmt.Println("Results: ", len(results))

	// Execute any (first success)
	bp2 := batch.NewProcessor()
	bp2.AddGet("mirror1", testApiURL+"/file1")
	bp2.AddGet("mirror2", testApiURL+"/file2")
	bp2.AddGet("mirror3", testApiURL+"/file3")

	result := bp2.ExecuteAny()
	fmt.Println("First successful result: ", result.ID)

	// Convenience functions
	urls := []string{testApiURL + "/api1", testApiURL + "/api2", testApiURL + "/api3"}
	allResults := batch.GetAll(urls)
	fmt.Println("All results: ", len(allResults))

	firstResult := batch.GetAny(urls)
	fmt.Println("First successful result: ", firstResult.ID)
}

func TestExecuteAll(t *testing.T) {
	// 准备测试数据
	requests := []*batch.Request{
		{
			ID:     "req1",
			Method: "GET",
			URL:    testApiURL + "/get",
		},
		{
			ID:     "req2",
			Method: "POST",
			URL:    testApiURL + "/post",
			Body:   map[string]string{"key": "value"},
		},
	}

	// 执行测试
	results := batch.ExecuteAll(requests)

	// 验证结果
	assert.Eq(t, 2, len(results))
	// 检查两个请求都成功完成，但不依赖于顺序
	var req1Result, req2Result *batch.Result
	for _, result := range results {
		if result.ID == "req1" {
			req1Result = result
		} else if result.ID == "req2" {
			req2Result = result
		}
	}

	assert.NotNil(t, req1Result)
	assert.NotNil(t, req2Result)
	//goland:noinspection ALL
	assert.Nil(t, req1Result.Error)
	//goland:noinspection ALL
	assert.Nil(t, req2Result.Error)
	assert.NotNil(t, req1Result.Response)
	assert.NotNil(t, req2Result.Response)
}

func TestExecuteAny(t *testing.T) {
	// 准备测试数据
	requests := []*batch.Request{
		{
			ID:     "req1",
			Method: "GET",
			URL:    testApiURL + "/get",
		},
		{
			ID:     "req2",
			Method: "GET",
			URL:    testApiURL + "/get",
		},
	}

	// 执行测试
	result := batch.ExecuteAny(requests)

	// 验证结果
	assert.NotNil(t, result)
	assert.Contains(t, []string{"req1", "req2"}, result.ID)
	assert.Nil(t, result.Error)
	assert.NotNil(t, result.Response)
}

func TestGetAll(t *testing.T) {
	// 准备测试数据
	urls := []string{
		testApiURL + "/get1",
		testApiURL + "/get2",
		testApiURL + "/get3",
	}

	// 执行测试
	results := batch.GetAll(urls)

	// 验证结果
	assert.Eq(t, 3, len(results))
	for _, result := range results {
		assert.Nil(t, result.Error)
		assert.NotNil(t, result.Response)
		assert.Eq(t, "GET", result.Request.Method)
	}
}

func TestGetAny(t *testing.T) {
	// 准备测试数据
	urls := []string{
		testApiURL + "/get1",
		testApiURL + "/get2",
	}

	// 执行测试
	result := batch.GetAny(urls)

	// 验证结果
	assert.NotNil(t, result)
	assert.Nil(t, result.Error)
	assert.NotNil(t, result.Response)
	assert.Eq(t, "GET", result.Request.Method)
}

func TestPostAll(t *testing.T) {
	// 准备测试数据
	urls := []string{
		testApiURL + "/post1",
		testApiURL + "/post2",
	}
	bodies := []any{
		map[string]string{"key1": "value1"},
		map[string]string{"key2": "value2"},
	}

	// 执行测试
	results := batch.PostAll(urls, bodies)

	// 验证结果
	assert.Eq(t, 2, len(results))
	for _, result := range results {
		assert.Nil(t, result.Error)
		assert.NotNil(t, result.Response)
		assert.Eq(t, "POST", result.Request.Method)
	}
}

func TestPostAll_PanicOnMismatchedLengths(t *testing.T) {
	// 准备测试数据
	urls := []string{
		testApiURL + "/post1",
	}
	bodies := []any{
		map[string]string{"key1": "value1"},
		map[string]string{"key2": "value2"}, // 多一个body
	}

	// 验证会panic
	assert.Panics(t, func() {
		batch.PostAll(urls, bodies)
	})
}

func TestExecuteAllWithEmptyRequests(t *testing.T) {
	// 执行测试
	results := batch.ExecuteAll([]*batch.Request{})

	// 验证结果
	assert.Eq(t, 0, len(results))
}

func TestExecuteAnyWithEmptyRequests(t *testing.T) {
	// 执行测试
	result := batch.ExecuteAny([]*batch.Request{})

	// 验证结果
	assert.Nil(t, result)
}

func TestGetAllWithOptions(t *testing.T) {
	// 准备测试数据
	urls := []string{
		testApiURL + "/get",
	}

	// 使用选项执行测试
	results := batch.GetAll(urls, greq.WithHeader("X-Test", "value"))

	// 验证结果
	assert.Eq(t, 1, len(results))
	assert.Nil(t, results["id_0"].Error)
	assert.NotNil(t, results["id_0"].Response)

	respData := testutil.ParseRespToReply(results["id_0"].Response.Response)
	assert.Eq(t, "GET", respData.Method)
	assert.StrContains(t, respData.URL, "/get")
	assert.StrContains(t, respData.Headers["X-Test"].(string), "value")
}

func TestGetAnyWithOptions(t *testing.T) {
	// 准备测试数据
	urls := []string{
		testApiURL + "/get",
	}

	// 使用选项执行测试
	result := batch.GetAny(urls, greq.WithHeader("X-Test", "value"))

	// 验证结果
	assert.NotNil(t, result)
	assert.Nil(t, result.Error)
	assert.NotNil(t, result.Response)
}

func TestExecuteAllWithError(t *testing.T) {
	// 准备测试数据，包含一个无效URL
	requests := []*batch.Request{
		{
			ID:     "valid",
			Method: "GET",
			URL:    testApiURL + "/get",
		},
		{
			ID:     "invalid",
			Method: "GET",
			URL:    "http://invalid.invalid", // 无效的域名
		},
	}

	// 执行测试
	results := batch.ExecuteAll(requests)

	// 验证结果
	assert.Eq(t, 2, len(results))

	// 验证有效请求成功
	assert.Eq(t, "valid", results["valid"].ID)
	assert.Nil(t, results["valid"].Error)
	assert.NotNil(t, results["valid"].Response)

	// 验证无效请求失败
	assert.Eq(t, "invalid", results["invalid"].ID)
	assert.NotNil(t, results["invalid"].Error)
}
