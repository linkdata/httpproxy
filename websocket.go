package httpproxy

import (
	"errors"
	"io"
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

func copyUntilClosed(ch chan<- error, dst io.Writer, src io.Reader) {
	_, err := io.Copy(dst, src)
	if errors.Is(err, net.ErrClosed) {
		err = nil
	}
	ch <- err
}

func proxyWebsocket(remoteConn io.ReadWriter, proxyClient io.ReadWriter) error {
	ch := make(chan error, 2)
	go copyUntilClosed(ch, remoteConn, proxyClient)
	go copyUntilClosed(ch, proxyClient, remoteConn)
	return <-ch
}
