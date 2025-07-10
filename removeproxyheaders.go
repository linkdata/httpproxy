package httpproxy

import (
	"net/http"
	"strings"
)

func headerContains(header http.Header, name, value string) bool {
	for _, vv := range header[name] {
		for v := range strings.SplitSeq(vv, ",") {
			for partv := range strings.SplitSeq(v, ";") {
				if strings.EqualFold(value, strings.TrimSpace(partv)) {
					return true
				}
			}
		}
	}
	return false
}

func isWebSocketHandshake(header http.Header) bool {
	return headerContains(header, "Upgrade", "websocket") && headerContains(header, "Connection", "Upgrade")
}

// RemoveRequestHeaders removes request headers which should not propagate to the next hop.
func RemoveRequestHeaders(r *http.Request) {
	r.RequestURI = ""
	delete(r.Header, "Accept-Encoding")
	delete(r.Header, "Proxy-Connection")
	delete(r.Header, "Proxy-Authenticate")
	delete(r.Header, "Proxy-Authorization")
	if !isWebSocketHandshake(r.Header) {
		delete(r.Header, "Connection")
	}
}
