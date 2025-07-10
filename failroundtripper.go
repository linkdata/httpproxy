package httpproxy

import "net/http"

type failRoundTripper struct {
	err error
}

func (f failRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, f.err
}
