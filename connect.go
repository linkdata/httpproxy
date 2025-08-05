package httpproxy

import (
	"net"
	"net/http"
	"sync"
)

func (srv *Server) connect(w http.ResponseWriter, r *http.Request) {
	var err error
	var clientConn net.Conn
	if clientConn, err = hijack(w); err == nil {
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
							_ = proxyUntilClosed(targetConn, clientConn)
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
	if clientConn == nil {
		// w was not a http.Hijacker or hijack failed
		w.WriteHeader(http.StatusInternalServerError)
	}
	if err != nil && srv.Logger != nil {
		srv.Logger.Error("connect", "error", err)
	}
}
