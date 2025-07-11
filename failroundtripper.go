package httpproxy

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

type failRoundTripper struct {
	err error // error, will be rendered in response if code nonzero, otherwise returned from RoundTrip
}

var failRoundTripperHeader = http.Header{"Content-Type": {"text/plain; charset=utf-8"}}

func (f failRoundTripper) WriteResponse(w io.Writer) {
	code := f.StatusCode()
	if code == 0 {
		code = 500
	}
	_, _ = fmt.Fprintf(w, "HTTP/1.0 %03d %s\r\n\r\n", code, http.StatusText(code))
}

func (f failRoundTripper) StatusCode() (code int) {
	switch f.err {
	case ErrUnauthorized:
		code = http.StatusUnauthorized
	}
	return
}

func (f failRoundTripper) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	err = f.err
	code := f.StatusCode()
	if code != 0 {
		err = nil
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
			Status:     http.StatusText(code),
			StatusCode: code,
			Header:     hdr,
			Body:       body,
		}
	}
	return
}
