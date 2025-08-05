package httpproxy_test

import (
	"fmt"
	"net/http"

	"github.com/linkdata/httpproxy"
)

func Example() {
	mux := http.NewServeMux()
	mux.HandleFunc("/{$}", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "This is a web proxy server.")
	})
	// the httpproxy.Server will handle CONNECT and absolute URL requests,
	// and all others will be forwarded to the http.Handler we set.
	go http.ListenAndServe(":8080", &httpproxy.Server{Handler: mux})
}
