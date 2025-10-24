package main

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/gookit/goutil/cflag"
	"github.com/gookit/goutil/fsutil"
	"github.com/gookit/goutil/strutil"
	"github.com/gookit/goutil/x/ccolor"
	"github.com/gookit/greq"
)

var cmdOpts = struct {
	method   string
	data     string
	headers  cflag.KVString
	timeout  int
	output   string
	raw      string
	down     bool
	verbose  bool
	silent   bool
	follow   bool
	insecure bool
	json     bool // quick set Content-Type: application/json
	agent    string // custom user-agent
}{
	headers: cflag.KVString{Sep: ":"},
}

// 实现类似curl的http请求工具
func main() {
	cmd := cflag.New(func(c *cflag.CFlags) {
		c.Desc = "Simple HTTP request tool, like curl"
		c.Version = "1.0.0"
	})

	// 添加选项
	cmd.StringVar(&cmdOpts.method, "method", "GET", "HTTP method;;X")
	cmd.StringVar(&cmdOpts.data, "data", "", "HTTP request body data;;d")
	cmd.StringVar(&cmdOpts.agent, "agent", "", "Custom set User-Agent;;A")
	cmd.Var(&cmdOpts.headers, "header", `Custom HTTP header, allow multi. eg: "Foo: bar";;H`)
	cmd.IntVar(&cmdOpts.timeout, "timeout", 30, "Request timeout in seconds;;t")
	cmd.StringVar(&cmdOpts.output, "output", "", "Output file for response;;o")
	cmd.StringVar(&cmdOpts.raw, "raw", "", `Parse and send IDE .http format request file;;r`)
	cmd.BoolVar(&cmdOpts.down, "down", false, "Treat URL as download link;;O")
	cmd.BoolVar(&cmdOpts.verbose, "verbose", false, "Verbose output;;v")
	cmd.BoolVar(&cmdOpts.silent, "silent", false, "Silent mode;;s")
	cmd.BoolVar(&cmdOpts.follow, "follow", false, "Follow redirects;;L")
	cmd.BoolVar(&cmdOpts.insecure, "insecure", false, "Allow insecure SSL connections;;k")
	cmd.BoolVar(&cmdOpts.json, "json", false, "Quick set Content-Type: application/json")

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
	if !fsutil.IsFile(filename) {
		return fmt.Errorf("raw file not found: %s", filename)
	}

	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read raw file: %v", err)
	}

	// 解析 .http 文件格式
	request, err := parseHTTPFile(string(content))
	if err != nil {
		return fmt.Errorf("failed to parse HTTP file: %v", err)
	}

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
			ccolor.Printf("  Body: %s\n", strutil.Substr(request.Body, 0, 100))
		}
	}

	// 发送请求
	return sendParsedRequest(request)
}

// handleDownload 处理文件下载
func handleDownload(url string) error {
	if !cmdOpts.silent {
		ccolor.Infof("Downloading from: %s\n", url)
	}

	// 创建请求选项
	optFns := []greq.OptionFn{}
	if cmdOpts.timeout > 0 {
		optFns = append(optFns, greq.WithTimeout(cmdOpts.timeout*1000)) // 转换为毫秒
	}

	// 发送GET请求
	resp, err := greq.GetDo(url, optFns...)
	if err != nil {
		return fmt.Errorf("download failed: %v", err)
	}

	if resp.IsFail() {
		return fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	// 获取文件名
	filename := getFilenameFromURL(url, resp)
	if cmdOpts.output != "" {
		filename = cmdOpts.output
	}

	// 写入文件
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	written, err := io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	if !cmdOpts.silent {
		ccolor.Successf("Download completed: %s (%s)\n", filename, formatBytes(written))
	}

	return nil
}

// handleNormalRequest 处理普通HTTP请求
func handleNormalRequest(url string) error {
	if !cmdOpts.silent {
		ccolor.Infof("Requesting: %s %s\n", cmdOpts.method, url)
	}

	// 创建请求选项
	optFns := []greq.OptionFn{}

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

	// 快速设置Content-Type: application/json
	if cmdOpts.json {
		optFns = append(optFns, greq.WithContentType("application/json"))
	}

	// 准备请求数据
	var bodyData []byte
	if cmdOpts.data != "" {
		bodyData = []byte(cmdOpts.data)
	}

	// 发送请求
	var resp *greq.Response
	var err error

	switch strings.ToUpper(cmdOpts.method) {
	case "GET":
		resp, err = greq.GetDo(url, optFns...)
	case "POST":
		if len(bodyData) > 0 {
			optFns = append(optFns, greq.WithContentType("application/x-www-form-urlencoded"))
			optFns = append(optFns, greq.WithBody(bodyData))
		}
		resp, err = greq.PostDo(url, bodyData, optFns...)
	case "PUT":
		if len(bodyData) > 0 {
			optFns = append(optFns, greq.WithBody(bodyData))
		}
		resp, err = greq.PutDo(url, bodyData, optFns...)
	case "DELETE":
		resp, err = greq.DeleteDo(url, optFns...)
	case "HEAD":
		resp, err = greq.HeadDo(url, optFns...)
	case "PATCH":
		if len(bodyData) > 0 {
			optFns = append(optFns, greq.WithBody(bodyData))
		}
		resp, err = greq.PatchDo(url, bodyData, optFns...)
	default:
		return fmt.Errorf("unsupported method: %s", cmdOpts.method)
	}

	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}

	// 输出响应
	return outputResponse(resp)
}

