[![build](https://github.com/linkdata/httpproxy/actions/workflows/build.yml/badge.svg)](https://github.com/linkdata/httpproxy/actions/workflows/build.yml)
[![coverage](https://github.com/linkdata/httpproxy/blob/coverage/main/badge.svg)](https://htmlpreview.github.io/?https://github.com/linkdata/httpproxy/blob/coverage/main/report.html)
[![Docs](https://godoc.org/github.com/linkdata/httpproxy?status.svg)](https://godoc.org/github.com/linkdata/httpproxy)

# httpproxy

HTTP(S) and WebSocket forward proxy.

Supports user authentication, ContextDialer selection per proxy request and custom RoundTripper construction.

Only depends on the standard library. (Though the WebSocket tests use github.com/coder/websocket).

```go
package main

import (
	"fmt"
	"net/http"

	"github.com/linkdata/httpproxy"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/{$}", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "This is a web proxy server.")
	})
	// the httpproxy.Server will handle CONNECT and absolute URL requests,
	// and all others will be forwarded to the http.Handler we set.
	go http.ListenAndServe(":8080", &httpproxy.Server{Handler: mux})
}
```
