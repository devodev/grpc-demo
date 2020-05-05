package hub

import (
	"fmt"
	"io"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

var (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = int64(512)

	// CloseNormalClosure is the default websocket close message.
	// The substring "closed" must be present for the
	// gorilla client to return without error.
	closeNormalClosureMessage = "closed"
)

// PongHandlerFunc .
type PongHandlerFunc func(appData string) error

// RWCOption provide a functional wa of
type RWCOption func(*RWC) error

// WithPingDisabled disables the ping process.
func WithPingDisabled() RWCOption {
	return func(c *RWC) error {
		c.pingEnabled = false
		return nil
	}
}

// WithPongHandler sets a pong handler.
func WithPongHandler(f PongHandlerFunc) RWCOption {
	return func(c *RWC) error {
		c.setPongHandler(f)
		return nil
	}
}

// WithMessageType sets the message type to use in Read/Write.
func WithMessageType(mt int) RWCOption {
	if mt != websocket.BinaryMessage && mt != websocket.TextMessage {
		return func(c *RWC) error {
			return fmt.Errorf("invalid message type")
		}
	}
	return func(c *RWC) error {
		c.mt = mt
		return nil
	}
}

// RWC .
type RWC struct {
	r  io.Reader
	mt int
	c  *websocket.Conn

	pingEnabled     bool
	pingTicker      *time.Ticker
	pongHandlerFunc PongHandlerFunc
}

// ReadWriteCloser returns a websocket ReadWriteCloser enforcing the provided
// message type on write/read.
func ReadWriteCloser(conn *websocket.Conn, options ...RWCOption) (*RWC, error) {
	rwc := &RWC{
		mt:              websocket.BinaryMessage,
		c:               conn,
		pingEnabled:     true,
		pongHandlerFunc: func(string) error { conn.SetReadDeadline(time.Now().Add(pongWait)); return nil },
	}
	for _, opt := range options {
		if err := opt(rwc); err != nil {
			return nil, err
		}
	}
	conn.SetPongHandler(rwc.pongHandlerFunc)
	if rwc.pingEnabled {
		rwc.enablePing()
	}
	return rwc, nil
}

// Write .
func (c *RWC) Write(p []byte) (int, error) {
	c.c.SetWriteDeadline(time.Now().Add(writeWait))
	err := c.c.WriteMessage(c.mt, p)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

// Read .
func (c *RWC) Read(p []byte) (int, error) {
	c.c.SetReadLimit(maxMessageSize)
	c.c.SetReadDeadline(time.Now().Add(pongWait))
	for {
		if c.r == nil {
			// Advance to next message.
			var mt int
			var err error
			mt, c.r, err = c.c.NextReader()
			if err != nil {
				return 0, err
			}
			if mt != c.mt {
				return 0, fmt.Errorf("invalid message type received")
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

// Close closes the underlying websocket connection
// and uses the default close message.
func (c *RWC) Close() error {
	return c.CloseWithMessage(closeNormalClosureMessage)
}

// CloseWithMessage closes the underlying websocket connection
// and uses the provided string as close message.
func (c *RWC) CloseWithMessage(m string) error {
	c.c.SetWriteDeadline(time.Now().Add(writeWait))
	c.c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, m))
	return c.c.Close()
}

func (c *RWC) setPongHandler(f PongHandlerFunc) {
	c.pongHandlerFunc = f
}

// EnablePing .
func (c *RWC) enablePing() {
	if c.pingTicker != nil {
		return
	}
	done := make(chan struct{})
	c.c.SetCloseHandler(func(code int, text string) error {
		close(done)
		return nil
	})
	go c.ping(done)
}

// ping is a long running goroutine that sends ping messages.
func (c *RWC) ping(done chan struct{}) {
	if done == nil {
		return
	}
	c.pingTicker = time.NewTicker(pingPeriod)
	defer c.pingTicker.Stop()
	for {
		select {
		case <-done:
			return
		case <-c.pingTicker.C:
			c.c.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.c.WriteMessage(websocket.PingMessage, nil); err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure) {
					log.Printf("error sending ping: %v", err)
				}
				return
			}
		}
	}
}
