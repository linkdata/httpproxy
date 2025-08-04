package httpproxy

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

type failMakeRoundTripper struct{}

func (failMakeRoundTripper) MakeRoundTripper(cd ContextDialer) (rt http.RoundTripper) {
	return fakeRoundTripper{errors.New("failMakeRoundTripper")}
}

func TestMakeRoundTripper(t *testing.T) {
	destsrv := makeDestSrv(t)
	defer destsrv.Close()

	proxysrv := httptest.NewServer(&Server{
		Logger:            slog.Default(),
		RoundTripperMaker: failMakeRoundTripper{},
		DialerSelector:    nil,
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

	if string(body) != "failMakeRoundTripper" {
		t.Error(string(body))
	}
}

type failSelectDialer struct{}

func (failSelectDialer) SelectDialer(username, network, address string) (cd ContextDialer, err error) {
	return nil, errors.New("failSelectDialer")
}

func TestSelectDialer(t *testing.T) {
	destsrv := makeDestSrv(t)
	defer destsrv.Close()

	proxysrv := httptest.NewServer(&Server{
		Logger:         slog.Default(),
		DialerSelector: failSelectDialer{},
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

	if string(body) != "failSelectDialer" {
		t.Error(string(body))
	}
}

type testContextDialer string

func (testContextDialer) SelectDialer(username, network, address string) (cd ContextDialer, err error) {
	return testContextDialer(address), nil
}

func (testContextDialer) DialContext(ctx context.Context, network, address string) (conn net.Conn, err error) {
	return DefaultContextDialer.DialContext(ctx, network, address)
}

func TestRoundTripperMap(t *testing.T) {
	srv := &Server{
		Logger:         slog.Default(),
		DialerSelector: testContextDialer(""),
	}

	m := map[int]http.RoundTripper{}
	for i := range MaxCachedRoundTrippers {
		m[i] = srv.ensureTripper(testContextDialer(strconv.Itoa(i)))
	}
	if len(srv.trippers) != MaxCachedRoundTrippers {
		t.Fatal(len(srv.trippers))
	}
	for i := range MaxCachedRoundTrippers {
		if m[i] != srv.ensureTripper(testContextDialer(strconv.Itoa(i))) {
			t.Fatal(i)
		}
		if i > 0 {
			if m[i-1] == m[i] {
				t.Fatal(i)
			}
		}
	}
	_ = srv.ensureTripper(testContextDialer(strconv.Itoa(MaxCachedRoundTrippers)))
	if len(srv.trippers) != MaxCachedRoundTrippers/2+1 {
		t.Fatal("tripper MaxCachedRoundTrippers was in cache")
	}
	_ = srv.ensureTripper(testContextDialer(strconv.Itoa(MaxCachedRoundTrippers - 1)))
	if len(srv.trippers) != MaxCachedRoundTrippers/2+1 {
		t.Fatal("tripper MaxCachedRoundTrippers-1 should have been in cache")
	}
}
