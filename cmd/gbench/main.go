package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/gookit/goutil/cflag"
	"github.com/gookit/goutil/fsutil"
	"github.com/gookit/goutil/timex"
	"github.com/gookit/goutil/x/ccolor"
	"github.com/gookit/greq/ext/bench"
)

// Build-time variables injected via -ldflags
var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildTime = "unknown"
)

var showVersion bool
var benchOpts = struct {
	number      int
	concurrency int
	// Duration of application to send requests. If duration is specified, n is ignored.
	duration string
	// Data http request body for POST/PUT requests.
	data     string
	qpsLimit int
	timeout  int // Timeout for each request
	output   string
	headers  cflag.KVString
	json     bool
	method   string
	progress bool // 是否显示进度条
}{
	headers: cflag.KVString{Sep: ":"},
}

// RUN:
//
//	gbench -n 1000 -c 10 https://www.baidu.com
func main() {
	cmd := cflag.New(func(c *cflag.CFlags) {
		c.Desc = fmt.Sprintf("Lightweight Benchmark HTTP requests, like ab.\n Commit: %s, Build: %s", GitCommit, BuildTime)
		c.Version = Version
	})

	// add options
	cmd.IntVar(&benchOpts.number, "number", 100, "number of requests to run;;n")
	cmd.IntVar(&benchOpts.concurrency, "concurrency", 3, "number of multiple requests to make at a time;;c")
	cmd.StringVar(&benchOpts.duration, "duration", "", `duration of application to send requests. If duration is specified, <green>n</> is ignored.
Example:
 - 10s: 10 seconds
 - 1m: 1 minute
 - 1h: 1 hour;;z`)
	cmd.StringVar(&benchOpts.data, "data", "", `data http request body for POST/PUT requests.
Allow use <green>@filename</> to read data from file.
	;;d`)
	cmd.IntVar(&benchOpts.qpsLimit, "qps", 0, "rate limit for all, in queries per second (QPS);;q")
	cmd.StringVar(&benchOpts.output, "output", "stdout", `Output file to write the results to.
If not specified, results are written to stdout.;;o`)
	cmd.Var(&benchOpts.headers, "header", `Custom HTTP header. Examples: -H "foo: bar";;H`)
	cmd.IntVar(&benchOpts.timeout, "timeout", 10, "Timeout(seconds) for each request. Default to infinite.;;t")
	cmd.StringVar(&benchOpts.method, "method", "GET", "HTTP method;;m")
	cmd.BoolVar(&benchOpts.json, "json", false, "Quick add Content-Type: application/json header.")
	cmd.BoolVar(&benchOpts.progress, "progress", true, "Show progress bar.")
	cmd.BoolVar(&showVersion, "version", false, "Show version information.;;V")

	cmd.AddArg("url", "the URL to benchmark test", true, nil)

	cmd.Example = `
  # Simple benchmark test
  gbench -n 1000 -c 10 https://www.example.com

  # With header
  gbench -n 1000 -c 10  -H "Authorization: Bearer token" https://www.example.com

  # POST request with JSON data
  gbench -n 1000 -c 10 -m POST -d '{"key":"value"}' https://www.example.com
	`

	cmd.AfterFlagParse = func(c *cflag.CFlags) bool {
		if showVersion {
			ccolor.Printf("Version: <green>%s</> (%s, %s)\n", c.Version, GitCommit, BuildTime)
			return false
		}
		return true
	}

	cmd.Func = func(c *cflag.CFlags) error {
		return runBenchmark(c)
	}
	cmd.MustRun(nil)
}

