package bench

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gookit/goutil/strutil"
	"github.com/gookit/greq"
)

// Snapshot is a point-in-time view of bench progress, delivered to OnProgress.
// All fields are populated from atomically-loaded counters so the callback
// may run in a separate goroutine.
type Snapshot struct {
	Completed int64         // requests dispatched so far (== totalReqs)
	Total     int64         // target Number (0 if Duration-driven)
	Elapsed   time.Duration // time since the run started
	Duration  time.Duration // configured run duration (0 if Number-driven)
	Done      bool          // true on the final snapshot fired after wg.Wait
}

// ProgressFn is the callback invoked on each tick and at completion.
// Implementations should be non-blocking — it runs in the bench's progress goroutine.
type ProgressFn func(s Snapshot)

// HTTPBench 实现类似 apache bench 的测试功能
type HTTPBench struct {
	// 基础配置
	URL      string
	Method   string
	Headers  map[string]string
	Body     []byte
	Timeout  time.Duration // 每个请求的超时时间
	QPSLimit int           // QPS限制 (0表示不限制)

	// 测试参数
	Number      int           // 总请求数
	Concurrency int           // 并发数
	Duration    time.Duration // 测试时长 (如果指定，则忽略Number)

	// 统计信息
	startTime  time.Time
	endTime    time.Time
	totalReqs  int64
	totalBytes int64

	// 结果统计
	successReqs int64
	failReqs    int64
	statusCodes map[int]int64
	// 响应时间统计
	respTimes []time.Duration

	// Progress 回调 — nil 时不启动进度协程。CLI 渲染通过这个钩子接入，
	// 让库本身不依赖任何 UI 包。
	onProgress     ProgressFn
	progressTick   time.Duration // 触发间隔, 0 时默认 200ms

	// 客户端
	client *greq.Client

	// 上下文控制
	ctx    context.Context
	mu     sync.Mutex
	cancel context.CancelFunc

	// QPS限制
	rateLimiter *time.Ticker
}

// BenchResult 保存测试结果
type BenchResult struct {
	URL            string
	TotalReqs      int64
	SuccessReqs    int64
	FailReqs       int64
	Duration       time.Duration
	AvgRespTime    time.Duration
	MinRespTime    time.Duration
	MaxRespTime    time.Duration
	ReqsPerSecond  float64
	BytesPerSecond float64
	StatusCodes    map[int]int64
}

// NewHTTPBench 创建新的HTTPBench实例
func NewHTTPBench(url string) *HTTPBench {
	return &HTTPBench{
		URL:         url,
		Method:      "GET",
		Headers:     make(map[string]string),
		Concurrency: 1,
		Number:      100,
		statusCodes: make(map[int]int64),
		respTimes:   make([]time.Duration, 0, 1000),
	}
}

// SetMethod 设置HTTP方法
func (b *HTTPBench) SetMethod(method string) *HTTPBench {
	b.Method = strutil.Uppercase(method)
	return b
}

// SetHeaders 设置请求头
func (b *HTTPBench) SetHeaders(headers map[string]string) *HTTPBench {
	b.Headers = headers
	return b
}

// SetBody 设置请求体
func (b *HTTPBench) SetBody(body []byte) *HTTPBench {
	b.Body = body
	return b
}

// SetTimeout 设置超时时间
func (b *HTTPBench) SetTimeout(timeout time.Duration) *HTTPBench {
	b.Timeout = timeout
	return b
}

// SetQPSLimit 设置QPS限制
func (b *HTTPBench) SetQPSLimit(qps int) *HTTPBench {
	b.QPSLimit = qps
	return b
}

// SetConcurrency 设置并发数
func (b *HTTPBench) SetConcurrency(concurrency int) *HTTPBench {
	b.Concurrency = concurrency
	return b
}

// SetNumber 设置请求数
func (b *HTTPBench) SetNumber(number int) *HTTPBench {
	b.Number = number
	return b
}

// SetDuration 设置测试时长
func (b *HTTPBench) SetDuration(duration time.Duration) *HTTPBench {
	b.Duration = duration
	return b
}

// OnProgress installs a callback invoked periodically (and once at completion)
// while RunCtx is executing. Pass nil to disable progress reporting (default).
//
// The callback runs in the bench's progress goroutine — keep it non-blocking.
// The library does NOT render anything itself; callers (such as cmd/gbench)
// drive their own UI (e.g. cliui/progress) via this hook.
func (b *HTTPBench) OnProgress(fn ProgressFn) *HTTPBench {
	b.onProgress = fn
	return b
}

// ProgressTick controls how often OnProgress fires. Default is 200ms.
func (b *HTTPBench) ProgressTick(d time.Duration) *HTTPBench {
	b.progressTick = d
	return b
}

