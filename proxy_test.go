package httpproxy

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func maybeFatal(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

func makeClient(s *httptest.Server) *http.Client {
	u, _ := url.Parse(s.URL)
	tr := &http.Transport{Proxy: http.ProxyURL(u)}
	return &http.Client{Transport: tr}
}

func TestSimpleHTTPRequest(t *testing.T) {
	bodyText := []byte("Hello world!")

	destsrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(bodyText)
	}))
	defer destsrv.Close()

	proxysrv := httptest.NewServer(&Server{})
	defer proxysrv.Close()

	resp, err := makeClient(proxysrv).Get(destsrv.URL)
	maybeFatal(t, err)

	body, err := io.ReadAll(resp.Body)
	maybeFatal(t, err)
	if !bytes.Equal(body, bodyText) {
		t.Errorf("status %q: got %q, wanted %q\n", resp.Status, string(body), string(bodyText))
	}
}
