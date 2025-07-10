package httpproxy

import (
	"context"
	"net"
	"net/http"
)

var DefaultContextDialer ContextDialer = &net.Dialer{}

type Server struct {
	Logger               Logger            // optional logger to use
	Handler              http.Handler      // optional handler for requests that aren't proxy requests
	RoundTripperSelector                   // optional handler to override default RoundTripper per proxy request
	RoundTripper         http.RoundTripper // default RoundTripper, if nil uses http.DefaultTransport
	ContextDialer        ContextDialer     // default ContextDialer for CONNECT proxy requests, if nil uses DefaultContextDialer
}

func (srv *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodConnect {
		srv.connect(w, r)
	} else if r.URL.IsAbs() {
		srv.proxy(w, r)
	} else if srv.Handler != nil {
		srv.Handler.ServeHTTP(w, r)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func (srv *Server) getRoundTripper(r *http.Request) (rt http.RoundTripper) {
	rt = srv.RoundTripper
	if srv.RoundTripperSelector != nil {
		var err error
		if rt, err = srv.RoundTripperSelector.SelectRoundTripper(r); err != nil {
			rt = failRoundTripper{err}
		}
	}
	if rt == nil {
		rt = http.DefaultTransport
	}
	return
}

func (srv *Server) DialContext(ctx context.Context, network, address string) (conn net.Conn, err error) {
	cd := DefaultContextDialer
	if srv.ContextDialer != nil {
		cd = srv.ContextDialer
	}
	return cd.DialContext(ctx, network, address)
}
