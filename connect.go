package httpproxy

import (
	"io"
	"net"
	"net/http"
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

func (srv *Server) connect(w http.ResponseWriter, r *http.Request) {
	err := ErrHijackingNotSupported
	if hj, ok := w.(http.Hijacker); ok {
		var clientConn net.Conn
		if clientConn, _, err = hj.Hijack(); err == nil {
			host := r.URL.Host
			if r.URL.Port() == "" {
				switch r.URL.Scheme {
				case "http", "ws":
					host += ":80"
				case "https", "wss":
					host += ":443"
				}
			}
			var targetConn net.Conn
			if targetConn, err = srv.DialContext(r.Context(), "tcp", host); err == nil {
				if _, err = clientConn.Write([]byte("HTTP/1.0 200 Connection established\r\n\r\n")); err == nil {
					targetTCP, targetOK := targetConn.(halfClosable)
					clientTCP, clientOK := clientConn.(halfClosable)
					if targetOK && clientOK {
						go func() {
							defer clientTCP.Close()
							defer targetTCP.Close()
							var wg sync.WaitGroup
							wg.Add(2)
							go copyAndClose(targetTCP, clientTCP, &wg)
							go copyAndClose(clientTCP, targetTCP, &wg)
							wg.Wait()
						}()
					} else {
						go func() {
							defer clientConn.Close()
							defer targetConn.Close()
							ch := make(chan error, 2)
							go copyUntilClosed(ch, targetConn, clientConn)
							go copyUntilClosed(ch, clientConn, targetConn)
							<-ch
						}()
					}
				}
			}
		}
	}
	if err != nil && srv.Logger != nil {
		srv.Logger.Error("connect", "error", err)
	}
}
