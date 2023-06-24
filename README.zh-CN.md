# Greq

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/gookit/greq?style=flat-square)
[![GitHub tag (latest SemVer)](https://img.shields.io/github/tag/gookit/greq)](https://github.com/gookit/goutil)
[![GoDoc](https://godoc.org/github.com/gookit/greq?status.svg)](https://pkg.go.dev/github.com/gookit/greq)
[![Go Report Card](https://goreportcard.com/badge/github.com/gookit/greq)](https://goreportcard.com/report/github.com/gookit/greq)
[![Unit-Tests](https://github.com/gookit/greq/workflows/Unit-Tests/badge.svg)](https://github.com/gookit/greq/actions)
[![Coverage Status](https://coveralls.io/repos/github/gookit/greq/badge.svg?branch=main)](https://coveralls.io/github/gookit/greq?branch=main)

**greq** A simple http client request builder and sender

## 功能说明

- 链式配置请求，支持 `GET,POST,PUT,PATCH,DELETE,HEAD` 等通用请求方法
- 自定义提供请求Body
- 自定义响应Body解析
- 支持定义添加任意的中间件

## Install

```bash
go get github.com/gookit/greq
```

## Quick start

```go
package main

import (
	"github.com/gookit/goutil/dump"
	"github.com/gookit/greq"
)

func main() {
	resp, err := greq.New("https://httpbin.org").
		JSONType().
		UserAgent("custom-client/1.0").
		Get("/get")

	if err != nil {
		panic(err)
	}

	retMp := make(map[string]any)
	err = resp.Decode(&retMp)
	if err != nil {
		panic(err)
	}

	dump.P(retMp)
}
```

Result:

```text
PRINT AT github.com/gookit/greq_test.TestHReq_Send(greq_test.go:73)
map[string]interface {} { #len=4
  "args": map[string]interface {} { #len=0
  },
  "headers": map[string]interface {} { #len=4
    "Host": string("httpbin.org"), #len=11
    "User-Agent": string("custom-client/1.0"), #len=17
    "X-Amzn-Trace-Id": string("Root=1-61e4d41e-06e27ae12ff872a224373ca7"), #len=40
    "Accept-Encoding": string("gzip"), #len=4
  },
  "origin": string("222.210.59.218"), #len=14
  "url": string("https://httpbin.org/get"), #len=23
},
```

## Use middleware

```go
	buf := &bytes.Buffer{}
	mid0 := greq.MiddleFunc(func(r *http.Request, next greq.HandleFunc) (*greq.Response, error) {
		buf.WriteString("MID0>>")
		w, err := next(r)
		buf.WriteString(">>MID0")
		return w, err
	})

	mid1 := greq.MiddleFunc(func(r *http.Request, next greq.HandleFunc) (*greq.Response, error) {
		buf.WriteString("MID1>>")
		w, err := next(r)
		buf.WriteString(">>MID1")
		return w, err
	})

	mid2 := greq.MiddleFunc(func(r *http.Request, next greq.HandleFunc) (*greq.Response, error) {
		buf.WriteString("MID2>>")
		w, err := next(r)
		buf.WriteString(">>MID2")
		return w, err
	})

	resp, err := greq.New("https://httpbin.org").
		Doer(httpreq.DoerFunc(func(req *http.Request) (*http.Response, error) {
			tw := httptest.NewRecorder()
			buf.WriteString("(CORE)")
			return tw.Result(), nil
		})).
		Middleware(mid0, mid1, mid2).
		Get("/get")

    fmt.Println(buf.String())
```

**Output**:

```text
MID2>>MID1>>MID0>>(CORE)>>MID0>>MID1>>MID2
```

## More usage

### Response to string

`greq.Response.String()` can convert response to string.

```go
package main

import (
	"fmt"
	
	"github.com/gookit/goutil/dump"
	"github.com/gookit/greq"
)

func main() {
	resp, err := greq.New("https://httpbin.org").
		UserAgent("custom-client/1.0").
		Send("/get")
	
	if err != nil {
		panic(err)
	}

	fmt.Print(resp.String())
}
```

**Output**:

```text
HTTP/2.0 200 OK
Access-Control-Allow-Origin: *
Access-Control-Allow-Credentials: true
Date: Tue, 18 Jan 2022 04:52:39 GMT
Content-Type: application/json
Content-Length: 272
Server: gunicorn/19.9.0

{
  "args": {}, 
  "headers": {
    "Accept-Encoding": "gzip", 
    "Host": "httpbin.org", 
    "User-Agent": "custom-client/1.0", 
    "X-Amzn-Trace-Id": "Root=1-61e64797-3e428a925f7709906a8b7c01"
  }, 
  "origin": "222.210.59.218", 
  "url": "https://httpbin.org/get"
}
```

## Refers

- https://github.com/dghubble/sling
- https://github.com/zhshch2002/goreq
- https://github.com/go-resty/resty
- https://github.com/monaco-io/request

