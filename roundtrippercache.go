package httpproxy

import (
	"net/http"
)

type roundTripperCache struct {
	http.RoundTripper       // actual roundtripper
	counter           int64 // srv.counter when last used
}
