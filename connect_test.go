package httpproxy

import (
	"bytes"
	"io"
	"net/http/httptest"
	"testing"
)

func TestSimpleHTTPSRequestViaHTTP(t *testing.T) {
	destsrv := makeHTTPSDestSrv(t)
	defer destsrv.Close()

	proxysrv := httptest.NewServer(&Server{})
	defer proxysrv.Close()

	resp, err := makeClient(t, proxysrv.URL).Get(destsrv.URL)
	maybeFatal(t, err)

	body, err := io.ReadAll(resp.Body)
	maybeFatal(t, err)
	maybeFatal(t, resp.Body.Close())

	if !bytes.Equal(body, testBody) {
		t.Errorf("status %q: got %q, wanted %q\n", resp.Status, string(body), string(testBody))
	}
}

func TestSimpleHTTPSRequestViaHTTPS(t *testing.T) {
	destsrv := makeHTTPSDestSrv(t)
	defer destsrv.Close()

	proxysrv := httptest.NewTLSServer(&Server{})
	defer proxysrv.Close()

	resp, err := makeClient(t, proxysrv.URL).Get(destsrv.URL)
	maybeFatal(t, err)

	body, err := io.ReadAll(resp.Body)
	maybeFatal(t, err)
	maybeFatal(t, resp.Body.Close())

	if !bytes.Equal(body, testBody) {
		t.Errorf("status %q: got %q, wanted %q\n", resp.Status, string(body), string(testBody))
	}
}
