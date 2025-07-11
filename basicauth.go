package httpproxy

import (
	"encoding/base64"
	"net/http"
	"strings"
)

const proxyAuthorizationHeader = "Proxy-Authorization"

func GetBasicAuth(hdr http.Header) (username, password string, err error) {
	if authkind, encoding, found := strings.Cut(hdr.Get(proxyAuthorizationHeader), " "); found && authkind == "Basic" {
		var userpassraw []byte
		if userpassraw, err = base64.StdEncoding.DecodeString(encoding); err == nil {
			username, password, _ = strings.Cut(string(userpassraw), ":")
		}
	}
	return
}

func SetBasicAuth(hdr http.Header, username, password string) {
	hdr.Set(proxyAuthorizationHeader, "Basic "+base64.StdEncoding.EncodeToString([]byte(username+":"+password)))
}
