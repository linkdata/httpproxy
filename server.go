package httpproxy

import (
	"errors"
	"net"
	"net/http"
	"sync"
)

var DefaultContextDialer ContextDialer = &net.Dialer{}

type Server struct {
	Logger               Logger               // optional logger to use
	Handler              http.Handler         // optional handler for requests that aren't proxy requests
	DialerSelector       DialerSelector       // optional handler to select ContextDialer per proxy request, otherwise uses DefaultContextDialer
	CredentialsValidator CredentialsValidator // optional credentials validator
	RoundTripperMaker    RoundTripperMaker    // optional RoundTripperMaker, defaults to DefaultMakeRoundTripper
	mu                   sync.Mutex           // protects following
	trippers             map[ContextDialer]http.RoundTripper
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

// DefaultMakeRoundTripper clones http.DefaultTransport, sets
// it's DialContext member and returns it.
func DefaultMakeRoundTripper(cd ContextDialer) http.RoundTripper {
	tp := http.DefaultTransport.(*http.Transport).Clone()
	tp.DialContext = cd.DialContext
	return tp
}

func (srv *Server) ensureTripper(cd ContextDialer) (rt http.RoundTripper) {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	if rt = srv.trippers[cd]; rt == nil {
		if len(srv.trippers) > 100 {
			clear(srv.trippers)
		}
		rtm := DefaultMakeRoundTripper
		if srv.RoundTripperMaker != nil {
			rtm = srv.RoundTripperMaker.MakeRoundTripper
		}
		rt = rtm(cd)
		srv.trippers[cd] = rt
	}
	return
}

var ErrUnauthorized = errors.New("unauthorized")

func (srv *Server) getDialer(r *http.Request) (cd ContextDialer, address string, err error) {
	var username string
	address = getAddress(r.URL)
	if srv.CredentialsValidator != nil {
		var password string
		username, password = BasicAuth(r.Header)
		if !srv.CredentialsValidator.ValidateCredentials(username, password, address) {
			err = ErrUnauthorized
		}
	}
	if err == nil {
		cd = DefaultContextDialer
		if srv.DialerSelector != nil {
			cd, err = srv.DialerSelector.SelectDialer(username, "tcp", address)
		}
	}
	return
}

func (srv *Server) getRoundTripper(r *http.Request) (rt http.RoundTripper) {
	if cd, _, err := srv.getDialer(r); err == nil {
		rt = srv.ensureTripper(cd)
	} else {
		rt = failRoundTripper{err: err}
	}
	return
}
