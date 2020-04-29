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
			server := NewServer(config.ListenAddr)

			done := make(chan struct{})
			quit := make(chan os.Signal, 1)
			signal.Notify(quit, os.Interrupt)

			go func() {
				<-quit

				log.Println("server is shutting down...")

				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()

				server.SetKeepAlivesEnabled(false)
				if err := server.Shutdown(ctx); err != nil {
					log.Printf("error during server shutdown: %v", err)
				}
				close(done)
			}()

			log.Printf("listening on: %v", server.Addr)
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Printf("listen error: %v", err)
			}

			<-done
			log.Println("server stopped")
			return nil
		},
	}
	config.AddFlags(cmd.Flags())
	return cmd
}

// NewServer .
func NewServer(addr string) *http.Server {
	hub := &Hub{}
	router := http.NewServeMux()
	router.Handle("/ws", hub)
	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ErrorLog:     nil,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}
	return server
}

// Hub .
type Hub struct {
}

func (h *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{}
	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("upgrade: %v", err)
		return
	}
	go h.handleGrpc(wsConn)
}

func (h *Hub) handleGrpc(conn *websocket.Conn) {
	// wrap websocket conn into ReadWriteCloser
	wsRwc, err := ws.NewRWC(websocket.BinaryMessage, conn)
	if err != nil {
		log.Println(err)
		return
	}

	// manage ReadWriteClose using yamux client
	incomingConn, err := yamux.Client(wsRwc, yamux.DefaultConfig())
	if err != nil {
		log.Printf("error creating yamux client: %s", err)
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
		log.Printf("error calling grpc.Dial: %v", err)
		return
	}

	fluentdClient := pb.NewFluentdClient(grpcConn)

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-incomingConn.CloseChan():
			log.Println("connecton closed")
			return
		case <-ticker.C:
			log.Println("calling fluentdClient.Start on remote server")
			req := pb.FluentdStartRequest{}
			resp, err := fluentdClient.Start(context.TODO(), &req)
			if err != nil {
				log.Printf("error calling fluentdClient.Start: %v", err)

				log.Printf("grpcConn.GetState: %v", grpcConn.GetState())
				log.Printf("incomingConn.IsClosed: %v", incomingConn.IsClosed())

				return
			}
			log.Printf("response: %v", resp)
		}
	}
}
