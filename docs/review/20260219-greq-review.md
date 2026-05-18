# Greq 代码审核报告

> 审核日期: 2026-02-17
> 审核范围: 连接池管理、Timeout 机制、Client 配置

## 问题摘要

| 严重程度 | 问题 | 位置 |
|---------|------|------|
| 🔴 严重 | Timeout 双重机制冲突 | `client.go` New(), DefaultTimeout(), NewRequestWithOptions() |
| 🔴 严重 | Sub() 方法未复制关键字段 | `client.go` Sub() |
| 🟡 中等 | 连接池配置不一致 | `client.go` New() vs `greq.go` NewTransport() |
| 🟡 中等 | SetMaxIdleConns 可能 panic | `client.go` SetMaxIdleConns() |

---

## 问题详情

### 1. Timeout 双重机制冲突 🔴

**问题描述**：代码中存在两套独立的 timeout 机制，同时生效可能产生不可预期的行为。

**当前实现**：

```go
// 机制 1: http.Client.Timeout (在 New() 中)
doer: &http.Client{
    Timeout: 10 * time.Second,  // 整个请求超时
}

// 机制 2: context.WithTimeout (在 NewRequestWithOptions() 中)
if opt.Timeout > 0 {
    ctx, opt.TCancelFn = context.WithTimeout(ctx, time.Duration(opt.Timeout)*time.Millisecond)
}
```

**问题分析**：

| 机制 | 作用范围 | 行为 |
|------|---------|------|
| `http.Client.Timeout` | 整个请求（含重定向、所有重试） | 超时后关闭连接 |
| `context.WithTimeout` | 单个请求 | 超时后取消请求 |

**潜在问题**：
1. 两者同时生效，timeout 值不同时行为不确定
2. `DefaultTimeout()` 同时设置两者，但语义不清晰
3. 用户难以预测哪个 timeout 会先触发

**建议**：统一使用 `context.WithTimeout`，移除 `http.Client.Timeout`。

---

### 2. Sub() 方法未复制关键字段 🔴

**问题描述**：`Sub()` 方法创建子 Client 时，丢失了多个关键配置字段。

**当前实现**：

```go
func (h *Client) Sub() *Client {
    headerCopy := make(http.Header)
    for k, v := range h.Header {
        headerCopy[k] = v
    }

    return &Client{
        doer:        h.doer,
        Method:      h.Method,
        BaseURL:     h.BaseURL,
        Header:      headerCopy,
        RespDecoder: h.RespDecoder,
        // 缺少以下字段!
    }
}
```

**缺失字段**：

| 字段 | 影响 |
|------|------|
| `Timeout` | 子 Client 的 timeout = 0，导致 context timeout 不生效 |
| `MaxRetries` | 重试功能丢失 |
| `RetryDelay` | 重试延迟丢失 |
| `RetryChecker` | 自定义重试检查器丢失 |
| `BeforeSend` | 请求前回调丢失 |
| `AfterSend` | 请求后回调丢失 |
| `ContentType` | 默认内容类型丢失 |

**后果**：子 Client 的行为与父 Client 不一致，可能导致：
- 请求超时机制失效
- 重试功能失效
- 回调函数不执行

---

### 3. 连接池配置不一致 🟡

**问题描述**：`New()` 和 `NewTransport()` 使用完全不同的 Transport 配置。

**配置对比**：

| 配置项 | `New()` Transport | `NewTransport()` | 说明 |
|--------|-------------------|------------------|------|
| DialContext | ❌ 默认值 | ✅ 30s+30s | 缺少连接建立超时 |
| MaxIdleConns | 200 | 500 | 最大空闲连接数 |
| MaxConnsPerHost | ❌ 未设置 | 200 | 每主机最大连接 |
| MaxIdleConnsPerHost | 50 | 100 | 每主机空闲连接 |
| IdleConnTimeout | 90s | 60s | 空闲连接超时 |
| TLSHandshakeTimeout | ❌ 默认值 | 1s | TLS 握手超时 |
| ExpectContinueTimeout | ❌ 默认值 | 1s | 100-continue 超时 |
| InsecureSkipVerify | ❌ false | ⚠️ **true** | 安全风险! |

**问题**：
1. `New()` 缺少 `DialContext`，连接建立无超时限制
2. `NewTransport()` 的 `InsecureSkipVerify: true` 是安全风险
3. 两套配置参数不一致，用户难以理解

---

### 4. SetMaxIdleConns 可能 panic 🟡

**问题描述**：方法中存在不安全的类型断言。

**当前实现**：

