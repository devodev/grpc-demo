package hub

import (
	"fmt"
	"io"

	"github.com/hashicorp/yamux"
)

type errEmptyAttribute struct {
	attName string
}

func (e errEmptyAttribute) Error() string {
	return fmt.Sprintf("%v is empty", e.attName)
}

// Client represents a remote gRPC server.
// The session stored wraps a RWC.
type Client struct {
	Name string

	session *yamux.Session
}

// NewClient .
func NewClient(rwc io.ReadWriteCloser, name string) (*Client, error) {
	if name == "" {
		return nil, &errEmptyAttribute{"name"}
	}
	s, err := yamux.Client(rwc, yamux.DefaultConfig())
	if err != nil {
		return nil, err
	}
	return &Client{Name: name, session: s}, nil
}
