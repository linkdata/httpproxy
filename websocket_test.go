package httpproxy

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
)

func TestSimpleWSRequest(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()

	destsrv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := websocket.Accept(w, r, nil)
			if err == nil {
				defer c.CloseNow()
				var mt websocket.MessageType
				var b []byte
				if mt, b, err = c.Read(ctx); err == nil {
					err = c.Write(ctx, mt, b)
				}
				c.Close(websocket.StatusNormalClosure, "")
			}
			if err != nil {
				t.Error(err)
			}
		}))
	defer destsrv.Close()

	proxysrv := httptest.NewServer(&Server{})
	defer proxysrv.Close()

	c, _, err := websocket.Dial(ctx, strings.ReplaceAll(destsrv.URL, "http:", "ws:"), &websocket.DialOptions{
		HTTPClient: makeClient(t, proxysrv.URL),
	})
	if err != nil {
		t.Fatal(err)
	}
	defer c.CloseNow()

	var testmessage = []byte("hi")
	err = c.Write(ctx, websocket.MessageText, testmessage)
	if err != nil {
		t.Error(err)
	}
	var mt websocket.MessageType
	var b []byte
	mt, b, err = c.Read(ctx)
	if err != nil {
		t.Error(err)
	}
	if mt != websocket.MessageText || !bytes.Equal(b, testmessage) {
		t.Errorf("%q != %q", b, testmessage)
	}
	err = c.Close(websocket.StatusNormalClosure, "")
	if err != nil {
		t.Error(err)
	}
}