// snapshot builds a point-in-time Snapshot from the current atomic counters.
func (b *HTTPBench) snapshot(done bool) Snapshot {
	return Snapshot{
		Completed: atomic.LoadInt64(&b.totalReqs),
		Total:     int64(b.Number),
		Elapsed:   time.Since(b.startTime),
		Duration:  b.Duration,
		Done:      done,
	}
}

// initClient 初始化HTTP客户端
func (b *HTTPBench) initClient() {
	b.client = greq.New()

	if b.Timeout > 0 {
		b.client.ConfigHClient(func(hc *http.Client) {
			hc.Timeout = b.Timeout
		})
	}

	// 设置默认头
	for k, v := range b.Headers {
		b.client.DefaultHeader(k, v)
	}
}

// Run executes the benchmark with a background context. Equivalent to RunCtx(context.Background()).
func (b *HTTPBench) Run() (*BenchResult, error) {
	return b.RunCtx(context.Background())
}

// RunCtx executes the benchmark with the given parent context. Cancel the context
// to stop the test gracefully — workers finish in-flight requests then exit, and
// partial results are returned.
func (b *HTTPBench) RunCtx(parent context.Context) (*BenchResult, error) {
	b.initClient()
	b.startTime = time.Now()

	b.ctx, b.cancel = context.WithCancel(parent)
	defer b.cancel()

	// If a test duration is set, layer a timeout on top of the parent cancel.
	if b.Duration > 0 {
		b.ctx, b.cancel = context.WithTimeout(b.ctx, b.Duration)
	}

	if b.QPSLimit > 0 {
		interval := time.Second / time.Duration(b.QPSLimit)
		b.rateLimiter = time.NewTicker(interval)
		defer b.rateLimiter.Stop()
	}

	var wg sync.WaitGroup
	workCh := make(chan struct{}, b.Concurrency)

	// Progress goroutine — only started when an OnProgress callback is installed.
	// progressDone lets us deterministically wait for it to exit before the final
	// snapshot so the caller's renderer can't observe the bench mid-shutdown.
	var progressDone chan struct{}
	if b.onProgress != nil {
		tick := b.progressTick
		if tick <= 0 {
			tick = 200 * time.Millisecond
		}
		ticker := time.NewTicker(tick)
		defer ticker.Stop()
		progressDone = make(chan struct{})
		go func() {
			defer close(progressDone)
			for {
				select {
				case <-ticker.C:
					b.onProgress(b.snapshot(false))
				case <-b.ctx.Done():
					return
				}
			}
		}()
	}

	for i := 0; i < b.Concurrency; i++ {
		wg.Add(1)
		go b.worker(workCh, &wg)
	}
	go b.dispatcher(workCh)

	wg.Wait()
	b.endTime = time.Now()

	// Deliver one final snapshot (may be partial if context was cancelled).
	if b.onProgress != nil {
		b.cancel()      // stop the progress goroutine
		<-progressDone  // wait for it to actually exit
		b.onProgress(b.snapshot(true))
	}

	return b.generateResult(), nil
}

// dispatcher 任务分发器
func (b *HTTPBench) dispatcher(workCh chan<- struct{}) {
	defer close(workCh)

	reqCount := 0
	for {
		select {
		case <-b.ctx.Done():
			return
		default:
			if b.Duration == 0 && reqCount >= b.Number {
				return
			}

			// QPS 限制 — 同时监听 ctx.Done 以便低 QPS 下取消能及时生效
			if b.rateLimiter != nil {
				select {
				case <-b.rateLimiter.C:
				case <-b.ctx.Done():
					return
				}
			}

			select {
			case workCh <- struct{}{}:
				reqCount++
				atomic.AddInt64(&b.totalReqs, 1)
			case <-b.ctx.Done():
				return
			}
		}
	}
}

// worker 工作协程
func (b *HTTPBench) worker(workCh <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	for range workCh {
		b.doRequest()
	}
}

