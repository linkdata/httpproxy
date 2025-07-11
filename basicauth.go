package httpproxy

import (
	"encoding/base64"
	"net/http"
	"strings"
)

func BasicAuth(hdr http.Header) (username, password string, err error) {
	if authkind, encoding, found := strings.Cut(hdr.Get("Proxy-Authorization"), " "); found && authkind == "Basic" {
		var userpassraw []byte
		if userpassraw, err = base64.StdEncoding.DecodeString(encoding); err == nil {
			username, password, _ = strings.Cut(string(userpassraw), ":")
		}
	}
	return
}
