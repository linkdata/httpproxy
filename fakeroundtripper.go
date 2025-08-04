package httpproxy

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type fakeRoundTripper struct {
	err error
}

var fakeRoundTripperHeader = http.Header{
	"Content-Type":           {"text/plain; charset=utf-8"},
	"X-Content-Type-Options": {"nosniff"},
}

var fakeRoundTripperFprintf = fmt.Fprintf

func (f fakeRoundTripper) WriteConnectResponse(w io.Writer) (err error) {
	code := f.StatusCode(http.StatusInternalServerError)
	_, err = fakeRoundTripperFprintf(w, "HTTP/1.0 %03d %s\r\n\r\n", code, http.StatusText(code))
	return
}

func (f fakeRoundTripper) WriteResponse(w http.ResponseWriter) {
	hdr := w.Header()
	clear(hdr)
	for k, vv := range fakeRoundTripperHeader {
		hdr[k] = append([]string{}, vv...)
	}
	code := f.StatusCode(http.StatusInternalServerError)
	w.WriteHeader(code)
	if code == http.StatusInternalServerError && f.err != nil {
		_, _ = w.Write([]byte(f.err.Error()))
	}
}

func (f fakeRoundTripper) StatusCode(defaultcode int) (code int) {
	code = defaultcode
	switch f.err {
	case nil:
		code = http.StatusOK
	case ErrUnauthorized:
		code = http.StatusUnauthorized
	}
	return
}

func (f fakeRoundTripper) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	err = f.err
	code := f.StatusCode(0)
	if code != 0 {
		err = nil
		var body io.Reader = bytes.NewReader(nil)
		var hdr http.Header
		if code == http.StatusInternalServerError && f.err != nil {
			hdr = fakeRoundTripperHeader
			body = strings.NewReader(f.err.Error())
		}
		resp = &http.Response{
			Request:    req,
			Proto:      "HTTP/1.0",
			ProtoMajor: 1,
			ProtoMinor: 0,
			Status:     http.StatusText(code),
			StatusCode: code,
			Header:     hdr,
			Body:       io.NopCloser(body),
		}
	}
	return
}