// doRequest 执行单个请求，统计直接写入 b 的计数器
func (b *HTTPBench) doRequest() {
	start := time.Now()

	var err error
	var resp *greq.Response

	if len(b.Body) > 0 {
		resp, err = b.client.Do(b.Method, b.URL, func(opt *greq.Options) {
			opt.Body = b.Body
		})
	} else {
		resp, err = b.client.Do(b.Method, b.URL)
	}

	duration := time.Since(start)

	if err != nil {
		atomic.AddInt64(&b.failReqs, 1)
		// resp may be non-nil even with err (e.g. failed redirect) — close it to avoid FD leak under high -n.
		if resp != nil {
			resp.QuietCloseBody()
		}
		return
	}

	// 统计状态码
	b.mu.Lock()
	b.statusCodes[resp.StatusCode]++
	b.respTimes = append(b.respTimes, duration)
	b.mu.Unlock()

	// 流式读 body 只为计数 — 不要把整个响应拉到内存里 (大响应 + 高并发 = OOM).
	// io.Discard + io.Copy 让 transport 复用连接并把字节扔掉。
	if resp.Body != nil {
		if n, cerr := io.Copy(io.Discard, resp.Body); cerr == nil {
			atomic.AddInt64(&b.totalBytes, n)
		}
	}

	if resp.IsOK() {
		atomic.AddInt64(&b.successReqs, 1)
	} else {
		atomic.AddInt64(&b.failReqs, 1)
	}

	resp.CloseBody()
}

// generateResult 生成最终结果
func (b *HTTPBench) generateResult() *BenchResult {
	duration := b.endTime.Sub(b.startTime)

	b.mu.Lock()
	defer b.mu.Unlock()

	result := &BenchResult{
		URL:         b.URL,
		TotalReqs:   b.totalReqs,
		SuccessReqs: b.successReqs,
		FailReqs:    b.failReqs,
		Duration:    duration,
		StatusCodes: make(map[int]int64),
	}
	// Guard against divide-by-zero when test was cancelled before any request completed.
	if secs := duration.Seconds(); secs > 0 {
		result.ReqsPerSecond = float64(b.totalReqs) / secs
		result.BytesPerSecond = float64(b.totalBytes) / secs
	}

	// 复制状态码统计
	for k, v := range b.statusCodes {
		result.StatusCodes[k] = v
	}

	// 计算响应时间统计
	if len(b.respTimes) > 0 {
		var totalTime time.Duration
		result.MinRespTime = b.respTimes[0]
		result.MaxRespTime = b.respTimes[0]

		for _, t := range b.respTimes {
			totalTime += t
			if t < result.MinRespTime {
				result.MinRespTime = t
			}
			if t > result.MaxRespTime {
				result.MaxRespTime = t
			}
		}

		result.AvgRespTime = totalTime / time.Duration(len(b.respTimes))
	}

	return result
}

// String returns the formatted result with ccolor tags (<green>...</> etc.).
// Suitable for printing via ccolor.Print on a terminal. For file output use
// PlainString to avoid raw tags ending up in the file.
func (r *BenchResult) String() string { return r.format(true) }

// PlainString returns the formatted result without color tags. Use this when
// writing the result to a file or any non-ccolor sink.
func (r *BenchResult) PlainString() string { return r.format(false) }

func (r *BenchResult) format(colored bool) string {
	// 一对小函数避免在每行都写 if colored 分支
	g := func(s string) string {
		if colored {
			return "<green>" + s + "</>"
		}
		return s
	}
	red := func(s string) string {
		if colored {
			return "<red>" + s + "</>"
		}
		return s
	}

	buf := make([]byte, 0, 256)

	var successRatio float64
	if r.TotalReqs > 0 {
		successRatio = float64(r.SuccessReqs) / float64(r.TotalReqs)
	}

	buf = append(buf, fmt.Sprintf("Total      requests: %d\n", r.TotalReqs)...)
	buf = append(buf, fmt.Sprintf("Successful requests: %s(%.2f%%)\n", g(fmt.Sprintf("%d", r.SuccessReqs)), successRatio*100)...)
	buf = append(buf, fmt.Sprintf("Failed     requests: %s\n", red(fmt.Sprintf("%d", r.FailReqs)))...)
	buf = append(buf, fmt.Sprintf("Duration       time: %s\n", r.Duration)...)
	buf = append(buf, fmt.Sprintf("Requests per second: %.2f\n", r.ReqsPerSecond)...)
	buf = append(buf, fmt.Sprintf("Bytes   per  second: %.2f\n", r.BytesPerSecond)...)

	if r.AvgRespTime > 0 {
		buf = append(buf, fmt.Sprintf("Average response time: %s\n", r.AvgRespTime)...)
		buf = append(buf, fmt.Sprintf("Minimum response time: %s\n", r.MinRespTime)...)
		buf = append(buf, fmt.Sprintf("Maximum response time: %s\n", r.MaxRespTime)...)
	}

	if len(r.StatusCodes) > 0 {
		buf = append(buf, "\nStatus code distribution:\n"...)
		for code, count := range r.StatusCodes {
			buf = append(buf, fmt.Sprintf("  %s: %d\n", g(fmt.Sprintf("%d", code)), count)...)
		}
	}

	return string(buf)
}
