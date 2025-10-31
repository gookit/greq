package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gookit/goutil/cflag"
	"github.com/gookit/goutil/netutil/httpctype"
	"github.com/gookit/goutil/netutil/httpreq"
	"github.com/gookit/goutil/strutil"
	"github.com/gookit/goutil/x/ccolor"
	"github.com/gookit/greq"
	"github.com/gookit/greq/ext/httpfile"
	"github.com/gookit/greq/requtil"
)

var cmdOpts = struct {
	method   string
	data     string
	headers  cflag.KVString
	formData cflag.KVString
	timeout  int
	output   string
	raw      string
	httpVars  cflag.KVString // HTTP request variables
	down     bool
	verbose  bool
	silent   bool
	follow   bool
	insecure bool
	json     bool // quick set Content-Type: application/json
	agent    string // custom user-agent
	headOnly bool // show response headers only
}{
	headers:  cflag.KVString{Sep: ":"},
	formData: cflag.KVString{Sep: "="},
	httpVars:  cflag.KVString{Sep: "="},
}

// 实现类似curl的http请求工具
//
// Install:
//
//  go install ./cmd/greq # install from source code
//	go install github.com/gookit/greq/cmd/greq@latest
func main() {
	cmd := cflag.New(func(c *cflag.CFlags) {
		c.Desc = "Lightweight HTTP request tool, like curl"
		c.Version = "1.0.0"
	})

	// 添加选项
	cmd.StringVar(&cmdOpts.method, "method", "GET", "HTTP method;;X")
	cmd.StringVar(&cmdOpts.data, "data", "", "HTTP request body data;;d")
	cmd.StringVar(&cmdOpts.agent, "agent", "", "Custom set User-Agent;;A")
	cmd.Var(&cmdOpts.headers, "header", `Custom HTTP header, allow multi. eg: "Foo: bar";;H`)
	cmd.Var(&cmdOpts.formData, "form", `Custom HTTP form data, allow multi. eg: "key=value";;F`)
	cmd.IntVar(&cmdOpts.timeout, "timeout", 30, "Request timeout in seconds;;t")
	cmd.StringVar(&cmdOpts.output, "output", "", "Output file for response;;o")
	cmd.StringVar(&cmdOpts.raw, "raw", "", `Parse and send IDE .http format request file
With request match:
 filepath#keywords  - match request by keywords in .http file content,
                      multiple keywords separated by comma ','
                      on exists multiple requests, will prompt to select one. // TODO
;;r`)
	cmd.Var(&cmdOpts.httpVars, "var", `(.http file)HTTP request variables, allow multi. eg: "key=value";;V`)

	cmd.BoolVar(&cmdOpts.down, "down", false, "Treat URL as download link;;O")
	cmd.BoolVar(&cmdOpts.verbose, "verbose", false, "Verbose output;;v")
	cmd.BoolVar(&cmdOpts.silent, "silent", false, "Silent mode;;s")
	cmd.BoolVar(&cmdOpts.follow, "follow", false, "Follow redirects;;L")
	cmd.BoolVar(&cmdOpts.insecure, "insecure", false, "Allow insecure SSL connections;;k")
	cmd.BoolVar(&cmdOpts.json, "json", false, "Quick set Content-Type: application/json")
	cmd.BoolVar(&cmdOpts.headOnly, "head", false, "Show response headers only;;I")

	cmd.AddArg("url", "the URL to request", false, nil)

	cmd.Example = `
  # GET request
  greq https://example.com

  # With headers
  greq -H "Authorization: Bearer token" -H "Content-Type: application/json" https://example.com

  # POST request with JSON data
  greq -X POST --json -d '{"key":"value"}' https://example.com

  # Download file
  greq -O https://example.com/file.zip
	`

	cmd.Func = func(c *cflag.CFlags) error {
		return runRequest(c)
	}

	cmd.MustRun(nil)
}

// runRequest 执行HTTP请求
func runRequest(c *cflag.CFlags) error {
	url := c.Arg("url").String()

	// 处理 --raw 选项：解析IDE .http格式文件
	if cmdOpts.raw != "" {
		return handleRawRequest(cmdOpts.raw)
	}

	// 处理 --down 选项：下载文件
	if cmdOpts.down {
		if url == "" {
			return fmt.Errorf("URL is required for download")
		}
		return handleDownload(url)
	}

	// 普通HTTP请求需要URL
	if url == "" {
		return fmt.Errorf("the URL is required")
	}

	// 普通HTTP请求
	return handleNormalRequest(url)
}

