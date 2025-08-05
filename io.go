package httpproxy

import (
	"errors"
	"io"
	"net"
	"sync"
)

type halfClosable interface {
	net.Conn
	CloseWrite() error
	CloseRead() error
}

var _ halfClosable = (*net.TCPConn)(nil)

func copyAndClose(dst, src halfClosable, wg *sync.WaitGroup) {
	defer wg.Done()
	_, _ = io.Copy(dst, src)
	_ = dst.CloseWrite()
	_ = src.CloseRead()
}

func copyUntilClosed(ch chan<- error, dst io.Writer, src io.Reader) {
	_, err := io.Copy(dst, src)
	if errors.Is(err, net.ErrClosed) {
		err = nil
	}
	ch <- err
}

func proxyUntilClosed(targetConn io.ReadWriter, clientConn io.ReadWriter) error {
	ch := make(chan error, 2)
	go copyUntilClosed(ch, targetConn, clientConn)
	go copyUntilClosed(ch, clientConn, targetConn)
	return <-ch
}
