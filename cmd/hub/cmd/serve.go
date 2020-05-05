package cmd

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"

	"github.com/kelseyhightower/envconfig"
	"github.com/spf13/cobra"

	"github.com/devodev/grpc-demo/internal/hub"
)

// serverConfig holds serverConfig for the Fluentd command.
type serverConfig struct {
	HTTPListenAddr string `envconfig:"HTTP_LISTEN_ADDR" default:":8080"`
	GRPCListenAddr string `envconfig:"GRPC_LISTEN_ADDR" default:":9090"`
	TLS            bool   `envconfig:"TLS"`
	CACertFile     string `envconfig:"TLS_CA_CERT_FILE"`
	CertFile       string `envconfig:"TLS_CERT_FILE"`
	KeyFile        string `envconfig:"TLS_KEY_FILE"`
}

// setupCmd sets flags on the provided cmd and resolve env variables using the provided Config.
func setupCmd(cmd *cobra.Command, c *serverConfig) *cobra.Command {
	envconfig.Process("", c)
	cmd.Flags().StringVar(&c.HTTPListenAddr, "http-listen", c.HTTPListenAddr, "HTTP server listening address.")
	cmd.Flags().StringVar(&c.GRPCListenAddr, "grpc-listen", c.GRPCListenAddr, "GRPC server listening address.")
	cmd.Flags().BoolVar(&c.TLS, "tls", c.TLS, "enable tls")
	cmd.Flags().StringVar(&c.CACertFile, "tls-ca-cert-file", c.CACertFile, "ca certificate file")
	cmd.Flags().StringVar(&c.CertFile, "tls-cert-file", c.CertFile, "certificate file")
	cmd.Flags().StringVar(&c.KeyFile, "tls-key-file", c.KeyFile, "key file")
	return cmd
}

func makeTLSConfig(caPath, certPath, keyPath string) (*tls.Config, error) {
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

func newCommandServe() *cobra.Command {
	var cfg *serverConfig
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
			if cfg.TLS {
				var err error
				tlsConfig, err = makeTLSConfig(cfg.CACertFile, cfg.CertFile, cfg.KeyFile)
				if err != nil {
					return err
				}
			}

			hubCfg := &hub.Config{
				HTTPListenAddr: cfg.HTTPListenAddr,
				GRPCListenAddr: cfg.GRPCListenAddr,
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
	return setupCmd(cmd, cfg)
}
