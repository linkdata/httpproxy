package httpproxy

import (
	"bytes"
	"crypto/tls"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

var testBody = []byte("Hello world!")

func maybeFatal(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func makeHTTPDestSrv(t *testing.T) *httptest.Server {
	t.Helper()
	destsrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(testBody)
	}))
	return destsrv
}

func makeHTTPSDestSrv(t *testing.T) *httptest.Server {
	t.Helper()
	destsrv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(testBody)
	}))
	return destsrv
}

func makeClient(t *testing.T, urlstr string) *http.Client {
	t.Helper()
	u, err := url.Parse(urlstr)
	if err != nil {
		t.Fatal(err)
	}
	tr := &http.Transport{Proxy: http.ProxyURL(u), TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	return &http.Client{Transport: tr}
}

func TestSimpleHTTPRequest(t *testing.T) {
	destsrv := makeHTTPDestSrv(t)
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

func TestUnauthorizedResponse(t *testing.T) {
	destsrv := makeHTTPDestSrv(t)
	defer destsrv.Close()

	proxysrv := httptest.NewServer(&Server{
		CredentialsValidator: StaticCredentials{"foo": "bar"},
	})
	defer proxysrv.Close()

	resp, err := makeClient(t, proxysrv.URL).Get(destsrv.URL)
	maybeFatal(t, err)

	if resp.StatusCode != http.StatusUnauthorized {
		t.Error(resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	maybeFatal(t, err)
	maybeFatal(t, resp.Body.Close())

	if resp.ContentLength != int64(len(body)) {
		t.Error("ContentLength", resp.ContentLength, "len(body)", len(body))
	}
}

func TestAuthorizedResponse(t *testing.T) {
	destsrv := makeHTTPDestSrv(t)
	defer destsrv.Close()

	proxysrv := httptest.NewServer(&Server{
		CredentialsValidator: StaticCredentials{"foo": "bar"},
	})
	defer proxysrv.Close()

	client := makeClient(t, proxysrv.URL)
	req, err := http.NewRequest(http.MethodGet, destsrv.URL, nil)
	maybeFatal(t, err)
	SetBasicAuth(req.Header, "foo", "bar")

	resp, err := client.Do(req)
	maybeFatal(t, err)

	if resp.StatusCode != http.StatusOK {
		t.Error(resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	maybeFatal(t, err)
	maybeFatal(t, resp.Body.Close())

	if resp.ContentLength != int64(len(body)) {
		t.Error("ContentLength", resp.ContentLength, "len(body)", len(body))
	}

	if !bytes.Equal(body, testBody) {
		t.Errorf("status %q: got %q, wanted %q\n", resp.Status, string(body), string(testBody))
	}
}
