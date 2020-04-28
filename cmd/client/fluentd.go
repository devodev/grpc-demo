package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/devodev/grpc-demo/cmd/client/codec"
	pb "github.com/devodev/grpc-demo/internal/pb"
	"github.com/kelseyhightower/envconfig"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/oauth"
)

// TransportConfig holds network transport configuration and utility methods.
type TransportConfig struct {
	ServerAddr         string        `envconfig:"SERVER_ADDR" default:"localhost:8080"`
	Timeout            time.Duration `envconfig:"TIMEOUT" default:"10s"`
	TLS                bool          `envconfig:"TLS"`
	ServerName         string        `envconfig:"TLS_SERVER_NAME"`
	InsecureSkipVerify bool          `envconfig:"TLS_INSECURE_SKIP_VERIFY"`
	CACertFile         string        `envconfig:"TLS_CA_CERT_FILE"`
	CertFile           string        `envconfig:"TLS_CERT_FILE"`
	KeyFile            string        `envconfig:"TLS_KEY_FILE"`
}

// NewTransportConfig returns TransportConfig after being processed
// using envconfig to set default values and environment variables.
func NewTransportConfig() *TransportConfig {
	c := &TransportConfig{}
	envconfig.Process("", c)
	return c
}

// AddFlags adds flags to the provided flagset.
func (c *TransportConfig) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&c.ServerAddr, "server-addr", "s", c.ServerAddr, "server address in form of host:port")
	fs.DurationVar(&c.Timeout, "timeout", c.Timeout, "client connection timeout")
	fs.BoolVar(&c.TLS, "tls", c.TLS, "enable tls")
	fs.StringVar(&c.ServerName, "tls-server-name", c.ServerName, "tls server name override")
	fs.BoolVar(&c.InsecureSkipVerify, "tls-insecure-skip-verify", c.InsecureSkipVerify, "INSECURE: skip tls checks")
	fs.StringVar(&c.CACertFile, "tls-ca-cert-file", c.CACertFile, "ca certificate file")
	fs.StringVar(&c.CertFile, "tls-cert-file", c.CertFile, "client certificate file")
	fs.StringVar(&c.KeyFile, "tls-key-file", c.KeyFile, "client key file")
}

// Options returns the options computed from TransportConfig
// or an error if invalid config provided.
func (c *TransportConfig) Options() ([]grpc.DialOption, error) {
	opts := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithTimeout(c.Timeout),
	}
	if c.TLS {
		tlsConfig := &tls.Config{}
		if c.InsecureSkipVerify {
			tlsConfig.InsecureSkipVerify = true
		}
		if c.CACertFile != "" {
			cacert, err := ioutil.ReadFile(c.CACertFile)
			if err != nil {
				return nil, fmt.Errorf("ca cert: %v", err)
			}
			certpool := x509.NewCertPool()
			certpool.AppendCertsFromPEM(cacert)
			tlsConfig.RootCAs = certpool
		}
		if c.CertFile != "" {
			if c.KeyFile == "" {
				return nil, fmt.Errorf("missing key file")
			}
			pair, err := tls.LoadX509KeyPair(c.CertFile, c.KeyFile)
			if err != nil {
				return nil, fmt.Errorf("cert/key: %v", err)
			}
			tlsConfig.Certificates = []tls.Certificate{pair}
		}
		if c.ServerName != "" {
			tlsConfig.ServerName = c.ServerName
		} else {
			addr, _, _ := net.SplitHostPort(c.ServerAddr)
			tlsConfig.ServerName = addr
		}
		cred := credentials.NewTLS(tlsConfig)
		opts = append(opts, grpc.WithTransportCredentials(cred))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}
	return opts, nil
}

// AuthConfig holds authentication configuration and utility methods.
type AuthConfig struct {
	AuthToken     string `envconfig:"AUTH_TOKEN"`
	AuthTokenType string `envconfig:"AUTH_TOKEN_TYPE" default:"Bearer"`
	JWTKey        string `envconfig:"JWT_KEY"`
	JWTKeyFile    string `envconfig:"JWT_KEY_FILE"`
}

// NewAuthConfig returns AuthConfig after being processed
// using envconfig to set default values and environment variables.
func NewAuthConfig() *AuthConfig {
	c := &AuthConfig{}
	envconfig.Process("", c)
	return c
}

// AddFlags adds flags to the provided flagset.
func (c *AuthConfig) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&c.AuthToken, "auth-token", c.AuthToken, "authorization token")
	fs.StringVar(&c.AuthTokenType, "auth-token-type", c.AuthTokenType, "authorization token type")
	fs.StringVar(&c.JWTKey, "jwt-key", c.JWTKey, "jwt key")
	fs.StringVar(&c.JWTKeyFile, "jwt-key-file", c.JWTKeyFile, "jwt key file")
}

