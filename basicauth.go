package httpproxy

import (
	"encoding/base64"
	"net/http"
	"strings"
)

func BasicAuth(hdr http.Header) (username, password string) {
	const proxyAuthorizationHeader = "Proxy-Authorization"
	authheader := strings.SplitN(hdr.Get(proxyAuthorizationHeader), " ", 2)
	if len(authheader) == 2 && authheader[0] == "Basic" {
		if userpassraw, err := base64.StdEncoding.DecodeString(authheader[1]); err == nil {
			userpass := strings.SplitN(string(userpassraw), ":", 2)
			if len(userpass) == 2 {
				username, password = userpass[0], userpass[1]
			}
		}
	}
	return
}
