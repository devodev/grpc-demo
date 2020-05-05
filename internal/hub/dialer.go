package hub

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"

	"github.com/gorilla/websocket"
	"github.com/hashicorp/yamux"
)

// Dialer is used to dial a Hub.
type Dialer struct {
	addr   string
	header http.Header
	dialer *websocket.Dialer
}

// NewDialer returns a dialer meant to be used to reach a Hub.
func NewDialer(hubAddr string, insecureSkipVerify bool, name string) (*Dialer, error) {
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

	return &Dialer{addr: u.String(), dialer: dialer, header: header}, nil
}

// Dial will dial the hub, providing hub specific headers
// and returns a websocket connection meant to be used
// as transport layer for gRPC requests.
func (h *Dialer) Dial() (*websocket.Conn, error) {
	wsConn, _, err := h.dialer.Dial(h.addr, h.header)
	if err != nil {
		return nil, err
	}
	return wsConn, nil
}

// DialAndWrap dials the hub, wraps the returned websocket connection
// using a yamux session and returns it.
// The returned session can be used as a listener for a gRPC server.
func (h *Dialer) DialAndWrap() (*yamux.Session, error) {
	wsConn, err := h.Dial()
	if err != nil {
		return nil, err
	}
	wsRwc, err := hub.NewRWC(websocket.BinaryMessage, wsConn)
	if err != nil {
		wsConn.Close()
		return nil, err
	}
	srvConn, err := yamux.Server(wsRwc, yamux.DefaultConfig())
	if err != nil {
		wsRwc.CloseWithMessage(err.Error())
		return nil, err
	}
	return srvConn, nil
}
