package main

import (
	"bytes"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512

	// default send channel size
	defaultBufferedChannelSize = 256
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// WSClient .
type WSClient interface {
	ReadHandler()
	WriteHandler()
	Send([]byte) error
	CloseSend()
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub *Hub

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	// Make sure we close the cahnnel only once.
	sendCloseOnce *sync.Once
}

// ReadHandler pumps messages from the websocket connection to the hub.
//
// The application runs ReadHandler in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) ReadHandler() {
	defer func() {
		c.hub.Unregister(c)
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		mType, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		log.Printf("received message of type %v: %v", mType, string(message))
	}
}

// WriteHandler pumps messages from the hub to the websocket connection.
//
// A goroutine running WriteHandler is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) WriteHandler() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		log.Println("closing connection")
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				log.Println("channel was closed")
				return
			}
			log.Printf("sending message: %v", message)

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				log.Printf("error closing writer: %v", err)
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("error sending ping: %v", err)
				return
			}
		}
	}
}

// Send sends a message on the send channel, and does nothing
// if channel is closed or busy.
func (c *Client) Send(message []byte) error {
	select {
	case c.send <- message:
	default:
		return fmt.Errorf("channel closed")
	}
	return nil
}

// CloseSend closes the client send channel.
func (c *Client) CloseSend() {
	c.sendCloseOnce.Do(func() {
		close(c.send)
	})
}

// ClientGenerator provides a factory for generating clients.
type ClientGenerator interface {
	NewClient(conn *websocket.Conn) WSClient
}

// ClientFactory returns a factory for creating clients
// that inherits the provided hub.
type ClientFactory struct {
	hub                 *Hub
	bufferedChannelSize int
}

// NewClientFactory returns a factory that generates Clients.
func NewClientFactory(h *Hub, options ...ClientFactoryOption) *ClientFactory {
	factory := &ClientFactory{hub: h, bufferedChannelSize: defaultBufferedChannelSize}
	for _, opt := range options {
		opt(factory)
	}
	return factory
}

// ClientFactoryOption is a functional option representation
// for ClientFactory.
type ClientFactoryOption func(*ClientFactory)

// WithBufferedChannelSize allows updating the bufferedChannleSize attribute.
func WithBufferedChannelSize(size int) func(*ClientFactory) {
	return func(c *ClientFactory) {
		c.bufferedChannelSize = size
	}
}

// NewClient returns a new client using the provided connection and applying supplied options.
func (f *ClientFactory) NewClient(conn *websocket.Conn) WSClient {
	c := &Client{
		hub:           f.hub,
		conn:          conn,
		send:          make(chan []byte, f.bufferedChannelSize),
		sendCloseOnce: &sync.Once{},
	}
	return c
}
