# greq

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/gookit/greq?style=flat-square)
[![GitHub tag (latest SemVer)](https://img.shields.io/github/tag/gookit/greq)](https://github.com/gookit/goutil)
[![GoDoc](https://pkg.go.dev/badge/github.com/gookit/greq.svg)](https://pkg.go.dev/github.com/gookit/greq)
[![Go Report Card](https://goreportcard.com/badge/github.com/gookit/greq)](https://goreportcard.com/report/github.com/gookit/greq)
[![Unit-Tests](https://github.com/gookit/greq/workflows/Unit-Tests/badge.svg)](https://github.com/gookit/greq/actions)
[![Coverage Status](https://coveralls.io/repos/github/gookit/greq/badge.svg?branch=main)](https://coveralls.io/github/gookit/greq?branch=main)

> 中文说明 | [English](README.md)

**greq** 是一个轻量、可组合的 Go HTTP 客户端：链式请求构建器、中间件、重试、批量、文件上传/下载，并自带 `greq` / `gbench` 两个 CLI 工具。

## 功能特性

- 链式请求构建器，支持 `GET / POST / PUT / PATCH / DELETE / HEAD`
- 可插拔中间件链
- 可插拔的请求体 **Provider**（raw / JSON / form / multipart）和响应 **Decoder**（JSON / XML）
- 可配置 **重试**，自带默认重试条件（网络错误、5xx、429），支持按请求覆盖
- **批量并发** 请求，提供 `ExecuteAll` / `ExecuteAny` 两种语义（`ext/batch`）
- 文件 **上传 / 下载**：单文件、多文件、多文件 + 表单字段
- 直接解析并发送 IDE **`.http` 文件** 格式请求（`ext/httpfile`）
- `BeforeSend` / `AfterSend` 钩子；可替换 `Doer` 便于 mock 测试
- 自带 CLI 工具：
  - [`cmd/greq`](cmd/greq) —— 类 curl 的 HTTP 请求工具，支持 `.http` 文件
  - [`cmd/gbench`](cmd/gbench) —— 类 `ab` 的压测工具，含进度条 + Ctrl+C 优雅停止

## 安装

### 作为库使用

```bash
go get github.com/gookit/greq
```

### 安装命令行工具

Install by [Eget](https://github.com/inherelab/eget):

```bash
eget install gookit/greq
eget install --name gbench gookit/greq
```

Install by Go:

```bash
# HTTP 请求工具
go install github.com/gookit/greq/cmd/greq@latest
# HTTP 压测工具
go install github.com/gookit/greq/cmd/gbench@latest
```

## 快速开始

```go
package main

import (
    "fmt"

    "github.com/gookit/greq"
)

func main() {
    resp, err := greq.New("https://httpbin.org").
        JSONType().
        UserAgent("custom-client/1.0").
        PostDo("/post", `{"name": "inhere"}`)
    if err != nil {
        panic(err)
    }

    var ret map[string]any
    if err := resp.Decode(&ret); err != nil {
        panic(err)
    }
    fmt.Printf("%+v\n", ret)
}
```

一次性调用可以直接用包级快捷函数（共享一个默认 client）：

```go
resp, _ := greq.GetDo("https://httpbin.org/get")
_, _ = greq.PostDo("https://httpbin.org/post", `{"key":"val"}`)
```

## 构建请求

`greq.New(baseURL)` 返回 `*Client`。在它上面调用任意可链式方法（`JSONType`、`UserAgent`、`Get` …）都会返回一个 `*Builder`，可以继续链式配置；末尾的 `*Do` 方法实际发送请求。

### 请求头

按请求设置（返回 `*Builder`）：

```go
greq.New("https://api.example.com").
    UserAgent("my-client/1.0").
    BasicAuth("user", "pass").
    SetHeader("X-Request-ID", "abc-123").
    SetHeaderMap(map[string]string{
        "X-Trace-Id": "t-1",
        "X-Tenant":   "acme",
    }).
    GetDo("/items")
```

应用到 `*Client` 上所有请求的默认头：

```go
client := greq.New("https://api.example.com").
    DefaultUserAgent("my-client/1.0").
    DefaultHeader("X-Tenant", "acme")
```

### Content-Type 和请求体

```go
// JSON
greq.New("https://api.example.com").
    JSONType().
    PostDo("/items", map[string]any{"name": "widget"})

// 表单
greq.New("https://api.example.com").
    FormType().
    PostDo("/login", map[string]string{"user": "x", "pass": "y"})

// 原始字节 / reader / 字符串（BytesBody 装的是 body Provider，
// 所以 PostDo 的 data 传 nil 即可，否则会被覆盖）
greq.New("https://api.example.com").
    WithContentType("application/octet-stream").
    BytesBody([]byte{0x01, 0x02}).
    PostDo("/upload-binary", nil)
```

内置 content-type 助手：`JSONType()`、`FormType()`、`XMLType()`、`MultipartType()`、`WithContentType(value)`。

### Query 参数

```go
greq.New("https://api.example.com").
    QueryParams(map[string]string{"page": "1", "size": "20"}).
    GetDo("/items")
```

### 单次请求选项

不想重新链式构建时，用 `OptionFn` 助手按请求覆盖：

```go
greq.GetDo("https://api.example.com/items",
    greq.WithHeader("X-Trace", "abc"),
    greq.WithTimeout(2000),          // 毫秒
    greq.WithMaxRetries(3),
)
```

可用：`WithMethod`、`WithContentType`、`WithUserAgent`、`WithHeader`、`WithBody`、`WithData`、`WithTimeout`、`WithRetry`、`WithMaxRetries`、`WithRetryDelay`、`WithRetryChecker`。

## 处理响应

```go
resp, err := greq.New("https://api.example.com").GetDo("/items")
if err != nil { return err }

if resp.IsFail() {                  // 4xx/5xx
    return fmt.Errorf("status %d", resp.StatusCode)
}

// 解码到结构体（用 Client.RespDecoder，默认 JSON）
var items []Item
if err := resp.Decode(&items); err != nil {
    return err
}
```

Response 上的方法：

- 状态：`IsOK`、`IsSuccessful`、`IsFail`、`IsEmptyBody`
- Content-Type：`ContentType`、`IsContentType(prefix)`、`IsJSONType`
- Body：`Decode(ptr)`、`BodyString` / `BodyStringE`、`BodyBuffer` / `BodyBufferE`
- 文件输出：`SaveFile(path)`（非 2xx 拒绝写入）、`SaveBody(path)`（无条件写）
- 调试：`String()`（整个 HTTP 响应文本形式）、`HeaderString()`
- 生命周期：`CloseBody`、`QuietCloseBody`

> 旧的 `BodyBuffer` / `BodyString` 在读取出错时会 panic；如果你处理不受信端点或在高负载下要求健壮性，请用 `…E` 变体。

## 中间件

```go
mid := greq.MiddleFunc(func(r *http.Request, next greq.HandleFunc) (*greq.Response, error) {
    start := time.Now()
    resp, err := next(r)
    log.Printf("%s %s -> %v in %s", r.Method, r.URL, err, time.Since(start))
    return resp, err
})

greq.New("https://api.example.com").
    Middleware(mid).
    GetDo("/items")
```

中间件按声明顺序在请求方向上执行，并按相反顺序在响应方向上展开。

## 重试

默认不重试。客户端级配置：

```go
client := greq.New("https://api.example.com").
    WithMaxRetries(3).
    WithRetryDelay(200) // 重试间隔，单位 ms
```

`DefaultRetryChecker` 在网络错误、HTTP 5xx 和 HTTP 429 时触发重试。

单次请求覆盖：

```go
greq.GetDo("https://api.example.com/flaky",
    greq.WithRetry(5, 100, greq.DefaultRetryChecker),
)
```

自定义策略：

```go
onlyOn503 := func(resp *greq.Response, err error, attempt int) bool {
    return resp != nil && resp.StatusCode == 503
}
greq.New("https://api.example.com").
    WithRetryConfig(3, 200, onlyOn503).
    GetDo("/path")
```

## 上传 / 下载

```go
// 下载到文件（非 2xx 拒绝写；想保留错误页面用 Response.SaveBody）
client := greq.New("https://example.com")
n, err := client.Download("/file.zip", "./out.zip")

// 单文件
resp, err := client.UploadFile("/upload", "file", "./photo.jpg")

// 多文件
resp, err = client.UploadFiles("/upload", map[string]string{
    "image1": "./a.jpg",
    "image2": "./b.jpg",
})

// 多文件 + 额外表单字段
resp, err = client.UploadWithData("/upload",
    map[string]string{"avatar": "./me.png"},
    map[string]string{"user_id": "42"},
)
```

更详细的断点续传、进度回调、高级 multipart 用法见 [docs/upload-download.md](docs/upload-download.md)。

## 批量并发请求（`ext/batch`）

带 worker 池的并发扇出：

```go
import "github.com/gookit/greq/ext/batch"

// 等所有完成
results := batch.GetAll([]string{
    "https://api.example.com/a",
    "https://api.example.com/b",
    "https://api.example.com/c",
})
for id, r := range results {
    fmt.Println(id, r.Response.StatusCode, r.Duration)
}

// 第一个成功就返回（其它会被取消）
winner := batch.GetAny([]string{
    "https://mirror1.example.com/file",
    "https://mirror2.example.com/file",
})

// 混合方法 + 自定义处理器
bp := batch.NewProcessor(
    batch.WithMaxConcurrency(8),
    batch.WithBatchTimeout(10 * time.Second),
)
bp.AddGet("list", "https://api.example.com/list")
bp.AddPost("submit", "https://api.example.com/submit", map[string]string{"k": "v"})
all := bp.ExecuteAll()
```

所有请求都失败时，`ExecuteAny` 会立即返回 `nil`，不会死等到批量 timeout。

## `.http` 文件格式

`greq` 可以直接解析并发送 IDE 的 `.http` 请求文件格式：

```go
raw := `POST https://api.example.com/items?tenant=${tenant}
Content-Type: application/json
Authorization: Bearer ${token}

{"name": "widget"}`

resp, err := greq.New().SendRaw(raw, map[string]string{
    "tenant": "acme",
    "token":  os.Getenv("API_TOKEN"),
})
```

变量语法为 `${name}`，未匹配的变量会回退到进程环境变量，再不行则保留字面量。直接访问解析器请用 `ext/httpfile`。

## 自定义 Doer / 测试

`greq.Client.Doer(...)` 替换底层 transport，方便 mock：

```go
import "github.com/gookit/goutil/netutil/httpreq"

client := greq.New("https://api.example.com").
    Doer(httpreq.DoerFunc(func(req *http.Request) (*http.Response, error) {
        // 返回预录制响应，不真实发起网络请求
        return httptest.NewRecorder().Result(), nil
    }))
```

`Client` 上同样有 `BeforeSend` / `AfterSend` 钩子，可以用于请求签名、日志、指标采集。

## 克隆 Client

`Sub()` 返回一份浅拷贝，自带独立的 header map，适合在父 client 基础上做单次个性化而不影响父级：

```go
base := greq.New("https://api.example.com").
    DefaultUserAgent("svc/1.0").
    WithMaxRetries(3)

sub := base.Sub().DefaultHeader("X-Trace", "abc-123")
sub.GetDo("/items")   // 继承重试、UA；额外加 X-Trace
```

## 命令行工具

### `greq` —— HTTP 客户端

```bash
go install github.com/gookit/greq/cmd/greq@latest

greq https://httpbin.org/get
greq -X POST -d '{"name":"inhere"}' https://httpbin.org/post
greq -r req.http                          # 发送 .http 文件
greq -r req.http -V token=$API_TOKEN      # 带变量
greq -O https://example.com/file.zip      # 当作下载链接
```

完整选项：`greq -h`。

### `gbench` —— 压测工具

```bash
go install github.com/gookit/greq/cmd/gbench@latest

gbench -n 1000 -c 10 https://example.com
gbench -z 30s  -c 20 https://example.com           # 跑 30 秒
gbench -n 1000 -c 10 -m POST -d '{"k":"v"}' https://example.com/api
gbench -n 100  -c 5 -o results.txt https://example.com
```

`gbench` 有实时进度条，Ctrl+C 会优雅停止并打印部分结果。

## 参考

- [dghubble/sling](https://github.com/dghubble/sling)
- [zhshch2002/goreq](https://github.com/zhshch2002/goreq)
- [go-resty/resty](https://github.com/go-resty/resty)
- [monaco-io/request](https://github.com/monaco-io/request)