// HTTPRequest 表示解析的HTTP请求
type HTTPRequest struct {
	Method  string
	URL     string
	Headers map[string]string
	Body    string
}

// parseHTTPFile 解析IDE .http格式文件
func parseHTTPFile(content string) (*HTTPRequest, error) {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 {
		return nil, fmt.Errorf("empty file")
	}

	request := &HTTPRequest{
		Headers: make(map[string]string),
	}

	// 解析第一行：METHOD URL
	firstLine := strings.TrimSpace(lines[0])
	parts := strings.Fields(firstLine)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid request line: %s", firstLine)
	}

	request.Method = strings.ToUpper(parts[0])
	request.URL = parts[1]

	// 解析头部和主体
	i := 1
	for ; i < len(lines); i++ {
		line := lines[i]
		if line == "" {
			break // 空行表示头部结束
		}

		// 解析头部
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			request.Headers[key] = value
		}
	}

	// 解析主体（如果有）
	if i+1 < len(lines) {
		bodyLines := lines[i+1:]
		request.Body = strings.Join(bodyLines, "\n")
		request.Body = strings.TrimSpace(request.Body)
	}

	// 处理变量替换
	request.URL = replaceVariables(request.URL)
	request.Body = replaceVariables(request.Body)
	for k, v := range request.Headers {
		request.Headers[k] = replaceVariables(v)
	}

	return request, nil
}

// sendParsedRequest 发送解析后的HTTP请求
func sendParsedRequest(request *HTTPRequest) error {
	// 创建请求选项
	optFns := []greq.OptionFn{}

	if cmdOpts.timeout > 0 {
		optFns = append(optFns, greq.WithTimeout(cmdOpts.timeout*1000)) // 转换为毫秒
	}

	// 设置头部
	for k, v := range request.Headers {
		optFns = append(optFns, greq.WithHeader(k, v))
	}

	// 发送请求
	var resp *greq.Response
	var err error

	bodyData := []byte(request.Body)

	// 设置主体数据（使用WithData而不是WithBody，因为greq库有bug）
	if len(bodyData) > 0 {
		optFns = append(optFns, greq.WithData(bodyData))
	}

	switch request.Method {
	case "GET":
		resp, err = greq.GetDo(request.URL, optFns...)
	case "POST":
		resp, err = greq.PostDo(request.URL, nil, optFns...)
	case "PUT":
		resp, err = greq.PutDo(request.URL, nil, optFns...)
	case "DELETE":
		resp, err = greq.DeleteDo(request.URL, optFns...)
	case "HEAD":
		resp, err = greq.HeadDo(request.URL, optFns...)
	case "PATCH":
		resp, err = greq.PatchDo(request.URL, nil, optFns...)
	default:
		return fmt.Errorf("unsupported method: %s", request.Method)
	}

	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}

	// 处理下载
	if cmdOpts.down {
		return handleDownloadFromResponse(request.URL, resp)
	}

	return outputResponse(resp)
}

// outputResponse 输出响应结果
func outputResponse(resp *greq.Response) error {
	if cmdOpts.verbose {
		ccolor.Infoln("Response Headers:")
		for k, v := range resp.Header {
			ccolor.Printf("  %s: %s\n", k, strings.Join(v, ", "))
		}
		ccolor.Println()
	}

	// 输出到文件或标准输出
	if cmdOpts.output != "" {
		return os.WriteFile(cmdOpts.output, []byte(resp.BodyString()), 0644)
	}

	// 输出到标准输出
	fmt.Print(resp.BodyString())
	return nil
}

// replaceVariables 替换变量 ${var}
func replaceVariables(text string) string {
	re := regexp.MustCompile(`\$\{([^}]+)\}`)
	return re.ReplaceAllStringFunc(text, func(match string) string {
		varName := strings.TrimSpace(match[2 : len(match)-1])
		// 优先从环境变量获取
		if value := os.Getenv(varName); value != "" {
			return value
		}
		// 可以从其他配置源获取
		return match // 如果找不到变量，保持原样
	})
}

// getFilenameFromURL 从URL获取文件名
func getFilenameFromURL(url string, resp *greq.Response) string {
	// 尝试从Content-Disposition获取文件名
	if disposition := resp.Header.Get("Content-Disposition"); disposition != "" {
		if filename := extractFilename(disposition); filename != "" {
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
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	written, err := io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	if !cmdOpts.silent {
		ccolor.Successf("Download completed: %s (%s)\n", filename, formatBytes(written))
	}

	return nil
}

// extractFilename 从Content-Disposition提取文件名
func extractFilename(disposition string) string {
	parts := strings.Split(disposition, ";")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "filename=") {
			filename := strings.TrimPrefix(part, "filename=")
			return strings.Trim(filename, "\"'")
		}
	}
	return ""
}

// formatBytes 格式化字节大小
func formatBytes(bytes int64) string {
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
