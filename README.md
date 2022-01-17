# HReq

## Usage

```go
package main

import (
	"github.com/gookit/goutil/dump"
	"github.com/gookit/goutil/jsonutil"
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
	_ = jsonutil.DecodeReader(resp.Body, &retMp)
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

## Refers

- https://github.com/dghubble/sling
- https://github.com/zhshch2002/goreq

