package grpc

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/spf13/pflag"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/oauth"
)

// DialerConfig .
type DialerConfig struct {
	ServerAddr         string        `envconfig:"SERVER_ADDR" default:"localhost:9090"`
	Timeout            time.Duration `envconfig:"TIMEOUT" default:"10s"`
	TLS                bool          `envconfig:"TLS"`
	ServerName         string        `envconfig:"TLS_SERVER_NAME"`
	InsecureSkipVerify bool          `envconfig:"TLS_INSECURE_SKIP_VERIFY"`
	CACertFile         string        `envconfig:"TLS_CA_CERT_FILE"`
	CertFile           string        `envconfig:"TLS_CERT_FILE"`
	KeyFile            string        `envconfig:"TLS_KEY_FILE"`
	AuthToken          string        `envconfig:"AUTH_TOKEN"`
	AuthTokenType      string        `envconfig:"AUTH_TOKEN_TYPE" default:"Bearer"`
	JWTKey             string        `envconfig:"JWT_KEY"`
	JWTKeyFile         string        `envconfig:"JWT_KEY_FILE"`
}

// NewDialerConfig .
func NewDialerConfig() *DialerConfig {
	return &DialerConfig{}
}

// Dialer .
type Dialer struct {
	*DialerConfig
}

// NewDialer creates a Dialer and applies DialerOptions.
func NewDialer(cfg *DialerConfig) (*Dialer, error) {
	if cfg == nil {
		return nil, fmt.Errorf("DialerConfig must be non-nil")
	}
	d := &Dialer{cfg}
	return d, nil
}

// Dial computes the dialOptions and then call grpc.Dial.
// It returns a grpc.ClientConn.
func (d *Dialer) Dial() (*grpc.ClientConn, error) {
	opts, err := d.options()
	if err != nil {
		return nil, err
	}
	conn, err := grpc.Dial(d.ServerAddr, opts...)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (d *Dialer) options() ([]grpc.DialOption, error) {
	opts := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithTimeout(d.Timeout),
	}
	if d.TLS {
		tlsConfig := &tls.Config{}
		if d.InsecureSkipVerify {
			tlsConfig.InsecureSkipVerify = true
		}
		if d.CACertFile != "" {
			cacert, err := ioutil.ReadFile(d.CACertFile)
			if err != nil {
				return nil, fmt.Errorf("ca cert: %v", err)
			}
			certpool := x509.NewCertPool()
			certpool.AppendCertsFromPEM(cacert)
			tlsConfig.RootCAs = certpool
		}
		if d.CertFile != "" {
			if d.KeyFile == "" {
				return nil, fmt.Errorf("missing key file")
			}
			pair, err := tls.LoadX509KeyPair(d.CertFile, d.KeyFile)
			if err != nil {
				return nil, fmt.Errorf("cert/key: %v", err)
			}
			tlsConfig.Certificates = []tls.Certificate{pair}
		}
		if d.ServerName != "" {
			tlsConfig.ServerName = d.ServerName
		} else {
			addr, _, _ := net.SplitHostPort(d.ServerAddr)
			tlsConfig.ServerName = addr
		}
		cred := credentials.NewTLS(tlsConfig)
		opts = append(opts, grpc.WithTransportCredentials(cred))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}
	if d.AuthToken != "" {
		cred := oauth.NewOauthAccess(&oauth2.Token{
			AccessToken: d.AuthToken,
			TokenType:   d.AuthTokenType,
		})
		opts = append(opts, grpc.WithPerRPCCredentials(cred))
	}
	if d.JWTKey != "" {
		cred, err := oauth.NewJWTAccessFromKey([]byte(d.JWTKey))
		if err != nil {
			return nil, fmt.Errorf("jwt key: %v", err)
		}
		opts = append(opts, grpc.WithPerRPCCredentials(cred))
	}
	if d.JWTKeyFile != "" {
		cred, err := oauth.NewJWTAccessFromFile(d.JWTKeyFile)
		if err != nil {
			return nil, fmt.Errorf("jwt key file: %v", err)
		}
		opts = append(opts, grpc.WithPerRPCCredentials(cred))
	}
	return opts, nil
}

// AddFlags adds flags to the provided flagset.
func (d *DialerConfig) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&d.ServerAddr, "server-addr", "s", d.ServerAddr, "server address in form of host:port")
	fs.DurationVar(&d.Timeout, "timeout", d.Timeout, "client connection timeout")
	fs.BoolVar(&d.TLS, "tls", d.TLS, "enable tls")
	fs.StringVar(&d.ServerName, "tls-server-name", d.ServerName, "tls server name override")
	fs.BoolVar(&d.InsecureSkipVerify, "tls-insecure-skip-verify", d.InsecureSkipVerify, "INSECURE: skip tls checks")
	fs.StringVar(&d.CACertFile, "tls-ca-cert-file", d.CACertFile, "ca certificate file")
	fs.StringVar(&d.CertFile, "tls-cert-file", d.CertFile, "client certificate file")
	fs.StringVar(&d.KeyFile, "tls-key-file", d.KeyFile, "client key file")
	fs.StringVar(&d.AuthToken, "auth-token", d.AuthToken, "authorization token")
	fs.StringVar(&d.AuthTokenType, "auth-token-type", d.AuthTokenType, "authorization token type")
	fs.StringVar(&d.JWTKey, "jwt-key", d.JWTKey, "jwt key")
	fs.StringVar(&d.JWTKeyFile, "jwt-key-file", d.JWTKeyFile, "jwt key file")
}

// ProcessEnv uses envconfig to fill Dialer attributes using
// environment variables.
func (d *DialerConfig) ProcessEnv() error {
	return envconfig.Process("", d)
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
type RoundTripFunc func(cfg *Config, in Decoder, out Encoder) error

// RoundTrip .
func (c *Config) RoundTrip(fn RoundTripFunc) error {
	// select encoder
	em := DefaultEncoders["json"]
	if c.ResponseFormat != "" {
		var ok bool
		em, ok = DefaultEncoders[c.ResponseFormat]
		if !ok {
			return fmt.Errorf("invalid response format: %q", c.ResponseFormat)
		}
	}
	e := em.NewEncoder(os.Stdout)
	// select decoder
	d := DefaultDecoders["json"].NewDecoder(os.Stdin)
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
		dm, ok := DefaultDecoders[ext]
		if !ok {
			return fmt.Errorf("invalid request file format: %q", ext)
		}
		d = dm.NewDecoder(f)
	}
	return fn(c, d, e)
}
