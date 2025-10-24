package ext

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gookit/goutil/strutil"
	"github.com/gookit/greq"
)

// HTTPBench 实现类似 apache bench 的测试功能
type HTTPBench struct {
	// 基础配置
	URL      string
	Method   string
	Headers  map[string]string
	Body     []byte
	Timeout  time.Duration
	QPSLimit int

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

// Run 执行基准测试
func (b *HTTPBench) Run() (*BenchResult, error) {
	b.initClient()
	b.startTime = time.Now()

	// 设置上下文
	b.ctx, b.cancel = context.WithCancel(context.Background())
	defer b.cancel()

	// 如果设置了测试时长，使用定时器
	if b.Duration > 0 {
		b.ctx, b.cancel = context.WithTimeout(b.ctx, b.Duration)
	}

	// 设置QPS限制
	if b.QPSLimit > 0 {
		interval := time.Second / time.Duration(b.QPSLimit)
		b.rateLimiter = time.NewTicker(interval)
		defer b.rateLimiter.Stop()
	}

	// 启动工作协程
	var wg sync.WaitGroup
	workCh := make(chan struct{}, b.Concurrency)

	// 启动统计协程
	resultCh := make(chan *benchReqResult, b.Concurrency*10)
	go b.collectResults(resultCh)

	// 启动工作协程
	for i := 0; i < b.Concurrency; i++ {
		wg.Add(1)
		go b.worker(workCh, resultCh, &wg)
	}

	// 发送任务
	go b.dispatcher(workCh)

	// 等待所有工作完成
	wg.Wait()
	close(resultCh)

	// 等待统计协程完成
	time.Sleep(100 * time.Millisecond)

	b.endTime = time.Now()

	return b.generateResult(), nil
}

// benchReqResult 保存单个请求的结果
type benchReqResult struct {
	statusCode int
	duration   time.Duration
	bytes      int64
	err        error
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

			// QPS限制
			if b.rateLimiter != nil {
				<-b.rateLimiter.C
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
func (b *HTTPBench) worker(workCh <-chan struct{}, resultCh chan<- *benchReqResult, wg *sync.WaitGroup) {
	defer wg.Done()

	for range workCh {
		result := b.doRequest()
		resultCh <- result
	}
}

// doRequest 执行单个请求
func (b *HTTPBench) doRequest() *benchReqResult {
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
	result := &benchReqResult{
		duration: duration,
		err:      err,
	}

	if err != nil {
		result.statusCode = 0
		atomic.AddInt64(&b.failReqs, 1)
		return result
	}

	result.statusCode = resp.StatusCode

	// 统计状态码
	b.mu.Lock()
	b.statusCodes[resp.StatusCode]++
	b.respTimes = append(b.respTimes, duration)
	b.mu.Unlock()

	// 统计字节数
	if resp.Body != nil {
		bodyBytes := resp.BodyBuffer().Bytes()
		result.bytes = int64(len(bodyBytes))
		atomic.AddInt64(&b.totalBytes, int64(len(bodyBytes)))
	}

	if resp.IsOK() {
		atomic.AddInt64(&b.successReqs, 1)
	} else {
		atomic.AddInt64(&b.failReqs, 1)
	}

	resp.CloseBody()
	return result
}

// collectResults 收集结果
func (b *HTTPBench) collectResults(resultCh <-chan *benchReqResult) {
	for result := range resultCh {
		_ = result // 结果已经在doRequest中统计了
	}
}

// generateResult 生成最终结果
func (b *HTTPBench) generateResult() *BenchResult {
	duration := b.endTime.Sub(b.startTime)

	b.mu.Lock()
	defer b.mu.Unlock()

	result := &BenchResult{
		URL:            b.URL,
		TotalReqs:      b.totalReqs,
		SuccessReqs:    b.successReqs,
		FailReqs:       b.failReqs,
		Duration:       duration,
		StatusCodes:    make(map[int]int64),
		ReqsPerSecond:  float64(b.totalReqs) / duration.Seconds(),
		BytesPerSecond: float64(b.totalBytes) / duration.Seconds(),
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

// String 格式化输出结果
func (r *BenchResult) String() string {
	// 设置 buf 初始容量
	buf := make([]byte, 0, 256)

	// buf = append(buf, fmt.Sprintf("Benchmarking %s\n", r.URL)...)
	buf = append(buf, fmt.Sprintf("Total      requests: %d\n", r.TotalReqs)...)
	buf = append(buf, fmt.Sprintf("Successful requests: <green>%d</>\n", r.SuccessReqs)...)
	buf = append(buf, fmt.Sprintf("Failed     requests: <red>%d</>\n", r.FailReqs)...)
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
			buf = append(buf, fmt.Sprintf("  <green>%d</>: %d\n", code, count)...)
		}
	}

	return string(buf)
}