// handleRawRequest 处理IDE .http格式文件
func handleRawRequest(filename string) error {
	var keywords []string
    if strings.Contains(filename, "#") {
        // 处理 filepath#keywords 格式
        parts := strings.SplitN(filename, "#", 2)
        filename = parts[0]
		keywords = strings.Split(parts[1], ",")
    }

	// 解析 .http 文件格式
	hf, err := httpfile.ParseHTTPFile(filename)
	if err != nil {
		return fmt.Errorf("failed to parse HTTP file: %v", err)
	}

	request := hf.SearchOne(keywords...)
	if request == nil {
		return fmt.Errorf("no request found with keywords: %v", keywords)
	}

	// 应用变量替换
	request.ApplyVars(cmdOpts.httpVars.Data())

	if !cmdOpts.silent {
		ccolor.Infoln("Parsed HTTP request from file:")
		ccolor.Printf("  Method: %s\n", request.Method)
		ccolor.Printf("  URL: %s\n", request.URL)
		if len(request.Headers) > 0 {
			ccolor.Println("  Headers:")
			for k, v := range request.Headers {
				ccolor.Printf("    %s: %s\n", k, v)
			}
		}
		if request.Body != "" {
			ccolor.Printf("  Body: %s\n", strutil.Substr(request.Body, 0, 256))
		}
	}

	// 发送请求
	return sendParsedRequest(request)
}

// sendParsedRequest 发送解析后的HTTP请求
func sendParsedRequest(request *httpfile.HTTPRequest) error {
	// 创建请求选项
	optFns := []greq.OptionFn{}

	if cmdOpts.timeout > 0 {
		optFns = append(optFns, greq.WithTimeout(cmdOpts.timeout*1000)) // 转换为毫秒
	}

	// 设置头部
	for k, v := range request.Headers {
		optFns = append(optFns, greq.WithHeader(k, v))
	}

	bodyData := []byte(request.Body)

	// 设置主体数据（使用WithData而不是WithBody，因为greq库有bug）
	if len(bodyData) > 0 {
		optFns = append(optFns, greq.WithData(bodyData))
	}

	// 发送请求
	resp, err := greq.SendDo(request.Method, request.URL, optFns...)
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}

	return outputResponse(resp)
}

// handleDownload 处理下载请求
func handleDownload(url string) error {
	// 创建请求选项
	optFns := []greq.OptionFn{}

	if cmdOpts.timeout > 0 {
		optFns = append(optFns, greq.WithTimeout(cmdOpts.timeout*1000)) // 转换为毫秒
	}

	// 设置请求头
	for k, v := range cmdOpts.headers.Data() {
		optFns = append(optFns, greq.WithHeader(k, v))
	}

	// 获取文件名
	filename := cmdOpts.output
	if filename == "" {
		// 临时使用空响应，稍后会用HEAD响应替换
		filename = getFilenameFromURL(url, nil)
	}

	// 首先发送HEAD请求获取文件信息
	headResp, err := greq.HeadDo(url, optFns...)
	if err != nil {
		return fmt.Errorf("head request failed: %v", err)
	}

	// 获取文件大小
	var totalSize int64
	contentLength := headResp.Header.Get("Content-Length")
	if contentLength != "" {
		if size, err := strconv.ParseInt(contentLength, 10, 64); err == nil {
			totalSize = size
		}
	}

	// 从HEAD响应更新文件名
	if cmdOpts.output == "" {
		filename = getFilenameFromURL(url, headResp)
	}

	// 检查本地文件是否存在，实现断点续传
	var startOffset int64
	fileInfo, err := os.Stat(filename)
	if err == nil && fileInfo.Size() > 0 {
		startOffset = fileInfo.Size()
		if startOffset >= totalSize && totalSize > 0 {
			if !cmdOpts.silent {
				ccolor.Greenf("File already downloaded completely: %s (%s)\n", filename, formatBytes(int(totalSize)))
			}
			return nil
		}
		if !cmdOpts.silent {
			ccolor.Yellowf("Resuming download from: %s (%s/%s)\n", filename, formatBytes(int(startOffset)), formatBytes(int(totalSize)))
		}
	}

	// 如果有断点，添加Range头
	if startOffset > 0 {
		optFns = append(optFns, greq.WithHeader("Range", fmt.Sprintf("bytes=%d-", startOffset)))
	}

	if !cmdOpts.silent {
		ccolor.Cyanf("Downloading to: %s\n", filename)
		if totalSize > 0 {
			ccolor.Cyanf("File size: %s\n", formatBytes(int(totalSize)))
		}
	}

	// 发送GET请求下载文件
	resp, err := greq.GetDo(url, optFns...)
	if err != nil {
		return fmt.Errorf("download request failed: %v", err)
	}

	// 使用带进度显示的下载
	err = downloadWithProgress(resp, filename, totalSize, startOffset)
	if err != nil {
		return fmt.Errorf("download failed: %v", err)
	}

	return nil
}

