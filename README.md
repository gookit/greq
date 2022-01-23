# HReq

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/gookit/hreq?style=flat-square)
[![GitHub tag (latest SemVer)](https://img.shields.io/github/tag/gookit/hreq)](https://github.com/gookit/goutil)
[![GoDoc](https://godoc.org/github.com/gookit/hreq?status.svg)](https://pkg.go.dev/github.com/gookit/hreq)
[![Go Report Card](https://goreportcard.com/badge/github.com/gookit/hreq)](https://goreportcard.com/report/github.com/gookit/hreq)
[![Unit-Tests](https://github.com/gookit/hreq/workflows/Unit-Tests/badge.svg)](https://github.com/gookit/hreq/actions)
[![Coverage Status](https://coveralls.io/repos/github/gookit/hreq/badge.svg?branch=main)](https://coveralls.io/github/gookit/hreq?branch=main)

**HReq** A simple http client request builder and sender

> `hreq` inspired from https://github.com/dghubble/sling and more projects, please see refers.

## Features

- Make http requests, supports `GET,POST,PUT,PATCH,DELETE,HEAD`
- Transform request and response data
- Supports chain configuration request
- Supports defining and adding middleware
- Supports defining request body provider and response decoder
- Built-In: fom, json request body provider
- Built-In: xml, json response body decoder

## Install

```bash
go get github.com/gookit/hreq
```

## Quick start

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
		PostDo("/post")

	if err != nil {
		panic(err)
	}

	ret := make(map[string]interface{})
	err = resp.Decode(&ret)
	if err != nil {
		panic(err)
	}

	dump.P(ret)
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
  "url": string("https://httpbin.org/post"), #len=24
},
```

## Use middleware

```go
	buf := &bytes.Buffer{}
	mid0 := hreq.MiddleFunc(func(r *http.Request, next hreq.HandleFunc) (*hreq.Response, error) {
		buf.WriteString("MID0>>")
		w, err := next(r)
		buf.WriteString(">>MID0")
		return w, err
	})

	mid1 := hreq.MiddleFunc(func(r *http.Request, next hreq.HandleFunc) (*hreq.Response, error) {
		buf.WriteString("MID1>>")
		w, err := next(r)
		buf.WriteString(">>MID1")
		return w, err
	})

	mid2 := hreq.MiddleFunc(func(r *http.Request, next hreq.HandleFunc) (*hreq.Response, error) {
		buf.WriteString("MID2>>")
		w, err := next(r)
		buf.WriteString(">>MID2")
		return w, err
	})

	resp, err := hreq.New("https://httpbin.org").
		Doer(httpreq.DoerFunc(func(req *http.Request) (*http.Response, error) {
			tw := httptest.NewRecorder()
			buf.WriteString("(CORE)")
			return tw.Result(), nil
		})).
		Middleware(mid0, mid1, mid2).
		GetDo("/get")

    fmt.Println(buf.String())
```

**Output**:

```text
MID2>>MID1>>MID0>>(CORE)>>MID0>>MID1>>MID2
```

## More usage

### Check response

- `Response.IsOK() bool`
- `Response.IsFail() bool`
- `Response.IsEmptyBody() bool`
- `Response.IsJSONType() bool`
- `Response.IsContentType(prefix string) bool`

### Get response data

- `Response.ContentType() string`
- `Response.Decode(ptr interface{}) error`

### Request to string

```go
    str := hreq.New("https://httpbin.org").
		UserAgent("some-client/1.0").
		BasicAuth("inhere", "some string").
		JSONType().
		Body("hi, with body").
		Post("/post").
		String()

	fmt.Println(str)
```

**Output**:

```text
POST https://httpbin.org/post/ HTTP/1.1
User-Agent: some-client/1.0
Authorization: Basic aW5oZXJlOnNvbWUgc3RyaW5n
Content-Type: application/json; charset=utf-8

hi, with body
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

