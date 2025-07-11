package httpproxy

import (
	"errors"
	"io"
	"net"
	"net/http"
)

var ErrBodyNotReadWriter = errors.New("response body not an io.ReadWriter")

func needsFlusher(hdr http.Header) (yes bool) {
	return headerContains(hdr, "Transfer-Encoding", "chunked") || headerContains(hdr, "Content-Type", "text/event-stream")
}

func (srv *Server) proxy(w http.ResponseWriter, r *http.Request) {
	rt := srv.getRoundTripper(r)
	RemoveRequestHeaders(r)
	resp, err := rt.RoundTrip(r)
	if err == nil && resp != nil {
		// replace headers and write them out
		hdr := w.Header()
		clear(hdr)
		for k, vv := range resp.Header {
			hdr[k] = append([]string{}, vv...)
		}
		w.WriteHeader(resp.StatusCode)

		// proxy the body data
		if isWebSocketHandshake(resp.Header) {
			var clientConn net.Conn
			if clientConn, err = hijack(w); err == nil {
				err = ErrBodyNotReadWriter
				if wsConn, ok := resp.Body.(io.ReadWriter); ok {
					err = proxyWebsocket(wsConn, clientConn)
				}
			}
		} else {
			var copyWriter io.Writer = w
			if needsFlusher(hdr) {
				copyWriter = newFlushWriter(w)
			}
			_, err = io.Copy(copyWriter, resp.Body)
			err = errors.Join(err, resp.Body.Close())
		}
	} else {
		w.WriteHeader(http.StatusBadGateway)
	}

	if err != nil && srv.Logger != nil {
		srv.Logger.Error("proxy", "error", err)
	}
}