// handleNormalRequest 处理普通HTTP请求
func handleNormalRequest(url string) error {
	// 创建请求选项
	optFns := []greq.OptionFn{}
	reqMethod := strings.ToUpper(cmdOpts.method)

	// 自定义User-Agent
	if cmdOpts.agent != "" {
		optFns = append(optFns, greq.WithUserAgent(cmdOpts.agent))
	}

	if cmdOpts.timeout > 0 {
		optFns = append(optFns, greq.WithTimeout(cmdOpts.timeout*1000)) // 转换为毫秒
	}

	// 设置请求头
	for k, v := range cmdOpts.headers.Data() {
		optFns = append(optFns, greq.WithHeader(k, v))
	}

	// 快速设置 Content-Type: application/json
	if cmdOpts.json {
		optFns = append(optFns, greq.WithContentType(httpctype.JSON))
		if httpreq.IsNoBodyMethod(reqMethod) {
			reqMethod = "POST"
		}
	}

	// 准备请求数据
	var bodyData []byte
	if cmdOpts.data != "" {
		bodyData = []byte(cmdOpts.data)
	} else if !cmdOpts.formData.IsEmpty() {
		optFns = append(optFns, greq.WithContentType(httpctype.Form))
		uvs := httpreq.MakeQuery(cmdOpts.formData.Data())
		if len(uvs) > 0 {
			bodyData = []byte(uvs.Encode())
		}
	}

	if len(bodyData) > 0 {
		optFns = append(optFns, greq.WithBody(bodyData))
		if httpreq.IsNoBodyMethod(reqMethod) {
			reqMethod = "POST"
		}
	}

	if !cmdOpts.silent {
		ccolor.Cyanf("Requesting URL: %s %s\n", reqMethod, url)
	}

	greq.Std().BeforeSend = func(req *http.Request) error {
		if !cmdOpts.verbose {
			return nil
		}
		ccolor.Cyanln("Request Header:")
		for k, v := range req.Header {
			ccolor.Printf("  <green>%s</>: %s\n", k, strings.Join(v, ", "))
		}
		if len(bodyData) > 0 {
			ccolor.Cyanln("Request   Body:")
			fmt.Printf("  %s\n\n", string(bodyData))
		}
		return nil
	}

	// 发送请求
	var err error
	var resp *greq.Response
	resp, err = greq.SendDo(reqMethod, url, optFns...)
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}

	// 输出响应
	return outputResponse(resp)
}

// outputResponse 输出响应结果
func outputResponse(resp *greq.Response) error {
	if cmdOpts.verbose || cmdOpts.headOnly {
		ccolor.Infoln("Response Headers:")
		for k, v := range resp.Header {
			ccolor.Printf("  %s: %s\n", k, strings.Join(v, ", "))
		}
		ccolor.Println()

		// 只显示响应头
		if cmdOpts.headOnly {
			return nil
		}
	}

	// 输出到文件或标准输出
	if cmdOpts.output != "" {
		return os.WriteFile(cmdOpts.output, []byte(resp.BodyString()), 0644)
	}

	// 输出到标准输出
	fmt.Print(resp.BodyString())
	return nil
}

