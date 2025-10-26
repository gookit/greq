package httpfile

import (
	"fmt"
	"strings"
)

// ParseOneRequest parse a HTTP request from content string.
func ParseOneRequest(content string) (*HTTPRequest, error) {
	lines := strings.Split(content, "\n")
	if len(lines) < 1 {
		return nil, fmt.Errorf("invalid http request content, at least 1 line")
	}

	req := &HTTPRequest{
		Headers: make(map[string]string),
		Comments: make([]string, 0),
	}

	var inHeaders bool // 标记是否在解析头部
	var inBody bool    // 标记是否在解析请求体
	var hasBodyStart bool // 标记是否已经遇到请求体开始
	var hasRequestLine bool // 标记是否已经处理请求行
	var hasHeaders bool // 标记是否已经有头部

	for _, line := range lines {
		// 保留原始行用于请求体（可能包含空格）
		trimmedLine := strings.TrimSpace(line)

		// 处理空行
		if trimmedLine == "" {
			// 空行表示头部结束，开始请求体
			if !inBody && !hasBodyStart && inHeaders {
				inHeaders = false
				inBody = true
				hasBodyStart = true
			} else if inBody {
				// 如果已经在请求体中，保留空行
				req.Body += "\n"
			}
			continue
		}

		// 处理注释行（单个#开头）
		if strings.HasPrefix(trimmedLine, "#") && !strings.HasPrefix(trimmedLine, "###") {
			req.Comments = append(req.Comments, line)
			continue
		}

		// 处理请求名称（###开头）
		if strings.HasPrefix(trimmedLine, "###") {
			req.Name = strings.TrimSpace(strings.TrimPrefix(trimmedLine, "###"))
			continue
		}

		// 处理请求体
		if inBody {
			req.Body += line + "\n"
			continue
		}

		// 处理请求行（方法 URL）
		if !inHeaders && !hasRequestLine {
			parts := strings.Fields(trimmedLine)
			if len(parts) >= 2 {
				// 检查是否是请求行（方法 URL），而不是头部行（Key: Value）
				// 请求行的第一个单词应该是 HTTP 方法
				method := parts[0]
				if isValidHTTPMethod(method) {
					req.Method = method
					req.URL = parts[1]
					inHeaders = true
					hasRequestLine = true
				} else {
					// 如果不是有效的 HTTP 方法，则认为是头部行，但没有请求行
					if colonIndex := strings.Index(trimmedLine, ":"); colonIndex > 0 {
						key := strings.TrimSpace(trimmedLine[:colonIndex])
						value := strings.TrimSpace(trimmedLine[colonIndex+1:])
						req.Headers[key] = value
						hasHeaders = true
					}
				}
			}
			continue
		}

		// 处理头部行（Key: Value）
		if inHeaders || hasHeaders {
			if colonIndex := strings.Index(trimmedLine, ":"); colonIndex > 0 {
				key := strings.TrimSpace(trimmedLine[:colonIndex])
				value := strings.TrimSpace(trimmedLine[colonIndex+1:])
				req.Headers[key] = value
				hasHeaders = true
				// 如果没有请求行但有头部，则设置inHeaders为true
				if !hasRequestLine {
					inHeaders = true
				}
			}
		}
	}

	// 去除请求体末尾的所有换行符
	req.Body = strings.TrimRight(req.Body, "\n")

	// 验证请求是否包含必要字段
	if req.Method == "" || req.URL == "" {
		// 如果有头部但没有请求行，则返回错误
		if hasHeaders && !hasRequestLine {
			return nil, fmt.Errorf("invalid http request: missing method or URL")
		}
		return nil, fmt.Errorf("invalid http request: missing method or URL")
	}

	return req, nil
}