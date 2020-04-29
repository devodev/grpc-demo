package ws

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
)

// RWC .
type RWC struct {
	r          io.Reader
	pingTicker *time.Ticker

	mt int
	c  *websocket.Conn
}

// RWCOption provide a functional wa of
type RWCOption func(*RWC)

// WithPingEnabled starts a pinger process.
func WithPingEnabled() RWCOption {
	return func(c *RWC) {
		c.enablePing()
	}
}

//WithPongHandler sets a pong handler.
func WithPongHandler() RWCOption {
	return func(c *RWC) {
		c.c.SetReadDeadline(time.Now().Add(pongWait))
		c.c.SetPongHandler(func(string) error { c.c.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	}
}

// NewRWC returns a websocket ReadWriteCloser enforcing the provided
// message type on write/read.
func NewRWC(mt int, conn *websocket.Conn, options ...RWCOption) (*RWC, error) {
	if mt != websocket.BinaryMessage && mt != websocket.TextMessage {
		return nil, fmt.Errorf("invalid message type")
	}
	rwc := &RWC{mt: mt, c: conn}
	for _, opt := range options {
		opt(rwc)
	}
	return rwc, nil
}

// Write .
func (c *RWC) Write(p []byte) (int, error) {
	c.c.SetWriteDeadline(time.Now().Add(writeWait))
	err := c.c.WriteMessage(websocket.BinaryMessage, p)
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

// Close .
func (c *RWC) Close() error {
	c.c.SetWriteDeadline(time.Now().Add(writeWait))
	c.c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "close"))
	return c.c.Close()
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
				log.Printf("error sending ping: %v", err)
				return
			}
		}
	}
}
