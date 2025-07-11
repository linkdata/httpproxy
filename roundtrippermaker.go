package httpproxy

import "net/http"

type RoundTripperMaker interface {
	MakeRoundTripper(cd ContextDialer) (rt http.RoundTripper)
}
