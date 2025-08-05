[![build](https://github.com/linkdata/httpproxy/actions/workflows/build.yml/badge.svg)](https://github.com/linkdata/httpproxy/actions/workflows/build.yml)
[![coverage](https://coveralls.io/repos/github/linkdata/httpproxy/badge.svg?branch=main)](https://coveralls.io/github/linkdata/httpproxy?branch=main)
[![goreport](https://goreportcard.com/badge/github.com/linkdata/httpproxy)](https://goreportcard.com/report/github.com/linkdata/httpproxy)
[![Docs](https://godoc.org/github.com/linkdata/httpproxy?status.svg)](https://godoc.org/github.com/linkdata/httpproxy)

# httpproxy

HTTP(S) and WebSocket forward proxy.

Supports user authentication, ContextDialer selection per proxy request and custom RoundTripper construction.

Only depends on the standard library. (Though the WebSocket tests use github.com/coder/websocket).
