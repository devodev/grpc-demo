package main

import (
	"crypto/tls"
	"os"
	"os/signal"

	"github.com/kelseyhightower/envconfig"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/devodev/grpc-demo/internal/hub"
)

// ServerConfig holds config for the Fluentd command.
type ServerConfig struct {
	HTTPListenAddr string `envconfig:"HTTP_LISTEN_ADDR" default:":8080"`
	GRPCListenAddr string `envconfig:"GRPC_LISTEN_ADDR" default:":9090"`
	TLS            bool   `envconfig:"TLS"`
	CACertFile     string `envconfig:"TLS_CA_CERT_FILE"`
	CertFile       string `envconfig:"TLS_CERT_FILE"`
	KeyFile        string `envconfig:"TLS_KEY_FILE"`
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
	fs.StringVar(&c.HTTPListenAddr, "http-listen", c.HTTPListenAddr, "HTTP server listening address.")
	fs.StringVar(&c.GRPCListenAddr, "grpc-listen", c.GRPCListenAddr, "GRPC server listening address.")
	fs.BoolVar(&c.TLS, "tls", c.TLS, "enable tls")
	fs.StringVar(&c.CACertFile, "tls-ca-cert-file", c.CACertFile, "ca certificate file")
	fs.StringVar(&c.CertFile, "tls-cert-file", c.CertFile, "certificate file")
	fs.StringVar(&c.KeyFile, "tls-key-file", c.KeyFile, "key file")
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

			// TODO: refactor this using options.
			// TODO: something like: WithTLS(caPath string, certPath string, keyPath string)
			var tlsConfig *tls.Config
			if config.TLS {
				var err error
				tlsConfig, err = hub.CreateServerTLSConfig(config.CACertFile, config.CertFile, config.KeyFile)
				if err != nil {
					return err
				}
			}

			hubCfg := &hub.Config{
				HTTPListenAddr: config.HTTPListenAddr,
				GRPCListenAddr: config.GRPCListenAddr,
				TLSConfig:      tlsConfig,
			}
			h := hub.New(hubCfg)

			select {
			case <-quit:
				h.Close()
			}
			return nil
		},
	}
	config.AddFlags(cmd.Flags())
	return cmd
}
