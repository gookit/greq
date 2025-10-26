package batch_test

import (
	"context"
	"testing"
	"time"

	"github.com/gookit/greq"
	"github.com/gookit/greq/ext/batch"
	"github.com/gookit/goutil/testutil/assert"
)

func TestWithMaxConcurrency(t *testing.T) {
	// 测试正常情况
	optFn := batch.WithMaxConcurrency(5)
	bp := batch.NewProcessor()
	optFn(bp)
	assert.Eq(t, 5, bp.MaxConcurrency())

	// 测试边界情况：0值
	optFn = batch.WithMaxConcurrency(0)
	bp = batch.NewProcessor()
	optFn(bp)
	assert.Eq(t, 10, bp.MaxConcurrency()) // 默认值

	// 测试负数情况
	optFn = batch.WithMaxConcurrency(-1)
	bp = batch.NewProcessor()
	optFn(bp)
	assert.Eq(t, 10, bp.MaxConcurrency()) // 默认值
}

func TestWithBatchTimeout(t *testing.T) {
	// 测试正常情况
	timeout := 5 * time.Second
	optFn := batch.WithBatchTimeout(timeout)
	bp := batch.NewProcessor()
	optFn(bp)
	assert.Eq(t, timeout, bp.Timeout())
}

func TestWithClient(t *testing.T) {
	// 测试正常情况
	client := greq.New()
	optFn := batch.WithClient(client)
	bp := batch.NewProcessor()
	optFn(bp)
	assert.Eq(t, client, bp.Client())
}

func TestWithContext(t *testing.T) {
	// 测试正常情况
	ctx := context.Background()
	optFn := batch.WithContext(ctx)
	bp := batch.NewProcessor()
	optFn(bp)
	assert.Eq(t, ctx, bp.Context())
}