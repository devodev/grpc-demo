package hub

import "github.com/hashicorp/yamux"

// Client .
type Client struct {
	id      uint64
	session *yamux.Session
}

// NewClient .
func NewClient(id uint64, s *yamux.Session) *Client {
	c := &Client{
		id:      id,
		session: s,
	}
	return c
}
