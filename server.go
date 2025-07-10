package httpproxy

import "net/http"

type Server struct {
	Logger               Logger            // logger to use, may be nil
	Handler              http.Handler      // handler for requests that aren't proxy requests
	RoundTripper         http.RoundTripper // default RoundTripper, if nil uses http.DefaultTransport
	RoundTripperSelector                   // handler to override default RoundTripper per proxy request
}

func (srv *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodConnect {
		srv.connect(w, r)
	} else if r.URL.IsAbs() {
		srv.proxy(w, r)
	} else if srv.Handler != nil {
		srv.Handler.ServeHTTP(w, r)
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
