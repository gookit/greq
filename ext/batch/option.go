package batch

import (
	"context"
	"time"

	"github.com/gookit/greq"
)

// ProcessOptionFn is a function to configure batch processor
type ProcessOptionFn func(bp *Processor)

// WithMaxConcurrency sets the maximum number of concurrent requests
func WithMaxConcurrency(n int) ProcessOptionFn {
	return func(bp *Processor) {
		if n > 0 {
			bp.maxConcurrency = n
		}
	}
}

// WithBatchTimeout sets the timeout for the batch operation
func WithBatchTimeout(timeout time.Duration) ProcessOptionFn {
	return func(bp *Processor) {
		bp.timeout = timeout
	}
}

// WithClient sets a custom client for batch processing
func WithClient(client *greq.Client) ProcessOptionFn {
	return func(bp *Processor) {
		bp.client = client
	}
}

// WithContext sets a custom context for batch processing
func WithContext(ctx context.Context) ProcessOptionFn {
	return func(bp *Processor) {
		bp.ctx = ctx
	}
}
