package httpproxy

// A DialerSelector returns the ContextDialer to use.
type DialerSelector interface {
	// SelectDialer returns the ContextDialer to use.
	SelectDialer(username, network, address string) (cd ContextDialer, err error)
}
