package hub

import "github.com/hashicorp/yamux"

// client .
type client struct {
	session *yamux.Session
}

// newClient .
func newClient(s *yamux.Session) *client {
	return &client{session: s}
}
