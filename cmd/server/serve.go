package main

import (
	"net"

	"github.com/kelseyhightower/envconfig"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"google.golang.org/grpc"
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
		Short: "serve the gRPC server.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			l, err := net.Listen("tcp", config.ListenAddr)
			if err != nil {
				return err
			}
			server := grpc.NewServer()
			fluentdService := FluentdService{}
			fluentdService.RegisterServer(server)
			return server.Serve(l)
		},
	}
	config.AddFlags(cmd.Flags())
	return cmd
}
