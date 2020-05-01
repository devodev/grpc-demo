package hub

import (
	"io"

	"github.com/hashicorp/yamux"
)

// Client represents a remote gRPC server.
// The session stored wraps a RWC.
type Client struct {
	session *yamux.Session
}

// NewClient .
func NewClient(rwc io.ReadWriteCloser) (*Client, error) {
	s, err := yamux.Client(rwc, yamux.DefaultConfig())
	if err != nil {
		return nil, err
	}
	return &Client{session: s}, nil
}
