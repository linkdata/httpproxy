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
	var clientConn net.Conn
	if hj, ok := w.(http.Hijacker); ok {
		if clientConn, _, err = hj.Hijack(); err == nil {
			var cd ContextDialer
			var address string
			if cd, address, err = srv.getDialer(r); err == nil {
				var targetConn net.Conn
				if targetConn, err = cd.DialContext(r.Context(), "tcp", address); err == nil {
					if err = (fakeRoundTripper{}.WriteConnectResponse(clientConn)); err == nil {
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
						// successfully started proxying
						return
					}
					// hijacked ok, but writing connect response failed
					_ = targetConn.Close()
				}
			}
			// hijacked ok, but dial or writing connect response failed
			_ = (fakeRoundTripper{err}.WriteConnectResponse(clientConn))
			_ = clientConn.Close()
		}
	}
	if clientConn == nil {
		// w was not a http.Hijacker or hijack failed
		w.WriteHeader(http.StatusInternalServerError)
	}
	if err != nil && srv.Logger != nil {
		srv.Logger.Error("connect", "error", err)
	}
}
