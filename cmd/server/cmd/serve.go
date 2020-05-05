package cmd

import (
	"log"
	"os"
	"os/signal"

	"github.com/kelseyhightower/envconfig"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	api "github.com/devodev/grpc-demo/internal/api/remote"
	"github.com/devodev/grpc-demo/internal/hub"
)

// Config holds config for the Fluentd command.
type Config struct {
	HubAddr            string `envconfig:"HUB_ADDR" default:"ws://localhost:8080/ws"`
	InsecureSkipVerify bool   `envconfig:"TLS_INSECURE_SKIP_VERIFY"`
}

// SetupCmd set flags on the provided cmd and resolve env variables using the provided Config.
func SetupCmd(cmd *cobra.Command, c *Config) *cobra.Command {
	envconfig.Process("", c)
	cmd.Flags().StringVar(&c.HubAddr, "hub-uri", c.HubAddr, "hub websocket uri.")
	cmd.Flags().BoolVar(&c.InsecureSkipVerify, "tls-insecure-skip-verify", c.InsecureSkipVerify, "INSECURE: skip tls checks")
	return cmd
}

func newCommandServe() *cobra.Command {
	var config Config

	cmd := &cobra.Command{
		Use:   "serve [name]",
		Short: "serve the gRPC server.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			// ? TODO: I dont think we need to have a struct.
			// ? TODO: Could move validation inside Dial()
			// ? TODO: The use case I see that might be useful is
			// ? TODO: provide helper methods to set hub specific headers.
			hubDialer, err := hub.NewConnector(config.HubAddr, config.InsecureSkipVerify, name)
			if err != nil {
				return err
			}
			hubListener, err := hubDialer.Listener()
			if err != nil {
				return err
			}

			interrupt := make(chan os.Signal, 1)
			signal.Notify(interrupt, os.Interrupt)

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

			if err := server.Serve(hubListener); err != nil && err != grpc.ErrServerStopped {
				log.Fatal(err)
				return err
			}
			return nil
		},
	}
	cmd = SetupCmd(cmd, &config)
	return cmd
}
