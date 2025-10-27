package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gookit/goutil/cflag"
	"github.com/gookit/goutil/netutil/httpctype"
	"github.com/gookit/goutil/netutil/httpreq"
	"github.com/gookit/goutil/strutil"
	"github.com/gookit/goutil/x/ccolor"
	"github.com/gookit/greq"
	"github.com/gookit/greq/ext/httpfile"
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

	greq.Std().BeforeSend = func(req *http.Request) {
		if !cmdOpts.verbose {
			return
		}
		ccolor.Cyanln("Request Header:")
		for k, v := range req.Header {
			ccolor.Printf("  <green>%s</>: %s\n", k, strings.Join(v, ", "))
		}
		if len(bodyData) > 0 {
			ccolor.Cyanln("Request   Body:")
			fmt.Printf("  %s\n\n", string(bodyData))
		}
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
