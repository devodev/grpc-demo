package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
	"github.com/hashicorp/yamux"
	"github.com/kelseyhightower/envconfig"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"google.golang.org/grpc"

	"github.com/devodev/grpc-demo/internal/pb"
	"github.com/devodev/grpc-demo/internal/ws"
)

// ServerConfig holds config for the Fluentd command.
type ServerConfig struct {
	ListenAddr string `envconfig:"LISTEN_ADDR" default:":8080"`
}

// NewServerConfig returns ServerConfig after being processed
// using envconfig to set default values and environment variables.
func NewServerConfig() *ServerConfig {
	c := &ServerConfig{}
	envconfig.Process("", c)
	return c
}

// AddFlags adds flags to the provided flagset.
func (c *ServerConfig) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&c.ListenAddr, "listen", c.ListenAddr, "listening address.")
}

func newCommandServe() *cobra.Command {
	config := NewServerConfig()
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "serve the gRPC hub.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			quit := make(chan os.Signal, 1)
			signal.Notify(quit, os.Interrupt)

			logger := log.New(os.Stderr, "hub: ", log.LstdFlags)

			hub := NewHub(config.ListenAddr, logger)
			done := hub.StartWeb(quit)

			<-done
			logger.Println("stopped")
			return nil
		},
	}
	config.AddFlags(cmd.Flags())
	return cmd
}

// Client .
type Client struct {
}

// Hub .
type Hub struct {
	// used to generate client ids
	counter uint64
	// Registered clients.
	clients map[uint64]*Client
	// Register requests from the clients.
	register chan *Client
	// Unregister requests from clients.
	unregister chan *Client

	logger *log.Logger
	server *http.Server
}

// NewHub .
func NewHub(addr string, l *log.Logger) *Hub {
	if l == nil {
		l = log.New(os.Stderr, "hub: ", log.LstdFlags)
	}
	h := &Hub{
		clients:    make(map[uint64]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		logger:     l,
	}
	router := http.NewServeMux()
	router.HandleFunc("/", h.httpHandler())
	router.HandleFunc("/ws", h.wsHandler())
	h.server = &http.Server{
		Addr:         addr,
		Handler:      router,
		ErrorLog:     h.logger,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}
	return h
}

// StartWeb is a long running goroutine.
func (h *Hub) StartWeb(quit chan os.Signal) <-chan struct{} {
	done := make(chan struct{})

	go func() {
		<-quit

		h.logger.Println("hub is shutting down...")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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

func (h *Hub) httpHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "unsupported method", http.StatusMethodNotAllowed)
			return
		}
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
		go h.handleGrpc(wsConn)
	}
}

func (h *Hub) handleGrpc(conn *websocket.Conn) {
	// wrap websocket conn into ReadWriteCloser
	wsRwc, err := ws.NewRWC(websocket.BinaryMessage, conn)
	if err != nil {
		h.logger.Println(err)
		return
	}

	// manage ReadWriteClose using yamux client
	incomingConn, err := yamux.Client(wsRwc, yamux.DefaultConfig())
	if err != nil {
		h.logger.Printf("error creating yamux client: %s", err)
	}
	defer incomingConn.Close()

	// use yamux client as Dialer to grpc
	grpcConn, err := grpc.Dial("websocket",
		grpc.WithInsecure(),
		grpc.WithDialer(func(s string, d time.Duration) (net.Conn, error) {
			return incomingConn.Open()
		}),
	)
	if err != nil {
		h.logger.Printf("error calling grpc.Dial: %v", err)
		return
	}

	fluentdClient := pb.NewFluentdClient(grpcConn)

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-incomingConn.CloseChan():
			h.logger.Println("connecton closed")
			return
		case <-ticker.C:
			h.logger.Println("calling fluentdClient.Start on remote server")
			req := pb.FluentdStartRequest{}
			resp, err := fluentdClient.Start(context.TODO(), &req)
			if err != nil {
				h.logger.Printf("error calling fluentdClient.Start: %v", err)
				return
			}
			h.logger.Printf("response: %v", resp)
		}
	}
}
