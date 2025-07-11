package httpproxy

import (
	"io"
	"net/http"
	"strings"
)

type failRoundTripper struct {
	code int   // HTTP status code to return, leave zero to not generate response and return no error from RoundTrip
	err  error // error, will be rendered in response if code nonzero, otherwise returned from RoundTrip
}

var failRoundTripperHeader = http.Header{"Content-Type": {"text/plain; charset=utf-8"}}

func (f failRoundTripper) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	if f.code == 0 {
		err = f.err
	} else {
		var body io.ReadCloser
		var hdr http.Header
		if f.err != nil {
			hdr = failRoundTripperHeader
			body = io.NopCloser(strings.NewReader(f.err.Error()))
		}
		resp = &http.Response{
			Request:    req,
			Proto:      "HTTP/1.0",
			ProtoMajor: 1,
			ProtoMinor: 0,
			Status:     http.StatusText(f.code),
			StatusCode: f.code,
			Header:     hdr,
			Body:       body,
		}
	}
	return
}
