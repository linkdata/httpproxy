package httpproxy

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
)

func TestGetAddress(t *testing.T) {
	u1, _ := url.Parse("http://foo.bar/")
	if x := getAddress(u1); x != "foo.bar:80" {
		t.Error(x)
	}
	u2, _ := url.Parse("https://foo.bar/")
	if x := getAddress(u2); x != "foo.bar:443" {
		t.Error(x)
	}
}

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

func TestNotAHijacker(t *testing.T) {
	var logbuf bytes.Buffer
	logger1 := slog.New(slog.NewTextHandler(&logbuf, nil))
	srv := &Server{Logger: logger1}
	rw := httptest.NewRecorder()
	srv.connect(rw, nil)
	if x := logbuf.String(); !strings.Contains(x, ErrHijackingNotSupported.Error()) {
		t.Error(x)
	}
}

var fakeRoundTripperFprintfFail atomic.Bool

func init() {
	fakeRoundTripperFprintf = func(w io.Writer, format string, a ...any) (n int, err error) {
		if fakeRoundTripperFprintfFail.Load() {
			return 0, io.EOF
		}
		return fmt.Fprintf(w, format, a...)
	}
}

func TestFailHTTPSRequestViaHTTP(t *testing.T) {
	fakeRoundTripperFprintfFail.Store(true)
	defer func() {
		fakeRoundTripperFprintfFail.Store(false)
	}()

	destsrv := makeHTTPSDestSrv(t)
	defer destsrv.Close()

	proxysrv := httptest.NewServer(&Server{
		Logger: slog.Default(),
	})
	defer proxysrv.Close()

	client := makeClient(t, proxysrv.URL)
	resp, err := client.Get(destsrv.URL)
	if resp != nil {
		resp.Body.Close()
	}
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), io.EOF.Error()) {
		t.Error(err)
	}
}
