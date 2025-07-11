package httpproxy

import (
	"io"
	"net"
	"net/http"
	"net/url"
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

func getAddress(u *url.URL) (address string) {
	address = u.Host
	if u.Port() == "" {
		switch u.Scheme {
		case "http", "ws":
			address += ":80"
		case "https", "wss":
			address += ":443"
		}
	}
	return
}

func (srv *Server) connect(w http.ResponseWriter, r *http.Request) {
	err := ErrHijackingNotSupported
	if hj, ok := w.(http.Hijacker); ok {
		var clientConn net.Conn
		if clientConn, _, err = hj.Hijack(); err == nil {
			var targetConn net.Conn
			var cd ContextDialer
			var address string
			if cd, address, err = srv.getDialer(r); err == nil {
				if targetConn, err = cd.DialContext(r.Context(), "tcp", address); err == nil {
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
						return
					}
				}
			}
			// hijacked ok, but dial or write failed
			_, _ = clientConn.Write([]byte("HTTP/1.0 502 Bad Gateway\r\n\r\n"))
			clientConn.Close()
			return
		}
	}
	// w was not a http.Hijacker or hijack failed
	w.WriteHeader(http.StatusInternalServerError)
	if err != nil && srv.Logger != nil {
		srv.Logger.Error("connect", "error", err)
	}
}
