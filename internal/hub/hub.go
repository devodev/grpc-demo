package hub

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
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

type middleware func(http.Handler) http.Handler

func chainMiddlewares(h http.Handler, m ...middleware) http.Handler {
	if len(m) < 1 {
		return h
	}
	wrapped := h
	for i := len(m) - 1; i >= 0; i-- {
		wrapped = m[i](wrapped)
	}
	return wrapped
}

type requestIDGenerator func() string

// Hub .
type Hub struct {
	healthy int64

	logger *log.Logger
	server *http.Server

	nextRequestID requestIDGenerator

	registry *registry
}

// New .
func New(addr string, l *log.Logger) *Hub {
	if l == nil {
		l = log.New(os.Stderr, "hub: ", log.LstdFlags)
	}
	h := &Hub{
		logger:        l,
		nextRequestID: func() string { return strconv.FormatInt(time.Now().UnixNano(), 36) },
		registry:      newRegistry(),
	}
	router := http.NewServeMux()
	router.HandleFunc("/api/list", h.handleListClients)
	router.HandleFunc("/health", h.healthz)
	router.HandleFunc("/ws", h.handleWS)
	h.server = &http.Server{
		Addr:         addr,
		Handler:      chainMiddlewares(router, h.tracingMiddleware, h.loggingMiddleware),
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

		atomic.StoreInt64(&h.healthy, 0)
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
		atomic.StoreInt64(&h.healthy, time.Now().UnixNano())

		h.logger.Printf("listening on: %v", h.server.Addr)
		if err := h.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			h.logger.Printf("listen error: %v", err)
		}
	}()
	return done
}

func (h *Hub) handleListClients(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "", http.StatusMethodNotAllowed)
		return
	}
	clientCount := h.registry.count()
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Number of clients: %+v", clientCount)
}

func (h *Hub) handleWS(w http.ResponseWriter, r *http.Request) {
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

	client := newClient(session)
	clientID := h.registry.registerClient(client)

	go func() {
		defer h.registry.unregisterClient(clientID)
		select {
		case <-session.CloseChan():
			h.logger.Printf("client%d: connecton closed", clientID)
			return
		}
	}()
}

func (h *Hub) healthz(w http.ResponseWriter, req *http.Request) {
	if health := atomic.LoadInt64(&h.healthy); health != 0 {
		fmt.Fprintf(w, "uptime: %s\n", time.Since(time.Unix(0, health)))
		return
	}
	w.WriteHeader(http.StatusServiceUnavailable)
}

func (h *Hub) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			requestID := w.Header().Get("X-Request-Id")
			if requestID == "" {
				requestID = "unknown"
			}
			h.logger.Println(requestID, r.Method, r.URL.Path, r.RemoteAddr, r.UserAgent())
		}()
		next.ServeHTTP(w, r)
	})
}

func (h *Hub) tracingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-Id")
		if requestID == "" {
			requestID = h.nextRequestID()
		}
		w.Header().Set("X-Request-Id", requestID)
		next.ServeHTTP(w, r)
	})
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
