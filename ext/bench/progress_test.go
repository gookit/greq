package bench

import (
	"testing"
	"time"
)

func TestShowProgressBar(t *testing.T) {
	bench := NewHTTPBench("http://example.com")
	bench.Number = 10
	bench.Concurrency = 2

	// 模拟一些请求
	bench.totalReqs = 5
	bench.startTime = time.Now().Add(-2 * time.Second)

	// 测试进度显示 50%
	bench.showProgressBar()

	// 模拟完成状态
	bench.totalReqs = 10
	bench.showProgressBar() // 应该只显示一次完成状态
	bench.showProgressBar() // 再次调用应该不会重复显示
}
