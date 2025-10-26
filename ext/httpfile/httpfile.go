// Package httpfile provides HTTP request file(.http) parsing utilities.
package httpfile

import (
	"fmt"
	"os"
	"strings"

	"github.com/gookit/goutil/strutil"
)

// isValidHTTPMethod 检查是否是有效的 HTTP 方法
func isValidHTTPMethod(method string) bool {
	validMethods := []string{
		"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS", "CONNECT", "TRACE",
	}

	for _, m := range validMethods {
		if strings.ToUpper(method) == m {
			return true
		}
	}

	return false
}

// HTTPRequest represents an HTTP request.
type HTTPRequest struct {
	// Name is the name of the HTTP request. parsed from ### line
	Name    string
	Comments []string
	// Method is the HTTP method of the request.
	Method  string
	URL     string
	Headers map[string]string
	Body    string
}

// HTTPFile represents an HTTP request file. It contains a list of HTTP requests.
//
// 文件内容格式:
//    - 每个HTTP请求以方法、URL和空行开始
//    - 头部键值对每行一个，格式为 "Key: Value"
//    - 空行后为请求体（可选） `@filename` 可以指定请求体内容从文件中读取
//    - 可以使用变量替换请求中的内容，格式为 `${var_name}`
//    - 每个请求之间用 `空行+###开头的行` 分隔
//    - 单个 # 开头的行是注释，会被忽略
type HTTPFile struct {
	// FilePath is the path of the HTTP request file.
	FilePath string
	Contents string // the contents of the HTTP request file
	Requests []*HTTPRequest
}

// ParseFileContent parse a HTTP request file content.
func ParseFileContent(contents string) (*HTTPFile, error) {
	// 如果内容为空，直接返回空列表
	if strings.TrimSpace(contents) == "" {
		return nil, fmt.Errorf("input contents is empty")
	}

	hf := &HTTPFile{Contents: contents}
	err := hf.Parse()
	return hf, err
}

// ParseHTTPFile parse a HTTP request file.
func ParseHTTPFile(filePath string) (*HTTPFile, error) {
	hf := &HTTPFile{FilePath: filePath}
	err := hf.Parse()
	return hf, err
}

// SearchName search requests by keywords. will return all requests that name contains all keywords.
func (hf *HTTPFile) SearchName(keywords ...string) []*HTTPRequest {
	var foundReqs []*HTTPRequest
	for _, req := range hf.Requests {
		if strutil.ContainsAll(req.Name, keywords) {
			foundReqs = append(foundReqs, req)
		}
	}
	return foundReqs
}

// SearchOne search one request by keywords. will return the first request that name contains all keywords.
func (hf *HTTPFile) SearchOne(keywords ...string) *HTTPRequest {
	for _, req := range hf.Requests {
		if strutil.ContainsAll(req.Name, keywords) {
			return req
		}
	}
	return nil
}

// FindByName find a HTTP request by name.
func (hf *HTTPFile) FindByName(name string) *HTTPRequest {
	for _, req := range hf.Requests {
		if req.Name == name {
			return req
		}
	}
	return nil
}

// Parse do parse HTTP request file content.
func (hf *HTTPFile) Parse() error {
	if hf.Contents == "" {
		// load file contents
		contents, err := os.ReadFile(hf.FilePath)
		if err != nil {
			return fmt.Errorf("read http file contents error: %w", err)
		}
		hf.Contents = string(contents)
	}

	// 如果内容为空，直接返回
	if strings.TrimSpace(hf.Contents) == "" {
		hf.Requests = []*HTTPRequest{}
		return nil
	}

	rawLines := strings.Split(hf.Contents, "\n")
	var currentReq *HTTPRequest
	var inHeaders bool // 标记是否在解析头部
	var inBody bool    // 标记是否在解析请求体
	var hasBodyStart bool // 标记是否已经遇到请求体开始
	var globalComments []string // 存储全局注释

	for _, line := range rawLines {
		// 保留原始行用于请求体（可能包含空格）
		trimmedLine := strings.TrimSpace(line)

		// 处理空行
		if trimmedLine == "" {
			// 空行表示头部结束，开始请求体
			if currentReq != nil && !inBody && !hasBodyStart && inHeaders {
				inHeaders = false
				inBody = true
				hasBodyStart = true
			} else if currentReq != nil && inBody {
				// 如果已经在请求体中，保留空行
				currentReq.Body += "\n"
			}
			continue
		}

		// 处理注释行（单个#开头）
		if strings.HasPrefix(trimmedLine, "#") && !strings.HasPrefix(trimmedLine, "###") {
			if currentReq != nil {
				currentReq.Comments = append(currentReq.Comments, line)
			} else {
				// 如果没有当前请求，将注释添加到全局注释中
				globalComments = append(globalComments, line)
			}
			continue
		}

		// 处理请求分隔符（###开头）
		if strings.HasPrefix(trimmedLine, "###") {
			// 保存当前请求
			if currentReq != nil {
				// 去除请求体末尾的所有换行符
				currentReq.Body = strings.TrimRight(currentReq.Body, "\n")
				hf.Requests = append(hf.Requests, currentReq)
			}

			// 创建新请求
			currentReq = &HTTPRequest{
				Name:    strings.TrimSpace(strings.TrimPrefix(trimmedLine, "###")),
				Headers: make(map[string]string),
				Comments: make([]string, 0),
			}
			// 将全局注释添加到当前请求
			currentReq.Comments = append(currentReq.Comments, globalComments...)
			inHeaders = false
			inBody = false
			hasBodyStart = false
			continue
		}

		// 如果没有当前请求，创建一个
		if currentReq == nil {
			currentReq = &HTTPRequest{
				Headers: make(map[string]string),
				Comments: make([]string, 0),
			}
			// 将全局注释添加到当前请求
			currentReq.Comments = append(currentReq.Comments, globalComments...)
		}

		// 处理请求体
		if inBody {
			currentReq.Body += line + "\n"
			continue
		}

		// 处理请求行（方法 URL）
		if !inHeaders {
			parts := strings.Fields(trimmedLine)
			if len(parts) >= 2 {
				currentReq.Method = parts[0]
				currentReq.URL = parts[1]
				inHeaders = true
			}
			continue
		}

		// 处理头部行（Key: Value）
		if inHeaders {
			if colonIndex := strings.Index(trimmedLine, ":"); colonIndex > 0 {
				key := strings.TrimSpace(trimmedLine[:colonIndex])
				value := strings.TrimSpace(trimmedLine[colonIndex+1:])
				currentReq.Headers[key] = value
			}
		}
	}

	// 添加最后一个请求
	if currentReq != nil {
		// 去除请求体末尾的所有换行符
		currentReq.Body = strings.TrimRight(currentReq.Body, "\n")
		hf.Requests = append(hf.Requests, currentReq)
	}

	return nil
}
