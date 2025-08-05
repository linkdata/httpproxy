package httpproxy

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
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

func TestChunkedResponse(t *testing.T) {
	const want = "This is the data in the first chunk\r\nand this is the second one\r\nconsequence"

	ctx, cancel := context.WithTimeout(t.Context(), time.Minute)
	defer cancel()

	l, err := net.Listen("tcp", "localhost:0")
	maybeFatal(t, err)
	defer l.Close()
	laddr := l.Addr().String()

	go func() {
		for i := 0; i < 2; i++ {
			c, err := l.Accept()
			if err == nil {
				if _, err = http.ReadRequest(bufio.NewReader(c)); err == nil {
					_, err = io.WriteString(c, "HTTP/1.1 200 OK\r\n"+
						"Content-Type: text/plain\r\n"+
						"Transfer-Encoding: chunked\r\n\r\n"+
						"25\r\n"+
						"This is the data in the first chunk\r\n\r\n"+
						"1C\r\n"+
						"and this is the second one\r\n\r\n"+
						"3\r\n"+
						"con\r\n"+
						"8\r\n"+
						"sequence\r\n0\r\n\r\n")
					err = errors.Join(err, c.Close())
				}
			}
			if err != nil {
				t.Error(err)
			}
		}
	}()

	// do a normal HTTP request to check correctness of goroutine above
	c, err := net.Dial("tcp", laddr)
	maybeFatal(t, err)
	defer c.Close()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/", nil)
	maybeFatal(t, err)
	err = req.Write(c)
	maybeFatal(t, err)
	resp, err := http.ReadResponse(bufio.NewReader(c), req)
	maybeFatal(t, err)
	b, err := io.ReadAll(resp.Body)
	maybeFatal(t, err)
	if string(b) != want {
		t.Errorf(" got %q\nwant %q\n", string(b), want)
	}

	proxysrv := httptest.NewServer(&Server{})
	defer proxysrv.Close()

	resp, err = makeClient(t, proxysrv.URL).Get("http://" + laddr)
	maybeFatal(t, err)

	body, err := io.ReadAll(resp.Body)
	maybeFatal(t, err)
	maybeFatal(t, resp.Body.Close())

	if !bytes.Equal(body, []byte(want)) {
		t.Errorf(" got %q\nwant %q\n", string(body), string(want))
	}
}
