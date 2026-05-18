package bench

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gookit/goutil/testutil/assert"
)

func TestSnapshotShape(t *testing.T) {
	b := NewHTTPBench("http://example.com")
	b.Number = 10
	b.startTime = time.Now().Add(-2 * time.Second)
	atomic.StoreInt64(&b.totalReqs, 5)

	s := b.snapshot(false)
	assert.Eq(t, int64(5), s.Completed)
	assert.Eq(t, int64(10), s.Total)
	assert.False(t, s.Done)
	assert.True(t, s.Elapsed >= 2*time.Second)

	done := b.snapshot(true)
	assert.True(t, done.Done)
}

// TestOnProgressFires verifies the OnProgress callback is invoked
// during a tiny self-cancelling run. Uses a fake server.
func TestOnProgressFires(t *testing.T) {
	b := NewHTTPBench("http://127.0.0.1:1/") // unreachable — every request fails fast
	b.Number = 1
	b.Concurrency = 1
	b.Timeout = 100 * time.Millisecond

	var calls int32
	b.OnProgress(func(s Snapshot) {
		atomic.AddInt32(&calls, 1)
	}).ProgressTick(20 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := b.RunCtx(ctx)
	assert.NoErr(t, err)

	// At least one (final) snapshot must have fired.
	assert.True(t, atomic.LoadInt32(&calls) >= 1)
}
