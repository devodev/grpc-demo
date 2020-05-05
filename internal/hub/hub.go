package hub

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mwitkow/grpc-proxy/proxy"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

var (
	defaultHTTPListenAddr     = ":8080"
	defaultGRPCListenAddr     = ":9090"
	defaultLogOutput          = os.Stderr
	defaultRequestIDGenerator = func() string { return strconv.FormatInt(time.Now().UnixNano(), 36) }

	defaultReadTimeout     = 5 * time.Second
	defaultWriteTimeout    = 10 * time.Second
	defaultIdleTimeout     = 15 * time.Second
	defaultShutdownTimeout = 30 * time.Second
)

// Middleware is used to decorate an http.Handler.
type Middleware func(http.Handler) http.Handler

func chainMiddlewares(h http.Handler, m ...Middleware) http.Handler {
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

// Config holds the Hub configuration.
type Config struct {
	// Logger is used to provide a custom logger.
	Logger *log.Logger
	// RequestIDGenerator is used by the tracing middleware.
	RequestIDGenerator RequestIDGenerator
	// Registry is used to store clients.
	Registry ClientRegistry

	// HTTPListenAddr is the address on which the HTTP server listens.
	HTTPListenAddr string
	// GRPCListenAddr is the address on which the GRPC server listens.
	GRPCListenAddr string
	// TLSConfig is provided to the underlying http server.
	TLSConfig *tls.Config
	// Middlewares are chained and applied on the main server router.
	Middlewares []Middleware
	// ReadTimeout is provided to the underlying http server.
	ReadTimeout time.Duration
	// WriteTimeout is provided to the underlying http server.
	WriteTimeout time.Duration
	// IdleTimeout is provided to the underlying http server.
	IdleTimeout time.Duration
}

// Hub .
type Hub struct {
	config *Config
	logger *log.Logger
	server *http.Server

	nextRequestID RequestIDGenerator

	clientRegistry ClientRegistry

	once       *sync.Once
	closingCh  chan struct{}
	shutdownCh chan struct{}
}

// New .
func New(cfg *Config) *Hub {
	if cfg == nil {
		cfg = &Config{}
	}
	HTTPAddr := cfg.HTTPListenAddr
	if HTTPAddr == "" {
		HTTPAddr = defaultHTTPListenAddr
	}
	GRPCAddr := cfg.GRPCListenAddr
	if GRPCAddr == "" {
		GRPCAddr = defaultGRPCListenAddr
	}
	logger := cfg.Logger
	if logger == nil {
		logger = log.New(defaultLogOutput, "hub: ", log.LstdFlags)
	}
	nextRequestID := cfg.RequestIDGenerator
	if nextRequestID == nil {
		nextRequestID = defaultRequestIDGenerator
	}
	registry := cfg.Registry
	if registry == nil {
		registry = NewRegistryMem()
	}

	h := &Hub{
		config:         cfg,
		logger:         logger,
		nextRequestID:  nextRequestID,
		clientRegistry: registry,
		once:           &sync.Once{},
		closingCh:      make(chan struct{}),
		shutdownCh:     make(chan struct{}),
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
	director := func(ctx context.Context, fullMethodName string) (context.Context, *grpc.ClientConn, error) {
		if strings.HasPrefix(fullMethodName, "/external.") {
			md, ok := metadata.FromIncomingContext(ctx)
			if !ok {
				return nil, nil, grpc.Errorf(codes.FailedPrecondition, "no metadata provided")
			}
			nameList, ok := md["name"]
			if !ok {
				return nil, nil, grpc.Errorf(codes.FailedPrecondition, "name not found in metadata")
			}
			name := nameList[0]
			client, err := h.clientRegistry.Get(name)
			if err != nil {
				return nil, nil, grpc.Errorf(codes.FailedPrecondition, err.Error())
			}
			conn, err := grpc.DialContext(ctx, fullMethodName,
				grpc.WithCodec(proxy.Codec()),
				grpc.WithInsecure(),
				grpc.WithDialer(func(s string, d time.Duration) (net.Conn, error) {
					return client.session.Open()
				}),
			)
			h.logger.Printf("proxying gRPC request (%v) to: %v", fullMethodName, name)
			return ctx, conn, err
		}
		return nil, nil, grpc.Errorf(codes.Unimplemented, "Unknown method")
	}

	server := grpc.NewServer(
		grpc.CustomCodec(proxy.Codec()),
		grpc.UnknownServiceHandler(proxy.TransparentHandler(director)),
	)

	go func() {
		<-h.closingCh

		h.logger.Println("grpc server is shutting down..")
		server.GracefulStop()
	}()

	l, err := net.Listen("tcp", h.config.GRPCListenAddr)
	if err != nil {
		h.logger.Fatalf("failed to listen: %v", err)
		return
	}

	h.logger.Printf("gRPC server listening on: %v", h.config.GRPCListenAddr)
	if err := server.Serve(l); err != nil && err != grpc.ErrServerStopped {
		h.logger.Printf("gRPC server listen error: %v", err)
	}
}

func (h *Hub) listenAndServe() {
	var healthy int64
	handleHealth := func(w http.ResponseWriter, req *http.Request) {
		if health := atomic.LoadInt64(&healthy); health != 0 {
			fmt.Fprintf(w, "uptime: %s\n", time.Since(time.Unix(0, health)))
			return
		}
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	router := http.NewServeMux()
	router.HandleFunc("/health", handleHealth)
	router.HandleFunc("/ws", h.handleWS)

	defaultMiddlewares := []Middleware{h.tracingMiddleware, h.loggingMiddleware}
	if len(h.config.Middlewares) > 0 {
		defaultMiddlewares = append(defaultMiddlewares, h.config.Middlewares...)
	}

	h.server = &http.Server{
		Addr:         h.config.HTTPListenAddr,
		Handler:      chainMiddlewares(router, defaultMiddlewares...),
		ErrorLog:     h.logger,
		TLSConfig:    h.config.TLSConfig,
		ReadTimeout:  defaultReadTimeout,
		WriteTimeout: defaultWriteTimeout,
		IdleTimeout:  defaultIdleTimeout,
	}

	go func() {
		<-h.closingCh

		atomic.StoreInt64(&healthy, 0)
		h.logger.Println("HTTP server is shutting down...")

		ctx, cancel := context.WithTimeout(context.Background(), defaultShutdownTimeout)
		defer cancel()

		h.server.SetKeepAlivesEnabled(false)
		if err := h.server.Shutdown(ctx); err != nil {
			h.logger.Printf("error during server shutdown: %v", err)
		}
		close(h.shutdownCh)
	}()

	atomic.StoreInt64(&healthy, time.Now().UnixNano())

	h.logger.Printf("HTTP server listening on: %v", h.server.Addr)
	if h.server.TLSConfig != nil {
		if err := h.server.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
			h.logger.Printf("HTTP server listen error: %v", err)
		}
	} else {
		if err := h.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			h.logger.Printf("HTTP server listen error: %v", err)
		}
	}
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
		wsConn.Close()
		h.logger.Println(err)
		return
	}

	metaName := r.Header.Get("X-Hub-Meta-Name")
	client, err := NewClient(wsRwc, metaName)
	if err != nil {
		wsRwc.CloseWithMessage(err.Error())
		h.logger.Println(err)
		if _, ok := err.(*errEmptyAttribute); ok {
			h.logger.Println("have you set the X-Hub-Meta-* headers?")
		}
		return
	}

	if err := h.clientRegistry.Register(client, metaName); err != nil {
		wsRwc.CloseWithMessage(err.Error())
		h.logger.Println(err)
	}
	h.logger.Printf("registered client with name: %v", metaName)

	go func() {
		defer h.clientRegistry.Unregister(metaName)
		select {
		case <-client.session.CloseChan():
			h.logger.Printf("[client: %v] connection closed", metaName)
			return
		}
	}()
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

// CreateServerTLSConfig creates a tls config using the provided.
func CreateServerTLSConfig(caPath, certPath, keyPath string) (*tls.Config, error) {
	tlsConfig := &tls.Config{}
	tlsConfig.PreferServerCipherSuites = true
	if caPath != "" {
		cacert, err := ioutil.ReadFile(caPath)
		if err != nil {
			return nil, fmt.Errorf("ca cert: %v", err)
		}
		certpool := x509.NewCertPool()
		certpool.AppendCertsFromPEM(cacert)
		tlsConfig.RootCAs = certpool
	}
	if certPath == "" {
		return nil, fmt.Errorf("missing cert file")
	}
	if keyPath == "" {
		return nil, fmt.Errorf("missing key file")
	}
	pair, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return nil, fmt.Errorf("cert/key: %v", err)
	}
	tlsConfig.Certificates = []tls.Certificate{pair}
	return tlsConfig, nil
}
