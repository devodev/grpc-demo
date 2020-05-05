package hub

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"

	ws "github.com/devodev/grpc-demo/internal/websocket"
	"github.com/gorilla/websocket"
	"github.com/hashicorp/yamux"
)

// Connector is used to dial a Hub.
type Connector struct {
	addr   string
	header http.Header
	dialer *websocket.Dialer
}

// NewConnector returns a connector that can reach a Hub and provide a listener
// to be used when serving HTTP.
func NewConnector(hubAddr string, insecureSkipVerify bool, name string) (*Connector, error) {
	u, err := url.Parse(hubAddr)
	if err != nil {
		return nil, err
	}

	dialer := websocket.DefaultDialer
	if insecureSkipVerify {
		dialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	if name == "" {
		return nil, fmt.Errorf("name is empty")
	}
	header := make(http.Header)
	header.Add("X-Hub-Meta-Name", name)

	return &Connector{addr: u.String(), dialer: dialer, header: header}, nil
}

// Listener dials the hub, wraps the underlying connection
// as a listener and returns it.
func (h *Connector) Listener() (net.Listener, error) {
	wsConn, err := h.dial()
	if err != nil {
		return nil, err
	}
	return h.asListener(wsConn)
}

// dial returns a valid websocket connection to be used
// as a io.ReadWriteCloser.
func (h *Connector) dial() (*websocket.Conn, error) {
	wsConn, _, err := h.dialer.Dial(h.addr, h.header)
	if err != nil {
		return nil, err
	}
	return wsConn, nil
}

func (h *Connector) asListener(c *websocket.Conn) (*yamux.Session, error) {
	wsRwc, err := ws.ReadWriteCloser(c)
	if err != nil {
		c.Close()
		return nil, err
	}
	srvConn, err := yamux.Server(wsRwc, yamux.DefaultConfig())
	if err != nil {
		wsRwc.CloseWithMessage(err.Error())
		return nil, err
	}
	return srvConn, nil
}