```go
func (h *Client) SetMaxIdleConns(maxIdleConns, maxIdleConnsPerHost int) *Client {
    if hc, ok := h.doer.(*http.Client); ok {
        transport := hc.Transport.(*http.Transport)  // ⚠️ 危险!
        transport.MaxIdleConns = maxIdleConns
        transport.MaxIdleConnsPerHost = maxIdleConnsPerHost
    }
    return h
}
```

**问题**：
1. `hc.Transport` 可能是 `nil`
2. `hc.Transport` 可能不是 `*http.Transport` 类型

**触发场景**：
- 用户通过 `Doer()` 设置了自定义的 Doer
- 用户通过 `HttpClient()` 设置了没有 Transport 的 http.Client

---

## Timeout 对连接池的影响分析

### 直接影响：无

修改 `http.Client.Timeout` **不会**直接改变连接池的配置参数：
- `MaxIdleConns`
- `MaxIdleConnsPerHost`
- `IdleConnTimeout`
- `MaxConnsPerHost`

### 间接影响：有

| 场景 | 对连接池的影响 |
|------|---------------|
| Timeout 过短 | 请求被中断，连接可能无法正常归还连接池，连接状态异常 |
| Context 取消 | 底层连接会被关闭，无法复用 |
| 频繁超时 | 可能导致连接池中的连接不稳定，需要重新建立 |

### 关键点

```
http.Client.Timeout: 影响请求生命周期，超时后连接可能被强制关闭
IdleConnTimeout:     影响空闲连接在池中存活的时间，与请求 timeout 独立
```

**建议**：
- 使用 `context.WithTimeout` 控制单个请求超时
- 不要设置 `http.Client.Timeout`（设为 0）
- `IdleConnTimeout` 保持合理值（如 90s）

---

## 修复建议

### 修复 1: 统一 Timeout 机制

```go
// New() - 不设置 http.Client.Timeout
func New(baseURL ...string) *Client {
    h := &Client{
        doer: &http.Client{
            // 不设置 Timeout，由 context 控制
            Transport: &http.Transport{...},
        },
        Timeout: 10000,  // 用于 context timeout
        // ...
    }
    return h
}

// DefaultTimeout() - 只设置 Client.Timeout，不修改 http.Client
func (h *Client) DefaultTimeout(timeoutMs int) *Client {
    h.Timeout = timeoutMs
    return h
}
```

### 修复 2: Sub() 完整复制

```go
func (h *Client) Sub() *Client {
    headerCopy := make(http.Header)
    for k, v := range h.Header {
        headerCopy[k] = v
    }
    return &Client{
        doer:         h.doer,
        Method:       h.Method,
        BaseURL:      h.BaseURL,
        Header:       headerCopy,
        ContentType:  h.ContentType,
        Timeout:      h.Timeout,
        RespDecoder:  h.RespDecoder,
        MaxRetries:   h.MaxRetries,
        RetryDelay:   h.RetryDelay,
        RetryChecker: h.RetryChecker,
        BeforeSend:   h.BeforeSend,
        AfterSend:    h.AfterSend,
    }
}
```

### 修复 3: 完善 Transport 配置

```go
func New(baseURL ...string) *Client {
    h := &Client{
        doer: &http.Client{
            Transport: &http.Transport{
                DialContext: (&net.Dialer{
                    Timeout:   30 * time.Second,
                    KeepAlive: 30 * time.Second,
                }).DialContext,
                MaxIdleConns:          200,
                MaxIdleConnsPerHost:   50,
                IdleConnTimeout:       90 * time.Second,
                TLSHandshakeTimeout:   10 * time.Second,
                ExpectContinueTimeout: 1 * time.Second,
            },
        },
        Timeout: 10000,
        // ...
    }
    return h
}
```

### 修复 4: 安全的类型断言

```go
func (h *Client) SetMaxIdleConns(maxIdleConns, maxIdleConnsPerHost int) *Client {
    if hc, ok := h.doer.(*http.Client); ok {
        if transport, ok := hc.Transport.(*http.Transport); ok && transport != nil {
            transport.MaxIdleConns = maxIdleConns
            transport.MaxIdleConnsPerHost = maxIdleConnsPerHost
        }
    }
    return h
}
```

---

## 测试建议

修复后应添加以下测试用例：

1. **Timeout 机制测试**
   - 验证 context timeout 正确触发
   - 验证超时后连接状态正确

2. **Sub() 继承测试**
   - 验证子 Client 继承所有父 Client 配置
   - 验证子 Client 修改不影响父 Client

3. **连接池配置测试**
   - 验证连接复用正常工作
   - 验证并发请求下连接池行为

---

## 参考链接

- [Go http.Client Timeout 机制](https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts/)
- [Go 连接池最佳实践](https://golang.org/src/net/http/transport.go)
