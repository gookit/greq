# AGENTS.md - Greq 项目开发指南

> 本文件为 AI 编程助手（如 Claude Code、Cursor 等）提供项目上下文和开发规范。

## 项目概述

**greq** 是一个简单、强大的 Go HTTP 客户端请求构建器和发送器，支持链式配置、中间件、重试等功能。

- **模块路径**: `github.com/gookit/greq`
- **Go 版本**: 1.21+
- **核心依赖**: `github.com/gookit/goutil`

### 主要功能

- 链式配置请求 (支持 GET/POST/PUT/PATCH/DELETE/HEAD)
- 中间件机制
- 请求重试 (支持自定义重试检查器)
- Body Provider 和 Response Decoder
- `.http` 文件格式解析
- 内置命令行工具: `greq` (HTTP 请求), `gbench` (压力测试)

## 构建与测试命令

### 运行所有测试
```bash
go test -v -cover ./...
```

### 运行单个测试
```bash
# 运行指定测试函数
go test -v -run TestFunctionName ./...

# 运行指定文件的测试
go test -v -run TestClient ./.

# 运行指定包的测试
go test -v ./ext/httpfile/...
```

### 生成覆盖率报告
```bash
go test -v -coverprofile="profile.cov" ./...
```

### 构建命令行工具
```bash
# 构建 greq 工具
go build -o bin/greq ./cmd/greq

# 构建 gbench 工具  
go build -o bin/gbench ./cmd/gbench
```

### 安装工具到 GOPATH
```bash
go install github.com/gookit/greq/cmd/greq@latest
go install github.com/gookit/greq/cmd/gbench@latest
```

## 项目结构

```
greq/
├── client.go          # 核心 Client 结构和方法
├── builder.go         # 请求构建器
├── option.go          # 请求选项配置
├── resp.go            # Response 包装器
├── resp_decoder.go    # 响应解码器
├── req_body.go        # 请求 Body Provider
├── middle.go          # 中间件机制
├── std.go             # 全局便捷函数
├── greq.go            # 公共类型和常量
├── cmd/
│   ├── greq/          # HTTP 请求命令行工具
│   └── gbench/        # 压力测试命令行工具
├── ext/
│   ├── httpfile/      # .http 文件解析
│   ├── batch/         # 批量请求处理
│   └── bench/         # 基准测试扩展
└── requtil/           # 请求工具函数
```

## 代码风格指南

### 导入顺序

按以下顺序分组导入，组间用空行分隔:
1. 标准库
2. 第三方库
3. 项目内部包

```go
import (
    // 标准库
    "context"
    "net/http"
    "time"
    
    // 第三方库
    "github.com/gookit/goutil/netutil/httpreq"
    
    // 内部包
    "github.com/gookit/greq/ext/httpfile"
)
```

### 命名约定

- **导出函数/类型**: PascalCase (如 `NewClient`, `BodyProvider`)
- **私有函数/字段**: camelCase (如 `sendRequest`, `doer`)
- **接口**: 通常以 `or`/`er` 结尾 (如 `Middleware`, `BodyProvider`, `RespDecoder`)
- **常量**: PascalCase 或全大写 (如 `HeaderUAgent`, `AgentCURL`)

### 链式调用模式

大部分配置方法返回 `*Client` 或 `*Builder` 以支持链式调用:

```go
// 正确示例
resp, err := greq.New("https://example.com").
    JSONType().
    UserAgent("my-client/1.0").
    WithMaxRetries(3).
    GetDo("/api/users")
```

### 函数式选项模式

使用 `OptionFn` 类型配置请求选项:

```go
// 定义选项函数
func WithTimeout(timeoutMs int) OptionFn {
    return func(opt *Options) {
        opt.Timeout = timeoutMs
    }
}

// 使用选项
client.GetDo("/path", greq.WithTimeout(5000))
```

### 错误处理

- 返回错误时使用 `error` 类型，不使用 panic (除非是 `Must*` 函数)
- 使用 `fmt.Errorf` 包装错误并添加上下文
- 在必要的地方添加 `defer` 关闭资源

```go
// 正确示例
if err != nil {
    return nil, fmt.Errorf("before send check failed: %w", err)
}
```

### 注释规范

- 导出函数/类型必须有注释说明
- 注释以函数名开头，使用英文
- 复杂逻辑添加行内注释

```go
// NewClient create a new http request client. alias of New()
func NewClient(baseURL ...string) *Client { return New(baseURL...) }

// IsOK check response status code is 200
func (r *Response) IsOK() bool {
    return httpreq.IsOK(r.StatusCode)
}
```

### 测试规范

- 测试文件命名: `*_test.go`
- 测试函数命名: `Test<FunctionName>` 或 `Test<Struct>_<Method>`
- 使用 `github.com/gookit/goutil/testutil/assert` 断言库
- 使用 `TestMain` 设置测试环境和 mock 服务器

```go
func TestClient_Send(t *testing.T) {
    resp, err := greq.New(testBaseURL).
        UserAgent("custom-cli/1.0").
        Send("GET", "/get")

    assert.NoErr(t, err)
    assert.True(t, resp.IsOK())
}
```

## 核心类型

### Client
HTTP 客户端，包含默认配置和中间件链。

### Builder
请求构建器，支持链式配置单个请求。

### Response
`http.Response` 包装器，提供便捷方法如 `IsOK()`, `Decode()`, `BodyString()`。

### Options
单个请求的配置选项。

### Middleware
中间件接口，用于拦截和处理请求。

### BodyProvider
请求体提供者接口，用于自定义请求体格式。

### RespDecoder
响应解码器接口，用于自定义响应解析。

## 常见任务示例

### 添加新的请求方法

在 `client.go` 中添加:
```go
// CustomDo sets the method to CUSTOM and sends request
func (h *Client) CustomDo(pathURL string, optFns ...OptionFn) (*Response, error) {
    return h.SendWithOpt(pathURL, NewOpt2(optFns, "CUSTOM"))
}
```

### 添加新的中间件

实现 `Middleware` 接口:
```go
type MyMiddleware struct{}

func (m *MyMiddleware) Handle(r *http.Request, next HandleFunc) (*Response, error) {
    // 前置处理
    resp, err := next(r)
    // 后置处理
    return resp, err
}
```

### 添加新的 OptionFn

在 `option.go` 中添加:
```go
func WithCustomOption(value string) OptionFn {
    return func(opt *Options) {
        opt.CustomField = value
    }
}
```

## Lint 工具

项目使用以下静态分析工具:
- **revive**: Go 代码质量检查
- **staticcheck**: 静态分析工具

CI 配置位于 `.github/workflows/go.yml`。

## 注意事项

1. 保持向后兼容性，避免破坏性变更
2. 新增公共 API 需要添加对应的测试
3. 链式方法必须返回正确的类型指针
4. 使用 `context.Context` 支持请求取消和超时
5. 资源关闭使用 `defer`，如 `defer resp.QuietCloseBody()`
