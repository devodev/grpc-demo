package hub

import "github.com/hashicorp/yamux"

// Client represents a remote gRPC server.
// The session stored wraps a RWC.
type Client struct {
	session *yamux.Session
}

// NewClient .
func NewClient(s *yamux.Session) *Client {
	return &Client{session: s}
}
