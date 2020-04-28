package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

var (
	maxBufferedChannelSize = 1024 * 1024
)

// Hub maintains the set of active clients and broadcasts messages to them.
type Hub struct {
	// Generates clients that uses this hub reference.
	clientFactory ClientGenerator

	// Registered clients.
	clients map[WSClient]bool

	// Register requests from the clients.
	registerChan chan WSClient

	// Unregister requests from clients.
	unregisterChan chan WSClient
}

// NewHub .
func NewHub() *Hub {
	h := &Hub{
		registerChan:   make(chan WSClient),
		unregisterChan: make(chan WSClient),
		clients:        make(map[WSClient]bool),
	}
	h.clientFactory = NewClientFactory(h)
	return h
}

// Run implements the Hub interface.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.registerChan:
			h.clients[client] = true
		case client := <-h.unregisterChan:
			if _, ok := h.clients[client]; ok {
				client.CloseSend()
				delete(h.clients, client)
			}
		}
	}
}

// Handler creates a new router and adds necessary routes.
func (h *Hub) Handler() http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/ws", h.serveWS)
	return r
}

// ServeHTTP handles websocket requests from the peer.
func (h *Hub) serveWS(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		err := fmt.Sprintf("%v HTTP Method not allowed", r.Method)
		log.Println(err)
		http.Error(w, err, http.StatusMethodNotAllowed)
		return
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := h.clientFactory.NewClient(conn)
	h.Register(client)

	go client.WriteHandler()
	go client.ReadHandler()
}

// Register .
func (h *Hub) Register(c WSClient) {
	h.registerChan <- c
}

// Unregister .
func (h *Hub) Unregister(c WSClient) {
	h.unregisterChan <- c
}
