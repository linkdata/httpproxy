package httpproxy

import (
	"errors"
	"net"
	"net/http"
)

var ErrHijackingNotSupported = errors.New("http.ResponseWriter does not support http.Hijacker")

func hijack(w http.ResponseWriter) (conn net.Conn, err error) {
	err = ErrHijackingNotSupported
	if hj, ok := w.(http.Hijacker); ok {
		conn, _, err = hj.Hijack()
	}
	return
}
