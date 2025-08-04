package httpproxy

import (
	"errors"
	"io"
	"net"
)

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
