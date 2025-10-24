package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gookit/goutil/cflag"
	"github.com/gookit/goutil/fsutil"
	"github.com/gookit/goutil/timex"
	"github.com/gookit/goutil/x/ccolor"
	"github.com/gookit/greq/ext"
)

var benchOpts = struct {
	number int
	concurrency int
	// Duration of application to send requests. If duration is specified, n is ignored.
	duration string
	// Data http request body for POST/PUT requests.
	data string
	qpsLimit int
	timeout int // Timeout for each request
	output string
	headers cflag.KVString
	method string
}{
	headers: cflag.KVString{Sep: ":"},
}

// RUN:
//  gbench -n 1000 -c 10 https://www.baidu.com
func main() {
	cmd := cflag.New(func(c *cflag.CFlags) {
		c.Desc = "Benchmark HTTP requests, like ab"
		c.Version = "v1.0.0"
	})

	// add options
	cmd.IntVar(&benchOpts.number,  "number", 100, "number of requests to run;;n")
	cmd.IntVar(&benchOpts.concurrency,  "concurrency", 3, "number of multiple requests to make at a time;;c")
	cmd.StringVar(&benchOpts.duration,  "duration", "", `duration of application to send requests. If duration is specified, <green>n</> is ignored.
Example:
 - 10s: 10 seconds
 - 1m: 1 minute
 - 1h: 1 hour;;z`)
	cmd.StringVar(&benchOpts.data,  "data", "", `data http request body for POST/PUT requests.
Allow use <green>@filename</> to read data from file.
	;;d`)
	cmd.IntVar(&benchOpts.qpsLimit,  "qps", 0, "rate limit for all, in queries per second (QPS);;q")
	cmd.StringVar(&benchOpts.output,  "output", "stdout", "Output file to write the results to. If not specified, results are written to stdout.;;o")
	cmd.Var(&benchOpts.headers,  "header", `Custom HTTP header. Examples: -H "foo: bar";;H`)
	cmd.IntVar(&benchOpts.timeout,  "timeout", 0, "Timeout for each request. Default to infinite.;;t")
	cmd.StringVar(&benchOpts.method,  "method", "GET", "HTTP method;;m")

	cmd.AddArg("url", "the URL to benchmark test", true, nil)

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

	ccolor.Magentaf("Benchmark URL: %s\n", url)
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

	// 创建基准测试
	bench := ext.NewHTTPBench(url)
	bench.SetMethod(benchOpts.method).
		SetConcurrency(benchOpts.concurrency).
		SetNumber(benchOpts.number).
		SetDuration(duration).
		SetHeaders(headers).
		SetBody(bodyData)

	if benchOpts.timeout > 0 {
		bench.SetTimeout(time.Duration(benchOpts.timeout) * time.Second)
	}

	if benchOpts.qpsLimit > 0 {
		bench.SetQPSLimit(benchOpts.qpsLimit)
	}

	ccolor.Infoln("Configuration:")
	ccolor.Printf("  Method=%s, concurrency=%d, number=%d, duration=%v\n",
		benchOpts.method, benchOpts.concurrency, benchOpts.number, duration)
	ccolor.Printf("  The QPS   Limit: %d\n", benchOpts.qpsLimit)
	ccolor.Printf("  Request Timeout: %d seconds\n", benchOpts.timeout)

	// 执行测试
	ccolor.Infof("Benchmarking %s (be patient)\n", url)
	result, err := bench.Run()
	if err != nil {
		return fmt.Errorf("benchmark failed: %v", err)
	}

	// 输出结果
	output := result.String()

	if benchOpts.output == "stdout" {
		ccolor.Successf("\nBenchmark Results:\n")
		ccolor.Print(output)
	} else {
		ccolor.Successf("Benchmark completed successfully!\n")
		err = os.WriteFile(benchOpts.output, []byte(output), 0644)
		if err != nil {
			return fmt.Errorf("failed to write output file: %v", err)
		}
		ccolor.Infof("Results written to %s\n", benchOpts.output)
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
