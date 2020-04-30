package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/kelseyhightower/envconfig"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/devodev/grpc-demo/internal/hub"
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

			h := hub.New(config.ListenAddr, logger)

			select {
			case <-quit:
				h.Close()
			}
			logger.Println("stopped")
			return nil
		},
	}
	config.AddFlags(cmd.Flags())
	return cmd
}
