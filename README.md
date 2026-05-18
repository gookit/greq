# greq

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/gookit/greq?style=flat-square)
[![GitHub tag (latest SemVer)](https://img.shields.io/github/tag/gookit/greq)](https://github.com/gookit/goutil)
[![GoDoc](https://pkg.go.dev/badge/github.com/gookit/greq.svg)](https://pkg.go.dev/github.com/gookit/greq)
[![Go Report Card](https://goreportcard.com/badge/github.com/gookit/greq)](https://goreportcard.com/report/github.com/gookit/greq)
[![Unit-Tests](https://github.com/gookit/greq/workflows/Unit-Tests/badge.svg)](https://github.com/gookit/greq/actions)
[![Coverage Status](https://coveralls.io/repos/github/gookit/greq/badge.svg?branch=main)](https://coveralls.io/github/gookit/greq?branch=main)

> [中文说明](README.zh-CN.md) | English

**greq** is a small, composable HTTP client for Go — a chainable request
builder, pluggable middleware, retry, batch, file upload/download, and
bundled CLI tools (`greq`, `gbench`).

## Features

- Chainable request builder for `GET / POST / PUT / PATCH / DELETE / HEAD`
- Pluggable middleware chain
- Pluggable body **providers** (raw, JSON, form, multipart) and response **decoders** (JSON, XML)
- Configurable **retry** with default checker (network errors, 5xx, 429) and per-request override
- **Batch** concurrent requests with `ExecuteAll` / `ExecuteAny` semantics (`ext/batch`)
- **Upload / download** helpers — single file, multi-file, multipart with form fields
- Parse and send **IDE `.http` file** request format directly (`ext/httpfile`)
- `BeforeSend` / `AfterSend` hooks and pluggable `Doer` for testing
- Bundled CLI tools:
  - [`cmd/greq`](cmd/greq) — curl-like HTTP client that understands `.http` files
  - [`cmd/gbench`](cmd/gbench) — `ab`-like benchmark tool with progress bar and graceful Ctrl+C

## Install

### Library

```bash
go get github.com/gookit/greq
```

### CLI tools

Install by [Eget](https://github.com/inherelab/eget):

```bash
eget install --asset "greq,gbench" gookit/greq
```

Install by Go:

```bash
# HTTP request tool
go install github.com/gookit/greq/cmd/greq@latest
# HTTP benchmark tool
go install github.com/gookit/greq/cmd/gbench@latest
```

## Quick start

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

For one-off calls there are package-level shortcuts that share a default client:

```go
resp, _ := greq.GetDo("https://httpbin.org/get")
_, _ = greq.PostDo("https://httpbin.org/post", `{"key":"val"}`)
```

## Building requests

`greq.New(baseURL)` returns a `*Client`. Calling any chainable method on
the client (`JSONType`, `UserAgent`, `Get`, …) returns a `*Builder` you
can keep configuring. The chain ends with a `*Do` method that actually
sends the request.

### Headers

Per-request (returns a `*Builder`):

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

Defaults applied to every request on a `*Client`:

```go
client := greq.New("https://api.example.com").
    DefaultUserAgent("my-client/1.0").
    DefaultHeader("X-Tenant", "acme")
```

### Content type and body

```go
// JSON
greq.New("https://api.example.com").
    JSONType().
    PostDo("/items", map[string]any{"name": "widget"})

// Form
greq.New("https://api.example.com").
    FormType().
    PostDo("/login", map[string]string{"user": "x", "pass": "y"})

// Raw bytes / reader / string (BytesBody installs a body Provider;
// pass nil as data so PostDo doesn't override it)
greq.New("https://api.example.com").
    WithContentType("application/octet-stream").
    BytesBody([]byte{0x01, 0x02}).
    PostDo("/upload-binary", nil)
```

Built-in content-type helpers: `JSONType()`, `FormType()`, `XMLType()`,
`MultipartType()`, `WithContentType(value)`.

### Query parameters

```go
greq.New("https://api.example.com").
    QueryParams(map[string]string{"page": "1", "size": "20"}).
    GetDo("/items")
```

### Per-request options

For per-call configuration without re-chaining, use `OptionFn` helpers:

```go
greq.GetDo("https://api.example.com/items",
    greq.WithHeader("X-Trace", "abc"),
    greq.WithTimeout(2000),          // 2s, in ms
    greq.WithMaxRetries(3),
)
```

Available: `WithMethod`, `WithContentType`, `WithUserAgent`, `WithHeader`,
`WithBody`, `WithData`, `WithTimeout`, `WithRetry`, `WithMaxRetries`,
`WithRetryDelay`, `WithRetryChecker`.

## Handling responses

```go
resp, err := greq.New("https://api.example.com").GetDo("/items")
if err != nil { return err }

if resp.IsFail() {                  // 4xx/5xx
    return fmt.Errorf("status %d", resp.StatusCode)
}

// Decode JSON into a struct (uses Client.RespDecoder, JSON by default)
var items []Item
if err := resp.Decode(&items); err != nil {
    return err
}
```

Response helpers:

- Status: `IsOK`, `IsSuccessful`, `IsFail`, `IsEmptyBody`
- Content type: `ContentType`, `IsContentType(prefix)`, `IsJSONType`
- Body: `Decode(ptr)`, `BodyString` / `BodyStringE`, `BodyBuffer` / `BodyBufferE`
- File output: `SaveFile(path)` (refuses non-2xx), `SaveBody(path)` (writes regardless)
- Inspection: `String()` (full HTTP response as text), `HeaderString()`
- Lifecycle: `CloseBody`, `QuietCloseBody`

> The plain `BodyBuffer` / `BodyString` panic on a read error; prefer the
> `…E` variants if you handle untrusted endpoints or care about
> resilience under load.

## Middleware

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

Middlewares execute in declaration order on the request and unwind in
reverse on the response.

## Retry

By default, no retries. Enable per-client:

```go
client := greq.New("https://api.example.com").
    WithMaxRetries(3).
    WithRetryDelay(200) // ms between attempts
```

`DefaultRetryChecker` retries on network errors, HTTP 5xx, and HTTP 429.

Per-request override:

```go
greq.GetDo("https://api.example.com/flaky",
    greq.WithRetry(5, 100, greq.DefaultRetryChecker),
)
```

Custom retry policy:

```go
onlyOn503 := func(resp *greq.Response, err error, attempt int) bool {
    return resp != nil && resp.StatusCode == 503
}
greq.New("https://api.example.com").
    WithRetryConfig(3, 200, onlyOn503).
    GetDo("/path")
```

## Upload / Download

```go
// Download to file (refuses non-2xx; use Response.SaveBody for unconditional)
client := greq.New("https://example.com")
n, err := client.Download("/file.zip", "./out.zip")

// Single file
resp, err := client.UploadFile("/upload", "file", "./photo.jpg")

// Multiple files
resp, err = client.UploadFiles("/upload", map[string]string{
    "image1": "./a.jpg",
    "image2": "./b.jpg",
})

// Files + extra form fields
resp, err = client.UploadWithData("/upload",
    map[string]string{"avatar": "./me.png"},
    map[string]string{"user_id": "42"},
)
```

See [docs/upload-download.md](docs/upload-download.md) for resumable
download, progress callbacks, and advanced multipart options.

## Batch requests (`ext/batch`)

Concurrent fan-out with a worker pool:

```go
import "github.com/gookit/greq/ext/batch"

// Wait for all
results := batch.GetAll([]string{
    "https://api.example.com/a",
    "https://api.example.com/b",
    "https://api.example.com/c",
})
for id, r := range results {
    fmt.Println(id, r.Response.StatusCode, r.Duration)
}

// Return the first success (cancels the rest)
winner := batch.GetAny([]string{
    "https://mirror1.example.com/file",
    "https://mirror2.example.com/file",
})

// Mixed methods, custom processor
bp := batch.NewProcessor(
    batch.WithMaxConcurrency(8),
    batch.WithBatchTimeout(10 * time.Second),
)
bp.AddGet("list", "https://api.example.com/list")
bp.AddPost("submit", "https://api.example.com/submit", map[string]string{"k": "v"})
all := bp.ExecuteAll()
```

If every request fails, `ExecuteAny` returns `nil` promptly rather than
waiting for the batch timeout.

## `.http` file format

`greq` can parse and send the IDE `.http` request format directly:

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

Variables use the `${name}` syntax. Unresolved variables fall back to
process environment variables, then are left as the literal name. See
`ext/httpfile` for direct access to the parser.

## Custom Doer / testing

`greq.Client.Doer(...)` swaps the underlying transport — useful for
mocking in tests:

```go
import "github.com/gookit/goutil/netutil/httpreq"

client := greq.New("https://api.example.com").
    Doer(httpreq.DoerFunc(func(req *http.Request) (*http.Response, error) {
        // return a recorded response without hitting the network
        return httptest.NewRecorder().Result(), nil
    }))
```

The same `BeforeSend` / `AfterSend` hooks are available on `Client` for
request signing, logging, and metric collection.

## Cloning a client

`Sub()` returns a shallow copy with its own headers map, suitable for
per-call customization without affecting the parent:

```go
base := greq.New("https://api.example.com").
    DefaultUserAgent("svc/1.0").
    WithMaxRetries(3)

sub := base.Sub().DefaultHeader("X-Trace", "abc-123")
sub.GetDo("/items")   // inherits retries, user-agent; adds trace header
```

## CLI tools

### `greq` — HTTP client

```bash
go install github.com/gookit/greq/cmd/greq@latest

greq https://httpbin.org/get
greq -X POST -d '{"name":"inhere"}' https://httpbin.org/post
greq -r req.http                          # send an .http file
greq -r req.http -V token=$API_TOKEN      # with variables
greq -O https://example.com/file.zip      # treat URL as download
```

Full flags: `greq -h`.

### `gbench` — benchmark tool

```bash
go install github.com/gookit/greq/cmd/gbench@latest

gbench -n 1000 -c 10 https://example.com
gbench -z 30s  -c 20 https://example.com           # run for 30 seconds
gbench -n 1000 -c 10 -m POST -d '{"k":"v"}' https://example.com/api
gbench -n 100  -c 5 -o results.txt https://example.com
```

`gbench` shows a live progress bar and honors Ctrl+C by stopping
gracefully and printing partial results.

## See also

- [dghubble/sling](https://github.com/dghubble/sling)
- [zhshch2002/goreq](https://github.com/zhshch2002/goreq)
- [go-resty/resty](https://github.com/go-resty/resty)
- [monaco-io/request](https://github.com/monaco-io/request)
