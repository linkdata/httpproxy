package httpproxy

import (
	"bytes"
	"errors"
	"io"
	"log/slog"
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

func makeDestSrv(t *testing.T) *httptest.Server {
	t.Helper()
	destsrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	tr := &http.Transport{Proxy: http.ProxyURL(u)}
	return &http.Client{Transport: tr}
}

func TestSimpleHTTPRequest(t *testing.T) {
	destsrv := makeDestSrv(t)
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

type failCredentials struct{}

func (failCredentials) ValidateCredentials(_, _, _ string) bool { return false }

func TestUnauthorizedResponse(t *testing.T) {
	destsrv := makeDestSrv(t)
	defer destsrv.Close()

	proxysrv := httptest.NewServer(&Server{
		CredentialsValidator: failCredentials{},
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
	destsrv := makeDestSrv(t)
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

type failMakeRoundTripper struct{}

func (failMakeRoundTripper) MakeRoundTripper(cd ContextDialer) (rt http.RoundTripper) {
	return fakeRoundTripper{errors.New("foo")}
}

func TestMakeRoundTripper(t *testing.T) {
	destsrv := makeDestSrv(t)
	defer destsrv.Close()

	proxysrv := httptest.NewServer(&Server{
		Logger:            slog.Default(),
		RoundTripperMaker: failMakeRoundTripper{},
	})
	defer proxysrv.Close()

	resp, err := makeClient(t, proxysrv.URL).Get(destsrv.URL)
	maybeFatal(t, err)

	if resp.StatusCode != http.StatusInternalServerError {
		t.Error(resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	maybeFatal(t, err)
	maybeFatal(t, resp.Body.Close())

	if resp.ContentLength != int64(len(body)) {
		t.Error("ContentLength", resp.ContentLength, "len(body)", len(body))
	}

	if string(body) != "foo" {
		t.Error(string(body))
	}
}
