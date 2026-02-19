# 文件上传与下载

greq 提供了简洁的文件上传和下载功能，支持单文件、多文件上传以及带表单字段的上传。

## 目录

- [文件下载](#文件下载)
- [文件上传](#文件上传)
  - [单文件上传](#单文件上传)
  - [多文件上传](#多文件上传)
  - [带表单字段上传](#带表单字段上传)
- [高级选项](#高级选项)
- [错误处理](#错误处理)

---

## 文件下载

### 方法签名

```go
func (h *Client) Download(url, savePath string, optFns ...OptionFn) (int, error)
```

### 参数说明

| 参数 | 类型 | 说明 |
|------|------|------|
| `url` | `string` | 文件下载地址（完整 URL） |
| `savePath` | `string` | 本地保存路径 |
| `optFns` | `...OptionFn` | 可选配置项 |

### 返回值

| 返回值 | 类型 | 说明 |
|--------|------|------|
| `n` | `int` | 写入的字节数 |
| `err` | `error` | 错误信息 |

### 基本用法

```go
package main

import (
    "fmt"
    "github.com/gookit/greq"
)

func main() {
    client := greq.New()

    // 下载文件到本地
    n, err := client.Download(
        "https://example.com/files/document.pdf",
        "/path/to/save/document.pdf",
    )
    if err != nil {
        panic(err)
    }

    fmt.Printf("下载完成，共 %d 字节\n", n)
}
```

### 带超时设置

```go
n, err := greq.New().Download(
    "https://example.com/large-file.zip",
    "/path/to/save/large-file.zip",
    greq.WithTimeout(60000), // 60秒超时
)
```

### 带请求头

```go
n, err := greq.New().Download(
    "https://example.com/protected-file.pdf",
    "/path/to/save/file.pdf",
    greq.WithHeader("Authorization", "Bearer token123"),
)
```

---

## 文件上传

greq 使用 `multipart/form-data` 格式上传文件，自动处理 boundary 和 Content-Type。

### 单文件上传

```go
func (h *Client) UploadFile(pathURL, fieldName, filePath string, optFns ...OptionFn) (*Response, error)
```

#### 参数说明

| 参数 | 类型 | 说明 |
|------|------|------|
| `pathURL` | `string` | 上传接口路径 |
| `fieldName` | `string` | 表单字段名（服务端接收的字段名） |
| `filePath` | `string` | 本地文件路径 |
| `optFns` | `...OptionFn` | 可选配置项 |

#### 示例

```go
package main

import (
    "fmt"
    "github.com/gookit/greq"
)

func main() {
    client := greq.New("https://api.example.com")

    // 上传单个文件
    resp, err := client.UploadFile(
        "/api/upload",      // 上传接口
        "file",             // 表单字段名
        "/path/to/file.txt", // 本地文件路径
    )
    if err != nil {
        panic(err)
    }

    fmt.Printf("上传成功: %s\n", resp.BodyString())
}
```

### 多文件上传

```go
func (h *Client) UploadFiles(pathURL string, files map[string]string, optFns ...OptionFn) (*Response, error)
```

#### 参数说明

| 参数 | 类型 | 说明 |
|------|------|------|
| `pathURL` | `string` | 上传接口路径 |
| `files` | `map[string]string` | 文件映射：字段名 -> 文件路径 |
| `optFns` | `...OptionFn` | 可选配置项 |

#### 示例

```go
client := greq.New("https://api.example.com")

// 上传多个文件
resp, err := client.UploadFiles("/api/upload", map[string]string{
    "avatar":    "/path/to/avatar.jpg",
    "document":  "/path/to/document.pdf",
    "thumbnail": "/path/to/thumb.png",
})
if err != nil {
    panic(err)
}

fmt.Printf("上传成功: %s\n", resp.BodyString())
```

### 带表单字段上传

```go
func (h *Client) UploadWithData(pathURL string, files map[string]string, fields map[string]string, optFns ...OptionFn) (*Response, error)
```

#### 参数说明

| 参数 | 类型 | 说明 |
|------|------|------|
| `pathURL` | `string` | 上传接口路径 |
| `files` | `map[string]string` | 文件映射：字段名 -> 文件路径 |
| `fields` | `map[string]string` | 表单字段：字段名 -> 值 |
| `optFns` | `...OptionFn` | 可选配置项 |

#### 示例

```go
client := greq.New("https://api.example.com")

// 上传文件并附带表单数据
resp, err := client.UploadWithData(
    "/api/upload",
    map[string]string{
        "document": "/path/to/report.pdf",
    },
    map[string]string{
        "title":  "年度报告",
        "author": "张三",
        "year":   "2024",
    },
)
if err != nil {
    panic(err)
}

fmt.Printf("上传成功: %s\n", resp.BodyString())
```

---

## 高级选项

上传下载方法支持所有标准 `OptionFn` 配置项：

### 超时设置

```go
resp, err := client.UploadFile("/upload", "file", "/path/to/file",
    greq.WithTimeout(30000), // 30秒超时
)
```

### 自定义请求头

```go
resp, err := client.UploadFile("/upload", "file", "/path/to/file",
    greq.WithHeader("Authorization", "Bearer your-token"),
    greq.WithHeader("X-Request-ID", "req-12345"),
)
```

### 带 Basic Auth

```go
client := greq.New("https://api.example.com").
    BasicAuth("username", "password")

resp, err := client.UploadFile("/upload", "file", "/path/to/file")
```

### 重试机制

```go
resp, err := client.UploadFile("/upload", "file", "/path/to/file",
    greq.WithMaxRetries(3),
    greq.WithRetryDelay(1000), // 1秒延迟
)
```

---

## 错误处理

### 下载错误

```go
n, err := client.Download(url, savePath)
if err != nil {
    if strings.Contains(err.Error(), "Download failed") {
        // 服务器返回错误状态码
        fmt.Println("文件不存在或无权限")
    } else {
        // 网络或其他错误
        fmt.Printf("下载失败: %v\n", err)
    }
}
```

### 上传错误

```go
resp, err := client.UploadFile("/upload", "file", filePath)
if err != nil {
    // 网络错误或文件读取错误
    fmt.Printf("上传失败: %v\n", err)
    return
}

if !resp.IsOK() {
    // 服务器返回非 200 状态码
    fmt.Printf("服务器错误: %d - %s\n", resp.StatusCode, resp.BodyString())
    return
}

// 检查响应内容
var result struct {
    Success bool   `json:"success"`
    Message string `json:"message"`
}
if err := resp.Decode(&result); err != nil {
    fmt.Printf("解析响应失败: %v\n", err)
    return
}

if !result.Success {
    fmt.Printf("上传被拒绝: %s\n", result.Message)
}
```

---

## 完整示例

### 图片上传服务

```go
package main

import (
    "fmt"
    "github.com/gookit/greq"
)

type UploadResult struct {
    Success  bool   `json:"success"`
    URL      string `json:"url"`
    Filename string `json:"filename"`
    Size     int64  `json:"size"`
}

func main() {
    client := greq.New("https://api.example.com").
        WithTimeout(30000).
        WithMaxRetries(2)

    resp, err := client.UploadWithData(
        "/api/images/upload",
        map[string]string{
            "image": "/path/to/photo.jpg",
        },
        map[string]string{
            "album":    "vacation",
            "isPublic": "true",
        },
    )

    if err != nil {
        panic(err)
    }

    var result UploadResult
    if err := resp.Decode(&result); err != nil {
        panic(err)
    }

    if result.Success {
        fmt.Printf("图片上传成功!\n")
        fmt.Printf("访问地址: %s\n", result.URL)
        fmt.Printf("文件名: %s\n", result.Filename)
        fmt.Printf("大小: %d bytes\n", result.Size)
    }
}
```

### 批量下载工具

```go
package main

import (
    "fmt"
    "os"
    "path/filepath"
    "github.com/gookit/greq"
)

func main() {
    client := greq.New().WithTimeout(60000)

    files := []struct {
        url      string
        filename string
    }{
        {"https://example.com/file1.pdf", "report-2024.pdf"},
        {"https://example.com/file2.zip", "backup.zip"},
        {"https://example.com/file3.jpg", "logo.jpg"},
    }

    saveDir := "./downloads"
    os.MkdirAll(saveDir, 0755)

    for _, f := range files {
        savePath := filepath.Join(saveDir, f.filename)
        n, err := client.Download(f.url, savePath)
        if err != nil {
            fmt.Printf("下载失败 %s: %v\n", f.filename, err)
            continue
        }
        fmt.Printf("下载完成 %s (%d bytes)\n", f.filename, n)
    }
}
```

---

## API 参考

### Client 方法

| 方法 | 说明 |
|------|------|
| `Download(url, savePath, opts...)` | 下载文件到本地 |
| `UploadFile(pathURL, fieldName, filePath, opts...)` | 上传单个文件 |
| `UploadFiles(pathURL, files, opts...)` | 上传多个文件 |
| `UploadWithData(pathURL, files, fields, opts...)` | 上传文件和表单字段 |

### OptionFn 选项

| 选项 | 说明 |
|------|------|
| `WithTimeout(ms)` | 设置超时时间（毫秒） |
| `WithHeader(key, value)` | 设置请求头 |
| `WithMaxRetries(n)` | 设置最大重试次数 |
| `WithRetryDelay(ms)` | 设置重试延迟（毫秒） |

### Response 方法

| 方法 | 说明 |
|------|------|
| `IsOK()` | 状态码是否为 200 |
| `IsSuccessful()` | 状态码是否在 200-299 范围 |
| `StatusCode` | HTTP 状态码 |
| `BodyString()` | 获取响应体字符串 |
| `Decode(ptr)` | 解码 JSON 响应 |
| `SaveFile(path)` | 保存响应体到文件 |
