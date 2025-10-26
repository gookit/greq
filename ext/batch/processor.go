package batch

import (
	"context"
	"sync"
	"time"

	"github.com/gookit/greq"
)

// Processor handles batch request processing
type Processor struct {
	// client is the HTTP client used for requests
	client *greq.Client
	// requests is the list of requests to process
	requests []*Request
	// results is the channel for receiving results
	results chan *Result
	// maxConcurrency is the maximum number of concurrent requests
	maxConcurrency int
	// timeout is the maximum time to wait for all requests
	timeout time.Duration
	// ctx is the context for the batch operation
	ctx context.Context
	// cancel is the cancel function for the context
	cancel context.CancelFunc
}

// NewProcessor creates a new batch processor
func NewProcessor(optFns ...ProcessOptionFn) *Processor {
	ctx, cancel := context.WithCancel(context.Background())

	bp := &Processor{
		client:   greq.Std(),
		requests: make([]*Request, 0),
		results:  make(chan *Result, 100),
		ctx:      ctx,
		cancel:   cancel,
		timeout:  30 * time.Second,
		// max concurrency
		maxConcurrency: 10,
	}

	// Apply options
	for _, optFn := range optFns {
		optFn(bp)
	}
	return bp
}

// AddRequest adds a request to the batch
func (bp *Processor) AddRequest(req *Request) *Processor {
	if req != nil {
		bp.requests = append(bp.requests, req)
	}
	return bp
}

// Add adds a simple GET request to the batch
func (bp *Processor) Add(id, method, url string, body any, optFns ...greq.OptionFn) *Processor {
	return bp.AddRequest(&Request{
		ID:      id,
		Method:  method,
		URL:     url,
		Body:    body,
		Options: optFns,
	})
}

// AddGet adds a GET request to the batch
func (bp *Processor) AddGet(id, url string, optFns ...greq.OptionFn) *Processor {
	return bp.Add(id, "GET", url, nil, optFns...)
}

// AddPost adds a POST request to the batch
func (bp *Processor) AddPost(id, url string, body any, optFns ...greq.OptionFn) *Processor {
	return bp.Add(id, "POST", url, body, optFns...)
}

// AddPut adds a PUT request to the batch
func (bp *Processor) AddPut(id, url string, body any, optFns ...greq.OptionFn) *Processor {
	return bp.Add(id, "PUT", url, body, optFns...)
}

// AddDelete adds a DELETE request to the batch
func (bp *Processor) AddDelete(id, url string, optFns ...greq.OptionFn) *Processor {
	return bp.Add(id, "DELETE", url, nil, optFns...)
}

// executeRequest executes a single request
func (bp *Processor) executeRequest(req *Request) *Result {
	start := time.Now()
	result := &Result{
		ID:      req.ID,
		Request: req,
	}

	// Execute the request based on method
	var resp *greq.Response
	var err error

	// Create options with method and body
	opts := make([]greq.OptionFn, 0, len(req.Options)+1)
	opts = append(opts, greq.WithMethod(req.Method))
	if req.Body != nil {
		opts = append(opts, greq.WithBody(req.Body))
	}
	opts = append(opts, req.Options...)

	resp, err = bp.client.SendWithOpt(req.URL, greq.NewOpt2(opts, req.Method))

	result.Response = resp
	result.Error = err
	result.Duration = time.Since(start)

	return result
}

// worker processes requests from the jobs channel
func (bp *Processor) worker(wg *sync.WaitGroup, jobs <-chan *Request) {
	defer wg.Done()

	for req := range jobs {
		select {
		case <-bp.ctx.Done():
			return
		default:
			result := bp.executeRequest(req)
			select {
			case bp.results <- result:
			case <-bp.ctx.Done():
				return
			}
		}
	}
}

// ExecuteAll executes all requests and waits for all to complete
func (bp *Processor) ExecuteAll() Results {
	if len(bp.requests) == 0 {
		return nil
	}

	// Create timeout context if timeout is set
	ctx := bp.ctx
	if bp.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(bp.ctx, bp.timeout)
		defer cancel()
	}

	// Create jobs channel
	jobs := make(chan *Request, len(bp.requests))

	// Start workers
	var wg sync.WaitGroup
	numWorkers := bp.maxConcurrency
	if numWorkers > len(bp.requests) {
		numWorkers = len(bp.requests)
	}

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go bp.worker(&wg, jobs)
	}

	// Send all requests to jobs channel
	go func() {
		defer close(jobs)
		for _, req := range bp.requests {
			select {
			case jobs <- req:
			case <-ctx.Done():
				return
			}
		}
	}()

	// Collect results
	results := make(Results, len(bp.requests))
	completed := 0
	total := len(bp.requests)

	for completed < total {
		select {
		case result := <-bp.results:
			results[result.ID] = result
			completed++
		case <-ctx.Done():
			// Context cancelled, return what we have so far
			return results
		}
	}

	// Wait for all workers to finish
	wg.Wait()

	return results
}

// ExecuteAny executes requests and returns when any one completes successfully
func (bp *Processor) ExecuteAny() *Result {
	if len(bp.requests) == 0 {
		return nil
	}

	// Create timeout context if timeout is set
	ctx := bp.ctx
	if bp.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(bp.ctx, bp.timeout)
		defer cancel()
	}

	// Create jobs channel
	jobs := make(chan *Request, len(bp.requests))

	// Start workers
	var wg sync.WaitGroup
	numWorkers := bp.maxConcurrency
	if numWorkers > len(bp.requests) {
		numWorkers = len(bp.requests)
	}

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go bp.worker(&wg, jobs)
	}

	// Send all requests to jobs channel
	go func() {
		defer close(jobs)
		for _, req := range bp.requests {
			select {
			case jobs <- req:
			case <-ctx.Done():
				return
			}
		}
	}()

	// Wait for any successful result
	for {
		select {
		case result := <-bp.results:
			if result.Error == nil && result.Response != nil && result.Response.IsOK() {
				bp.cancel() // Cancel other requests
				wg.Wait()
				return result
			}
			// If this result failed, continue waiting
		case <-ctx.Done():
			// Context cancelled, return last result or nil
			select {
			case result := <-bp.results:
				wg.Wait()
				return result
			default:
				wg.Wait()
				return nil
			}
		}
	}
}

// Close cleans up resources
func (bp *Processor) Close() {
	bp.cancel()
	close(bp.results)
}

// MaxConcurrency returns the maximum number of concurrent requests
func (bp *Processor) MaxConcurrency() int {
	return bp.maxConcurrency
}

// Timeout returns the timeout for the batch operation
func (bp *Processor) Timeout() time.Duration {
	return bp.timeout
}

// Client returns the HTTP client used for requests
func (bp *Processor) Client() *greq.Client {
	return bp.client
}

// Context returns the context for the batch operation
func (bp *Processor) Context() context.Context {
	return bp.ctx
}
