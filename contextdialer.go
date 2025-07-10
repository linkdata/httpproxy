package httpproxy

import (
	"context"
	"net"
)

type ContextDialer interface {
	DialContext(ctx context.Context, network, address string) (conn net.Conn, err error)
}