// runBenchmark 运行基准测试
func runBenchmark(c *cflag.CFlags) error {
	url := c.Arg("url").String()
	if url == "" {
		return fmt.Errorf("URL is required")
	}

	var duration time.Duration
	if benchOpts.duration != "" {
		var err error
		duration, err = parseDuration(benchOpts.duration)
		if err != nil {
			return fmt.Errorf("invalid duration format: %v", err)
		}
	}

	// 读取请求数据
	var bodyData []byte
	if benchOpts.data != "" {
		var err error
		bodyData, err = readData(benchOpts.data)
		if err != nil {
			return fmt.Errorf("failed to read data: %v", err)
		}
	}

	// 解析请求头
	headers := make(map[string]string)
	for k, v := range benchOpts.headers.Data() {
		headers[k] = v
	}

	// 如果指定了 JSON 数据，快速添加 Content-Type 头
	if benchOpts.json && benchOpts.method != "GET" {
		headers["Content-Type"] = "application/json"
	}

	// 创建基准测试
	hb := bench.NewHTTPBench(url)
	hb.SetMethod(benchOpts.method).
		SetConcurrency(benchOpts.concurrency).
		SetNumber(benchOpts.number).
		SetDuration(duration).
		SetHeaders(headers).
		SetBody(bodyData)

	if benchOpts.timeout > 0 {
		hb.SetTimeout(time.Duration(benchOpts.timeout) * time.Second)
	}

	if benchOpts.qpsLimit > 0 {
		hb.SetQPSLimit(benchOpts.qpsLimit)
	}

	ccolor.Cyanf("Benchmark URL: %s\n", url)
	ccolor.Infoln("Configuration:")
	ccolor.Printf("  Method=%s, Concurrency=%d, Number=%d, Duration=%v\n",
		benchOpts.method, benchOpts.concurrency, benchOpts.number, duration)
	ccolor.Printf("  The  QPS  Limit: %d\n", benchOpts.qpsLimit)
	ccolor.Printf("  Request Timeout: %d seconds\n", benchOpts.timeout)

	// 显示简单的 cli ascii 进度条
	if benchOpts.progress {
		fmt.Println()
		hb.SetShowProgress(true)
	} else {
		ccolor.Infof("Benchmarking ... please wait.\n")
	}

	// Wire Ctrl+C / SIGTERM into a cancellable context so the test can be
	// interrupted gracefully and still print partial results — without this,
	// large -n values would leave the user stuck watching the progress bar.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	var interrupted atomic.Bool
	go func() {
		if _, ok := <-sigCh; !ok {
			return
		}
		interrupted.Store(true)
		ccolor.Yellowf("\nInterrupt received, stopping benchmark and showing partial results...\n")
		cancel()
	}()

	// 执行测试
	result, err := hb.RunCtx(ctx)
	if err != nil {
		return fmt.Errorf("benchmark failed: %v", err)
	}

	// 输出结果 — 终端用带色版本，文件用纯文本版本（避免 <green>...</> 写进文件）
	if benchOpts.output == "stdout" {
		ccolor.Successf("\nBenchmark Results:\n")
		ccolor.Print(result.String())
	} else {
		ccolor.Successf("Benchmark completed successfully!\n")
		err = os.WriteFile(benchOpts.output, []byte(result.PlainString()), 0644)
		if err != nil {
			return fmt.Errorf("failed to write output file: %v", err)
		}
		ccolor.Infof("Results written to %s\n", benchOpts.output)
	}

	// Reflect interruption in the exit code so CI / scripts can tell a
	// signalled run apart from a clean completion.
	if interrupted.Load() {
		return fmt.Errorf("benchmark interrupted by signal")
	}
	return nil
}

// parseDuration 解析持续时间字符串
func parseDuration(s string) (time.Duration, error) {
	if s == "" {
		return 0, nil
	}

	// 尝试标准格式
	d, err := timex.ToDuration(s)
	if err == nil {
		return d, nil
	}

	return 0, fmt.Errorf("invalid duration format: %s", s)
}

// readData 读取数据，支持从文件读取
func readData(data string) ([]byte, error) {
	if strings.HasPrefix(data, "@") {
		// 从文件读取
		filename := data[1:]
		if filename != "" && fsutil.IsFile(filename) {
			ccolor.Infof("Read data from file: %s\n", filename)
			return os.ReadFile(filename)
		}
	}

	// 直接使用数据
	return []byte(data), nil
}
