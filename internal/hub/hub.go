package hub

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	api "github.com/devodev/grpc-demo/internal/api/local"
	"github.com/devodev/grpc-demo/internal/client"
	ws "github.com/devodev/grpc-demo/internal/websocket"
	"github.com/gorilla/websocket"
	"github.com/mwitkow/grpc-proxy/proxy"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

var (
	defaultGRPCListenAddr = ":9090"

	defaultHTTPListenAddr   = ":8080"
	defaultHTTPReadTimeout  = 5 * time.Second
	defaultHTTPWriteTimeout = 10 * time.Second
	defaultHTTPIdleTimeout  = 15 * time.Second

	defaultLogOutput = os.Stderr
	defaultLogger    = log.New(defaultLogOutput, "hub: ", log.LstdFlags)

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

// Option provide a way to configure the Hub.
type Option func(*Hub) error

// WithHTTPListenAddr sets the listening address of the HTTP server.
func WithHTTPListenAddr(a string) Option {
	return func(h *Hub) error {
		h.httpListenAddr = a
		return nil
	}
}

// WithGRPCListenAddr sets the listening address of the gRPC server.
func WithGRPCListenAddr(a string) Option {
	return func(h *Hub) error {
		h.grpcListenAddr = a
		return nil
	}
}

// WithTimeouts sets the timeout values of the http server.
func WithTimeouts(read, write, idle time.Duration) Option {
	return func(h *Hub) error {
		h.httpReadTimeout = read
		h.httpWriteTimeout = write
		h.httpIdleTimeout = idle
		return nil
	}
}

// WithMiddlewares adds to the set of middlewares used on the http server.
func WithMiddlewares(mws ...Middleware) Option {
	return func(h *Hub) error {
		h.httpMiddlewares = append(h.httpMiddlewares, mws...)
		return nil
	}
}

// WithTLSConfig sets the tlsConfig of the http server.
func WithTLSConfig(c *tls.Config) Option {
	return func(h *Hub) error {
		h.httpTLSConfig = c
		return nil
	}
}

// WithShutdownTimeout sets the tlsConfig of the http server.
func WithShutdownTimeout(t time.Duration) Option {
	return func(h *Hub) error {
		h.shutdownTimeout = t
		return nil
	}
}

// Hub acts as a gRPC proxy.
//
// It runs an HTTP server exposing a websocket endpoint
// for servers to dial in and register themselves.
//
// It also exposes a gRPC server that let clients
// make requests against remote server services or the hub own services.
//
// For a request to be proxied to a remote server, the client must include
// a gRPC metadata map containing hub specific authentication fields.
//
// Connector is provided as a helper for dialing in and registering to a hub.
// It sets the correct gRPC metadata and returns a listener that can be used
// to serve any gRPC server.
type Hub struct {
	ClientRegistry client.Registry

	logger *log.Logger
	server *http.Server

	grpcListenAddr string

	httpListenAddr   string
	httpReadTimeout  time.Duration
	httpWriteTimeout time.Duration
	httpIdleTimeout  time.Duration
	httpMiddlewares  []Middleware
	httpTLSConfig    *tls.Config

	shutdownTimeout time.Duration

	once       *sync.Once
	closingCh  chan struct{}
	shutdownCh chan struct{}
}

// New .
func New(opts ...Option) (*Hub, error) {
	h := &Hub{
		ClientRegistry: client.NewRegistryMem(),

		logger: defaultLogger,

		grpcListenAddr: defaultGRPCListenAddr,

		httpListenAddr:   defaultHTTPListenAddr,
		httpReadTimeout:  defaultHTTPReadTimeout,
		httpWriteTimeout: defaultHTTPWriteTimeout,
		httpIdleTimeout:  defaultHTTPIdleTimeout,
		httpMiddlewares:  []Middleware{tracingMiddleware, loggingMiddleware(defaultLogger)},

		shutdownTimeout: defaultShutdownTimeout,

		once:       &sync.Once{},
		closingCh:  make(chan struct{}),
		shutdownCh: make(chan struct{}),
	}

	for _, opt := range opts {
		if err := opt(h); err != nil {
			return nil, err
		}
	}

	go h.listenAndServe()
	go h.listenAndServeGRPC()

	return h, nil
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
		if strings.HasPrefix(fullMethodName, "/internal.") {
			return nil, nil, grpc.Errorf(codes.Unimplemented, "Unknown method")
		}
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
			client, err := h.ClientRegistry.Get(name)
			if err != nil {
				return nil, nil, grpc.Errorf(codes.FailedPrecondition, err.Error())
			}
			conn, err := grpc.DialContext(ctx, fullMethodName,
				grpc.WithCodec(proxy.Codec()),
				grpc.WithInsecure(),
				grpc.WithDialer(func(s string, d time.Duration) (net.Conn, error) {
					return client.Session.Open()
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
	hubService := &api.HubService{Registry: h.ClientRegistry}
	hubService.RegisterServer(server)

	go func() {
		<-h.closingCh

		h.logger.Println("grpc server is shutting down..")
		server.GracefulStop()
	}()

	l, err := net.Listen("tcp", h.grpcListenAddr)
	if err != nil {
		h.logger.Fatalf("failed to listen: %v", err)
		return
	}

	h.logger.Printf("gRPC server listening on: %v", h.grpcListenAddr)
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

	h.server = &http.Server{
		Addr:         h.httpListenAddr,
		Handler:      chainMiddlewares(router, h.httpMiddlewares...),
		ErrorLog:     h.logger,
		TLSConfig:    h.httpTLSConfig,
		ReadTimeout:  h.httpReadTimeout,
		WriteTimeout: h.httpWriteTimeout,
		IdleTimeout:  h.httpIdleTimeout,
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
	wsRwc, err := ws.ReadWriteCloser(wsConn)
	if err != nil {
		wsConn.Close()
		h.logger.Println(err)
		return
	}

	metaName := r.Header.Get("X-Hub-Meta-Name")
	cc, err := client.New(wsRwc, metaName)
	if err != nil {
		wsRwc.CloseWithMessage(err.Error())
		h.logger.Println(err)
		if _, ok := err.(*client.ErrEmptyAttribute); ok {
			h.logger.Println("have you set the X-Hub-Meta-* headers?")
		}
		return
	}

	if err := h.ClientRegistry.Register(cc, metaName); err != nil {
		wsRwc.CloseWithMessage(err.Error())
		h.logger.Println(err)
	}
	h.logger.Printf("registered client with name: %v", metaName)

	go func() {
		defer h.ClientRegistry.Unregister(metaName)
		select {
		case <-cc.Session.CloseChan():
			h.logger.Printf("[client: %v] connection closed", metaName)
			return
		}
	}()
}

func loggingMiddleware(logger *log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				requestID := w.Header().Get("X-Request-Id")
				if requestID == "" {
					requestID = "unknown"
				}
				logger.Println(requestID, r.Method, r.URL.Path, r.RemoteAddr, r.UserAgent())
			}()
			next.ServeHTTP(w, r)
		})
	}
}

func tracingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-Id")
		if requestID == "" {
			requestID = strconv.FormatInt(time.Now().UnixNano(), 36)
		}
		w.Header().Set("X-Request-Id", requestID)
		next.ServeHTTP(w, r)
	})
}
