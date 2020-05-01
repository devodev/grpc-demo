package hub

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/hashicorp/yamux"
)

var (
	defaultListenAddr = ":8080"
	defaultLogOutput  = os.Stderr

	defaultReadTimeout     = 5 * time.Second
	defaultWriteTimeout    = 10 * time.Second
	defaultIdleTimeout     = 15 * time.Second
	defaultShutdownTimeout = 30 * time.Second
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

// RequestIDGenerator is used by the hub registry
// to generate new client ids.
type RequestIDGenerator func() string

// Hub .
type Hub struct {
	healthy int64

	logger *log.Logger
	server *http.Server

	nextRequestID RequestIDGenerator

	registry *registry

	once       *sync.Once
	closingCh  chan struct{}
	shutdownCh chan struct{}
}

// Config holds the Hub configuration.
type Config struct {
	// ListenAddr is the address on which the server listens.
	ListenAddr string
	// Logger is used to provide a custom logger.
	Logger *log.Logger
	// RequestIDGenerator is used by the Hub registry to generate
	// new client ids.
	RequestIDGenerator RequestIDGenerator

	// ReadTimeout is provided to the underlying http server.
	ReadTimeout time.Duration
	// WriteTimeout is provided to the underlying http server.
	WriteTimeout time.Duration
	// IdleTimeout is provided to the underlying http server.
	IdleTimeout time.Duration
}

// New .
func New(cfg *Config) *Hub {
	if cfg == nil {
		cfg = &Config{}
	}
	addr := cfg.ListenAddr
	if addr == "" {
		addr = defaultListenAddr
	}
	logger := cfg.Logger
	if logger == nil {
		logger = log.New(defaultLogOutput, "hub: ", log.LstdFlags)
	}
	nextRequestID := cfg.RequestIDGenerator
	if nextRequestID == nil {
		nextRequestID = func() string { return strconv.FormatInt(time.Now().UnixNano(), 36) }
	}
	h := &Hub{
		logger:        logger,
		nextRequestID: nextRequestID,
		registry:      newRegistry(),
		once:          &sync.Once{},
		closingCh:     make(chan struct{}),
		shutdownCh:    make(chan struct{}),
	}
	router := http.NewServeMux()
	router.HandleFunc("/api/list", h.handleListClients)
	router.HandleFunc("/health", h.handleHealth)
	router.HandleFunc("/ws", h.handleWS)
	router.HandleFunc("/", h.handleGRPC)
	middlewares := []middleware{h.tracingMiddleware, h.loggingMiddleware}
	h.server = &http.Server{
		Addr:         addr,
		Handler:      chainMiddlewares(router, middlewares...),
		ErrorLog:     h.logger,
		ReadTimeout:  defaultReadTimeout,
		WriteTimeout: defaultWriteTimeout,
		IdleTimeout:  defaultIdleTimeout,
	}
	go h.listenAndServe()
	go h.listenAndServeGRPC()
	return h
}

// Close .
func (h *Hub) Close() {
	h.once.Do(func() {
		close(h.closingCh)
		<-h.shutdownCh
	})
}

func (h *Hub) listenAndServeGRPC() {
	l, err := net.Listen("tcp", ":9090")
	if err != nil {
		h.logger.Printf("could not listen: %v", err)
		return
	}
	defer l.Close()
	for {
		conn, err := l.Accept()
		if err != nil {
			conn.Close()
			h.logger.Printf("error on Accept: %v", err)
			return
		}
		h.logger.Println("accepted connection")
		clientID := uint64(1)
		client, err := h.registry.get(clientID)
		if err != nil {
			conn.Close()
			h.logger.Printf("error on registry.get: %v", err)
			return
		}
		clientConn, err := client.session.Open()
		if err != nil {
			clientConn.Close()
			conn.Close()
			h.logger.Printf("error on client.session.Accept: %v", err)
			return
		}
		go func(c1, c2 net.Conn) {
			defer func() {
				h.logger.Printf("connection to client %v leaving", clientID)
				c1.Close()
				c2.Close()
			}()
			h.logger.Printf("connected incoming call to client %v", clientID)
			Pipe(c1, c2)
		}(conn, clientConn)
	}
}

// This currently is not working since http2 request are only handled
// on servers using TLS.
// The server responds with PRI to the client, indicating it tried to handle an HTTP2
// request using an HTTP/1.1 enabled server.
func (h *Hub) handleGRPC(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "webserver doesn't support hijacking", http.StatusInternalServerError)
		return
	}
	conn, _, err := hj.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.logger.Println("accepted connection")
	clientID := uint64(1)
	client, err := h.registry.get(clientID)
	if err != nil {
		conn.Close()
		h.logger.Printf("error on registry.get: %v", err)
		return
	}
	clientConn, err := client.session.Open()
	if err != nil {
		clientConn.Close()
		conn.Close()
		h.logger.Printf("error on client.session.Accept: %v", err)
		return
	}
	go func(c1, c2 net.Conn) {
		defer func() {
			h.logger.Printf("connection to client %v leaving", clientID)
			c1.Close()
			c2.Close()
		}()
		h.logger.Printf("connected incoming call to client %v", clientID)
		Pipe(c1, c2)
	}(conn, clientConn)
}

func (h *Hub) listenAndServe() {
	go func() {
		<-h.closingCh

		atomic.StoreInt64(&h.healthy, 0)
		h.logger.Println("hub is shutting down...")

		ctx, cancel := context.WithTimeout(context.Background(), defaultShutdownTimeout)
		defer cancel()

		h.server.SetKeepAlivesEnabled(false)
		if err := h.server.Shutdown(ctx); err != nil {
			h.logger.Printf("error during server shutdown: %v", err)
		}
		close(h.shutdownCh)
	}()

	atomic.StoreInt64(&h.healthy, time.Now().UnixNano())

	h.logger.Printf("listening on: %v", h.server.Addr)
	if err := h.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		h.logger.Printf("listen error: %v", err)
	}
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

	// grpcConn, err := grpc.Dial("websocket",
	// 	grpc.WithInsecure(),
	// 	grpc.WithDialer(func(s string, d time.Duration) (net.Conn, error) {
	// 		return session.Open()
	// 	}),
	// )
	// if err != nil {
	// 	h.logger.Printf("error calling grpc.Dial: %v", err)
	// 	return
	// }

	// fluentdClient := pb.NewFluentdClient(grpcConn)

	// ticker := time.NewTicker(1 * time.Second)
	// defer ticker.Stop()
	// for {
	// 	select {
	// 	case <-session.CloseChan():
	// 		h.logger.Println("connecton closed")
	// 		return
	// 	case <-ticker.C:
	// 		h.logger.Println("calling fluentdClient.Start on remote server")
	// 		req := pb.FluentdStartRequest{}
	// 		resp, err := fluentdClient.Start(context.TODO(), &req)
	// 		if err != nil {
	// 			h.logger.Printf("error calling fluentdClient.Start: %v", err)
	// 			return
	// 		}
	// 		h.logger.Printf("response: %v", resp)
	// 	}
	// }
}

func (h *Hub) handleHealth(w http.ResponseWriter, req *http.Request) {
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
