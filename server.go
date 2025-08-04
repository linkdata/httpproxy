package httpproxy

import (
	"errors"
	"net"
	"net/http"
	"slices"
	"sync"
)

var MaxCachedRoundTrippers = 100
var DefaultContextDialer ContextDialer = &net.Dialer{}

type Server struct {
	Logger               Logger                               // optional logger to use
	Handler              http.Handler                         // optional handler for requests that aren't proxy requests
	DialerSelector       DialerSelector                       // optional handler to select ContextDialer per proxy request, otherwise uses DefaultContextDialer
	CredentialsValidator CredentialsValidator                 // optional credentials validator
	RoundTripperMaker    RoundTripperMaker                    // optional RoundTripperMaker, defaults to DefaultMakeRoundTripper
	mu                   sync.Mutex                           // protects following
	counter              int64                                // counts ensureTripper calls
	trippers             map[ContextDialer]*roundTripperCache // LRU cache mapping CD -> RT
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

func (srv *Server) cleanTripperCacheLocked() {
	type roundTripperCacheList struct {
		ContextDialer
		*roundTripperCache
	}
	var trippers []roundTripperCacheList
	for cd, rtc := range srv.trippers {
		trippers = append(trippers, roundTripperCacheList{ContextDialer: cd, roundTripperCache: rtc})
	}
	slices.SortFunc(trippers, func(a, b roundTripperCacheList) int { return int(b.counter - a.counter) })
	for i, rtcl := range trippers {
		if i >= MaxCachedRoundTrippers/2 {
			delete(srv.trippers, rtcl.ContextDialer)
		}
	}
}

func (srv *Server) ensureTripper(cd ContextDialer) (rt http.RoundTripper) {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	var rtc *roundTripperCache
	if rtc = srv.trippers[cd]; rtc == nil {
		if srv.trippers == nil {
			srv.trippers = make(map[ContextDialer]*roundTripperCache)
		}
		if len(srv.trippers) >= MaxCachedRoundTrippers {
			srv.cleanTripperCacheLocked()
		}
		rtm := DefaultMakeRoundTripper
		if srv.RoundTripperMaker != nil {
			rtm = srv.RoundTripperMaker.MakeRoundTripper
		}
		rt = rtm(cd)
		rtc = &roundTripperCache{RoundTripper: rt}
		srv.trippers[cd] = rtc
	}
	srv.counter++
	rtc.counter = srv.counter
	return rtc.RoundTripper
}

var ErrUnauthorized = errors.New("unauthorized")

func (srv *Server) getDialer(r *http.Request) (cd ContextDialer, address string, err error) {
	var username string
	address = getAddress(r.URL)
	if srv.CredentialsValidator != nil {
		var password string
		if username, password, err = GetBasicAuth(r.Header); err == nil {
			if !srv.CredentialsValidator.ValidateCredentials(username, password, address) {
				err = ErrUnauthorized
			}
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
		rt = fakeRoundTripper{err: err}
	}
	return
}