// getFilenameFromURL 从URL获取文件名
func getFilenameFromURL(url string, resp *greq.Response) string {
	// 尝试从Content-Disposition获取文件名
	if resp != nil {
		disposition := resp.Header.Get("Content-Disposition")
		if filename := requtil.FilenameFromDisposition(disposition); filename != "" {
			return filename
		}
	}

	// 从URL路径获取文件名
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		lastPart := parts[len(parts)-1]
		if lastPart != "" && !strings.Contains(lastPart, "?") {
			return lastPart
		}
	}

	// 默认文件名
	return fmt.Sprintf("download_%d", time.Now().Unix())
}

// handleDownloadFromResponse 处理从响应下载文件
func handleDownloadFromResponse(url string, resp *greq.Response) error {
	if resp.IsFail() {
		return fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	// 获取文件名
	filename := getFilenameFromURL(url, resp)
	if cmdOpts.output != "" {
		filename = cmdOpts.output
	}

	// 写入文件
	n, err := resp.SaveFile(filename)
	if err != nil {
		return fmt.Errorf("failed to save file: %v", err)
	}

	if !cmdOpts.silent {
		ccolor.Successf("Download completed: %s (%s)\n", filename, formatBytes(n))
	}
	return nil
}

// downloadWithProgress 带进度显示的下载
func downloadWithProgress(resp *greq.Response, filename string, totalSize int64, startOffset int64) error {
	// 创建或打开文件
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("open file failed: %v", err)
	}
	defer file.Close()

	// 获取响应体
	bodyReader := resp.Body
	if bodyReader == nil {
		return fmt.Errorf("response body is nil")
	}
	defer resp.CloseBody()

	// 创建缓冲区
	buffer := make([]byte, 32*1024) // 32KB 缓冲区
	var downloaded int64 = startOffset

	// 进度显示定时器
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	done := make(chan error, 1)

	// 启动进度显示协程
	go func() {
		for {
			select {
			case <-ticker.C:
				showDownloadProgress(downloaded, totalSize, time.Now())
			case <-done:
				return
			}
		}
	}()

	// 读取并写入数据
	for {
		n, err := bodyReader.Read(buffer)
		if n > 0 {
			_, writeErr := file.Write(buffer[:n])
			if writeErr != nil {
				done <- writeErr
				return writeErr
			}
			downloaded += int64(n)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			done <- err
			return err
		}
	}

	// 完成下载
	done <- nil
	close(done)

	// 显示最终进度
	showDownloadProgress(downloaded, totalSize, time.Now())
	fmt.Println() // 换行

	if !cmdOpts.silent {
		ccolor.Greenf("Download completed: %s (%s)\n", filename, formatBytes(int(downloaded)))
	}

	return nil
}

// showDownloadProgress 显示下载进度
func showDownloadProgress(downloaded int64, totalSize int64, startTime time.Time) {
	if totalSize <= 0 {
		// 未知总大小，只显示已下载量
		fmt.Printf("\rDownloaded: %s", formatBytes(int(downloaded)))
		return
	}

	percentage := float64(downloaded) / float64(totalSize) * 100
	if percentage > 100 {
		percentage = 100
	}

	// 计算速度（字节/秒）
	elapsed := time.Since(startTime).Seconds()
	if elapsed <= 0 {
		elapsed = 1 // 避免除零
	}
	speed := float64(downloaded) / elapsed

	// 计算剩余时间
	if speed > 0 && downloaded < totalSize {
		remaining := float64(totalSize-downloaded) / speed
		fmt.Printf("\r[%-30s] %.1f%% %s/s %s",
			strings.Repeat("█", int(percentage/100*30)) + strings.Repeat("░", 30-int(percentage/100*30)),
			percentage,
			formatBytes(int(speed)),
			formatDuration(time.Duration(int(remaining))*time.Second))
	} else {
		fmt.Printf("\r[%-30s] %.1f%% %s/s",
			strings.Repeat("█", int(percentage/100*30)) + strings.Repeat("░", 30-int(percentage/100*30)),
			percentage,
			formatBytes(int(speed)))
	}
}

// formatDuration 格式化时间间隔
func formatDuration(d time.Duration) string {
	seconds := int(d.Seconds())
	if seconds < 0 {
		seconds = 0
	}
	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	secs := seconds % 60

	if hours > 0 {
		return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, secs)
	}
	return fmt.Sprintf("%02d:%02d", minutes, secs)
}

// formatBytes 格式化字节大小
func formatBytes(bytes int) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
