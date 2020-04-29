package hub

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/hashicorp/yamux"
)

var (
	readTimeout     = 5 * time.Second
	writeTimeout    = 10 * time.Second
	idleTimeout     = 15 * time.Second
	shutdownTimeout = 30 * time.Second
)

// Hub .
type Hub struct {
	logger *log.Logger
	server *http.Server

	counter uint64

	mu      *sync.Mutex
	clients map[uint64]*Client
}

// New .
func New(addr string, l *log.Logger) *Hub {
	if l == nil {
		l = log.New(os.Stderr, "hub: ", log.LstdFlags)
	}
	h := &Hub{
		mu:      &sync.Mutex{},
		clients: make(map[uint64]*Client),
		logger:  l,
	}
	router := http.NewServeMux()
	router.HandleFunc("/api/list", h.handleListClients())
	router.HandleFunc("/ws", h.wsHandler())
	h.server = &http.Server{
		Addr:         addr,
		Handler:      tracingMiddleware(nextRequestID)(loggingMiddleware(l)(router)),
		ErrorLog:     h.logger,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}
	return h
}

// Start is a long running goroutine.
func (h *Hub) Start(quit chan os.Signal) <-chan struct{} {
	done := make(chan struct{})

	go func() {
		<-quit

		h.logger.Println("hub is shutting down...")

		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		h.server.SetKeepAlivesEnabled(false)
		if err := h.server.Shutdown(ctx); err != nil {
			h.logger.Printf("error during server shutdown: %v", err)
		}
		close(done)
	}()

	go func() {
		h.logger.Printf("listening on: %v", h.server.Addr)
		if err := h.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			h.logger.Printf("listen error: %v", err)
		}
	}()
	return done
}

func (h *Hub) handleListClients() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "", http.StatusMethodNotAllowed)
			return
		}
		h.mu.Lock()
		clientCount := len(h.clients)
		h.mu.Unlock()
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Number of clients: %+v", clientCount)
	}
}

func (h *Hub) wsHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		wsConn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			h.logger.Printf("upgrade: %v", err)
			return
		}
		// wrap websocket conn into ReadWriteCloser
		wsRwc, err := NewRWC(websocket.BinaryMessage, wsConn)
		if err != nil {
			h.logger.Println(err)
			return
		}

		// manage ReadWriteCloser using yamux client
		session, err := yamux.Client(wsRwc, yamux.DefaultConfig())
		if err != nil {
			h.logger.Printf("error creating yamux client: %s", err)
			return
		}

		clientID := h.getID()

		go func() {
			defer h.unregisterClient(clientID)
			select {
			case <-session.CloseChan():
				h.logger.Printf("client%d: connecton closed", clientID)
				return
			}
		}()

		client := NewClient(clientID, session)
		h.registerClient(clientID, client)
	}
}

func (h *Hub) getID() uint64 {
	return atomic.AddUint64(&h.counter, 1)
}

func (h *Hub) registerClient(id uint64, c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[id] = c
}

func (h *Hub) unregisterClient(id uint64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, id)
}

type requestIDGenerator func() string

func nextRequestID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

type key int

const requestIDKey key = 0

func loggingMiddleware(logger *log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				requestID, ok := r.Context().Value(requestIDKey).(string)
				if !ok {
					requestID = "unknown"
				}
				logger.Println(requestID, r.Method, r.URL.Path, r.RemoteAddr, r.UserAgent())
			}()
			next.ServeHTTP(w, r)
		})
	}
}

func tracingMiddleware(nextRequestID requestIDGenerator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get("X-Request-Id")
			if requestID == "" {
				requestID = nextRequestID()
			}
			ctx := context.WithValue(context.Background(), requestIDKey, requestID)
			w.Header().Set("X-Request-Id", requestID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// func (h *Hub) handleGrpc(conn *websocket.Conn) {
// 	// wrap websocket conn into ReadWriteCloser
// 	wsRwc, err := ws.NewRWC(websocket.BinaryMessage, conn)
// 	if err != nil {
// 		h.logger.Println(err)
// 		return
// 	}

// 	// manage ReadWriteClose using yamux client
// 	incomingConn, err := yamux.Client(wsRwc, yamux.DefaultConfig())
// 	if err != nil {
// 		h.logger.Printf("error creating yamux client: %s", err)
// 	}
// 	defer incomingConn.Close()

// 	// use yamux client as Dialer to grpc
// 	grpcConn, err := grpc.Dial("websocket",
// 		grpc.WithInsecure(),
// 		grpc.WithDialer(func(s string, d time.Duration) (net.Conn, error) {
// 			return incomingConn.Open()
// 		}),
// 	)
// 	if err != nil {
// 		h.logger.Printf("error calling grpc.Dial: %v", err)
// 		return
// 	}

// 	fluentdClient := pb.NewFluentdClient(grpcConn)

// 	ticker := time.NewTicker(1 * time.Second)
// 	defer ticker.Stop()
// 	for {
// 		select {
// 		case <-incomingConn.CloseChan():
// 			h.logger.Println("connecton closed")
// 			return
// 		case <-ticker.C:
// 			h.logger.Println("calling fluentdClient.Start on remote server")
// 			req := pb.FluentdStartRequest{}
// 			resp, err := fluentdClient.Start(context.TODO(), &req)
// 			if err != nil {
// 				h.logger.Printf("error calling fluentdClient.Start: %v", err)
// 				return
// 			}
// 			h.logger.Printf("response: %v", resp)
// 		}
// 	}
// }
