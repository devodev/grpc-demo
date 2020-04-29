package main

import (
	"context"
	"log"
	"net"
	"net/http"
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
			http.HandleFunc("/ws", echo)
			return http.ListenAndServe(config.ListenAddr, nil)
		},
	}
	config.AddFlags(cmd.Flags())
	return cmd
}

func echo(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{}
	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("upgrade: %v", err)
		return
	}

	wsRwc, err := ws.NewRWC(websocket.BinaryMessage, wsConn,
		ws.WithPingEnabled(), ws.WithPongHandler())
	if err != nil {
		log.Println(err)
		return
	}

	incomingConn, err := yamux.Client(wsRwc, yamux.DefaultConfig())
	if err != nil {
		log.Printf("error creating yamux client: %s", err)
	}
	defer incomingConn.Close()

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

	req := pb.FluentdStartRequest{}
	resp, err := fluentdClient.Start(context.TODO(), &req)
	if err != nil {
		log.Printf("error calling fluentdClient.Start: %v", err)
		return
	}
	log.Printf("response: %v", resp)
}