// Options returns the options computed from TransportConfig
// or an error if invalid config provided.
func (c *AuthConfig) Options() ([]grpc.DialOption, error) {
	opts := []grpc.DialOption{}
	if c.AuthToken != "" {
		cred := oauth.NewOauthAccess(&oauth2.Token{
			AccessToken: c.AuthToken,
			TokenType:   c.AuthTokenType,
		})
		opts = append(opts, grpc.WithPerRPCCredentials(cred))
	}
	if c.JWTKey != "" {
		cred, err := oauth.NewJWTAccessFromKey([]byte(c.JWTKey))
		if err != nil {
			return nil, fmt.Errorf("jwt key: %v", err)
		}
		opts = append(opts, grpc.WithPerRPCCredentials(cred))
	}
	if c.JWTKeyFile != "" {
		cred, err := oauth.NewJWTAccessFromFile(c.JWTKeyFile)
		if err != nil {
			return nil, fmt.Errorf("jwt key file: %v", err)
		}
		opts = append(opts, grpc.WithPerRPCCredentials(cred))
	}
	return opts, nil
}

// Config holds config for the Fluentd command.
type Config struct {
	RequestFile        string `envconfig:"REQUEST_FILE"`
	PrintSampleRequest bool   `envconfig:"PRINT_SAMPLE_REQUEST"`
	ResponseFormat     string `envconfig:"RESPONSE_FORMAT" default:"json"`
}

// NewConfig returns Config after being processed
// using envconfig to set default values and environment variables.
func NewConfig() *Config {
	c := &Config{}
	envconfig.Process("", c)
	return c
}

// AddFlags adds flags to the provided flagset.
func (c *Config) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&c.RequestFile, "request-file", "f", c.RequestFile, "client request file (must be json, yaml, or xml); use \"-\" for stdin + json")
	fs.BoolVarP(&c.PrintSampleRequest, "print-sample-request", "p", c.PrintSampleRequest, "print sample request file and exit")
	fs.StringVarP(&c.ResponseFormat, "response-format", "o", c.ResponseFormat, "response format (json, prettyjson, yaml, or xml)")
}

// RoundTripFunc .
type RoundTripFunc func(cfg *Config, in codec.Decoder, out codec.Encoder) error

// RoundTrip .
func (c *Config) RoundTrip(fn RoundTripFunc) error {
	// select encoder
	em := codec.DefaultEncoders["json"]
	if c.ResponseFormat != "" {
		var ok bool
		em, ok = codec.DefaultEncoders[c.ResponseFormat]
		if !ok {
			return fmt.Errorf("invalid response format: %q", c.ResponseFormat)
		}
	}
	e := em.NewEncoder(os.Stdout)
	// select decoder
	d := codec.DefaultDecoders["json"].NewDecoder(os.Stdin)
	if c.RequestFile != "" && c.RequestFile != "-" {
		f, err := os.Open(c.RequestFile)
		if err != nil {
			return fmt.Errorf("request file: %v", err)
		}
		defer f.Close()
		ext := filepath.Ext(c.RequestFile)
		if len(ext) > 0 && ext[0] == '.' {
			ext = ext[1:]
		}
		var ok bool
		dm, ok := codec.DefaultDecoders[ext]
		if !ok {
			return fmt.Errorf("invalid request file format: %q", ext)
		}
		d = dm.NewDecoder(f)
	}
	return fn(c, d, e)
}

func newCommandFluentd() *cobra.Command {
	transportCfg := NewTransportConfig()
	authCfg := NewAuthConfig()
	config := NewConfig()
	cmd := &cobra.Command{
		Use:   "fluentd [method]",
		Short: "Calls the FluentdService on the gRPC server.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			transportOpts, err := transportCfg.Options()
			if err != nil {
				return err
			}
			authOpts, err := authCfg.Options()
			if err != nil {
				return err
			}
			conn, err := grpc.Dial(transportCfg.ServerAddr, append(transportOpts, authOpts...)...)
			if err != nil {
				log.Fatal(err)
			}
			defer conn.Close()
			client := pb.NewFluentdClient(conn)

			switch args[0] {
			default:
				log.Fatal("unsupported method")
			case "start":
				var v pb.FluentdStartRequest
				fn := client.Start
				config.RoundTrip(func(cfg *Config, in codec.Decoder, out codec.Encoder) error {
					if cfg.PrintSampleRequest {
						return out.Encode(&v)
					}
					err := in.Decode(&v)
					if err != nil {
						return err
					}
					resp, err := fn(context.Background(), &v)
					if err != nil {
						return err
					}
					return out.Encode(resp)
				})
			case "stop":
				var v pb.FluentdStopRequest
				fn := client.Stop
				config.RoundTrip(func(cfg *Config, in codec.Decoder, out codec.Encoder) error {
					if cfg.PrintSampleRequest {
						return out.Encode(&v)
					}
					err := in.Decode(&v)
					if err != nil {
						return err
					}
					resp, err := fn(context.Background(), &v)
					if err != nil {
						return err
					}
					return out.Encode(resp)
				})
			case "restart":
				var v pb.FluentdRestartRequest
				fn := client.Restart
				config.RoundTrip(func(cfg *Config, in codec.Decoder, out codec.Encoder) error {
					if cfg.PrintSampleRequest {
						return out.Encode(&v)
					}
					err := in.Decode(&v)
					if err != nil {
						return err
					}
					resp, err := fn(context.Background(), &v)
					if err != nil {
						return err
					}
					return out.Encode(resp)
				})
			}
			return nil
		},
	}
	transportCfg.AddFlags(cmd.Flags())
	authCfg.AddFlags(cmd.Flags())
	config.AddFlags(cmd.Flags())
	return cmd
}
