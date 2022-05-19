# HReq

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/gookit/hireq?style=flat-square)
[![GitHub tag (latest SemVer)](https://img.shields.io/github/tag/gookit/hireq)](https://github.com/gookit/goutil)
[![GoDoc](https://godoc.org/github.com/gookit/hireq?status.svg)](https://pkg.go.dev/github.com/gookit/hireq)
[![Go Report Card](https://goreportcard.com/badge/github.com/gookit/hireq)](https://goreportcard.com/report/github.com/gookit/hireq)
[![Unit-Tests](https://github.com/gookit/hireq/workflows/Unit-Tests/badge.svg)](https://github.com/gookit/hireq/actions)
[![Coverage Status](https://coveralls.io/repos/github/gookit/hireq/badge.svg?branch=main)](https://coveralls.io/github/gookit/hireq?branch=main)

**HReq** A simple http client request builder and sender

> `hireq` inspired from [dghubble/sling][1] and more projects, please see refers.

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
go get github.com/gookit/hireq
```

## Quick start

```go
package main

import (
	"github.com/gookit/goutil/dump"
	"github.com/gookit/hireq"
)

func main() {
	resp, err := hireq.New("https://httpbin.org").
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
PRINT AT github.com/gookit/hireq_test.TestHReq_Send(hireq_test.go:73)
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

## Request headers

```go
hireq.New("some.host/api").
	SetHeader("req-id", "a string")
```

Set multi at once:

```go
hireq.New("some.host/api").
	SetHeaders(map[string]string{
		"req-id": "a string",
	})
```

### Set content type

```go
hireq.New("some.host/api").
    ContentType("text/html")
```

Built in `JSONType()` `FromType()` `XMLType()`

```go
hireq.New("some.host/api").JSONType()
```


## Use middleware

```go
	buf := &bytes.Buffer{}
	mid0 := hireq.MiddleFunc(func(r *http.Request, next hireq.HandleFunc) (*hireq.Response, error) {
		buf.WriteString("MID0>>")
		w, err := next(r)
		buf.WriteString(">>MID0")
		return w, err
	})

	mid1 := hireq.MiddleFunc(func(r *http.Request, next hireq.HandleFunc) (*hireq.Response, error) {
		buf.WriteString("MID1>>")
		w, err := next(r)
		buf.WriteString(">>MID1")
		return w, err
	})

	mid2 := hireq.MiddleFunc(func(r *http.Request, next hireq.HandleFunc) (*hireq.Response, error) {
		buf.WriteString("MID2>>")
		w, err := next(r)
		buf.WriteString(">>MID2")
		return w, err
	})

	resp, err := hireq.New("https://httpbin.org").
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
    str := hireq.New("https://httpbin.org").
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

`hireq.Response.String()` can convert response to string.

```go
package main

import (
	"fmt"
	
	"github.com/gookit/goutil/dump"
	"github.com/gookit/hireq"
)

func main() {
	resp, err := hireq.New("https://httpbin.org").
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

[1]: https://github.com/dghubble/sling