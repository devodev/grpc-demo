package main

import (
	"log"
	"net/url"
	"os"
	"os/signal"

	"github.com/gorilla/websocket"
	"github.com/hashicorp/yamux"
	"github.com/kelseyhightower/envconfig"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"google.golang.org/grpc"

	"github.com/devodev/grpc-demo/internal/api"
	"github.com/devodev/grpc-demo/internal/hub"
)

// ServerConfig holds config for the Fluentd command.
type ServerConfig struct {
	HubAddr string `envconfig:"HUB_ADDR" default:"ws://localhost:8080/ws"`
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
	fs.StringVar(&c.HubAddr, "hub-uri", c.HubAddr, "hub websocket uri.")
}

func newCommandServe() *cobra.Command {
	config := NewServerConfig()
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "serve the gRPC server.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			interrupt := make(chan os.Signal, 1)
			signal.Notify(interrupt, os.Interrupt)

			u, err := url.Parse(config.HubAddr)
			if err != nil {
				return err
			}
			wsConn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
			if err != nil {
				return err
			}

			wsRwc, err := hub.NewRWC(websocket.BinaryMessage, wsConn)
			if err != nil {
				return err
			}

			srvConn, err := yamux.Server(wsRwc, yamux.DefaultConfig())
			if err != nil {
				return err
			}

			server := grpc.NewServer()
			fluentdService := &api.FluentdService{}
			fluentdService.RegisterServer(server)

			go func() {
				defer func() {
					log.Println("graceful shutdown..")
					server.GracefulStop()
				}()
				select {
				case <-interrupt:
				}
			}()

			if err := server.Serve(srvConn); err != nil && err != grpc.ErrServerStopped {
				log.Fatal(err)
				return err
			}
			return nil
		},
	}
	config.AddFlags(cmd.Flags())
	return cmd
}
