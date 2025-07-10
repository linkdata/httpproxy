package httpproxy

import "net/http"

type RoundTripperSelector interface {
	SelectRoundTripper(r *http.Request) (rt http.RoundTripper, err error)
}
