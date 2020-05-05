package client

import (
	"fmt"
	"io"

	"github.com/hashicorp/yamux"
)

// ErrEmptyAttribute .
type ErrEmptyAttribute struct {
	attName string
}

func (e ErrEmptyAttribute) Error() string {
	return fmt.Sprintf("%v is empty", e.attName)
}

// Client represents a remote gRPC server.
// The session stored wraps a RWC.
type Client struct {
	Name string

	Session *yamux.Session
}

// New creates a client using the provided ReadWriteCloser and name.
func New(rwc io.ReadWriteCloser, name string) (*Client, error) {
	if name == "" {
		return nil, &ErrEmptyAttribute{"name"}
	}
	s, err := yamux.Client(rwc, yamux.DefaultConfig())
	if err != nil {
		return nil, err
	}
	return &Client{Name: name, Session: s}, nil
}
