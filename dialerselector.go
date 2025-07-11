package httpproxy

// A DialerSelector returns the ContextDialer to use.
type DialerSelector interface {
	// SelectDialer returns the ContextDialer to use.
	//
	// If username is the empty string no authorization has taken place (anonymous usage).
	SelectDialer(username, network, address string) (cd ContextDialer, err error)
}
