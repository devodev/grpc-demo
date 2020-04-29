package ws

import (
	"fmt"
	"io"

	"github.com/gorilla/websocket"
)

// RWC .
type RWC struct {
	r    io.Reader
	Conn *websocket.Conn
}

// Write .
func (c *RWC) Write(p []byte) (int, error) {
	err := c.Conn.WriteMessage(websocket.BinaryMessage, p)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

// Read .
func (c *RWC) Read(p []byte) (int, error) {
	for {
		if c.r == nil {
			// Advance to next message.
			var mt int
			var err error
			mt, c.r, err = c.Conn.NextReader()
			if err != nil {
				return 0, err
			}
			if mt != websocket.BinaryMessage {
				return 0, fmt.Errorf("message type is not BinaryMessage")
			}
		}
		n, err := c.r.Read(p)
		if err == io.EOF {
			// At end of message.
			c.r = nil
			if n > 0 {
				return n, nil
			}
			// No data read, continue to next message.
			continue
		}
		return n, err
	}
}

// Close .
func (c *RWC) Close() error {
	return c.Conn.Close()
}
