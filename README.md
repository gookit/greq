# HReq

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/gookit/hreq?style=flat-square)
[![GitHub tag (latest SemVer)](https://img.shields.io/github/tag/gookit/hreq)](https://github.com/gookit/goutil)
[![GoDoc](https://godoc.org/github.com/gookit/hreq?status.svg)](https://pkg.go.dev/github.com/gookit/hreq)
[![Go Report Card](https://goreportcard.com/badge/github.com/gookit/hreq)](https://goreportcard.com/report/github.com/gookit/hreq)
[![Unit-Tests](https://github.com/gookit/hreq/workflows/Unit-Tests/badge.svg)](https://github.com/gookit/hreq/actions)
[![Coverage Status](https://coveralls.io/repos/github/gookit/hreq/badge.svg?branch=main)](https://coveralls.io/github/gookit/hreq?branch=main)

**HReq** A simple http client request builder and sender

## Install

```bash
go get github.com/gookit/hreq
```

## Usage

```go
package main

import (
	"github.com/gookit/goutil/dump"
	"github.com/gookit/hreq"
)

func main() {
	resp, err := hreq.New("https://httpbin.org").
		JSONType().
		UserAgent("custom-client/1.0").
		Get("/get")

	if err != nil {
		panic(err)
	}

	retMp := make(map[string]interface{})
	err = resp.Decode(&retMp)
	if err != nil {
		panic(err)
	}

	dump.P(retMp)
}
```

Result:

```text
PRINT AT github.com/gookit/hreq_test.TestHReq_Send(hreq_test.go:73)
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

### Response to string

`hreq.Response.String()` can convert response to string.

```go
package main

import (
	"fmt"
	
	"github.com/gookit/goutil/dump"
	"github.com/gookit/hreq"
)

func main() {
	resp, err := hreq.New("https://httpbin.org").
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

