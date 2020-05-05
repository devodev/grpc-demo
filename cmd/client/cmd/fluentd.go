package cmd

import (
	"context"

	pb "github.com/devodev/grpc-demo/internal/pb/external"
	"google.golang.org/grpc/metadata"

	"github.com/devodev/grpc-demo/cmd/client/grpc"
	"github.com/spf13/cobra"
)

func newCommandFluentd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fluentd",
		Short: "Interact with the FluentD service.",
	}
	cmd.AddCommand(
		newCommandFluentdStart(),
		newCommandFluentdStop(),
		newCommandFluentdRestart(),
	)
	return cmd
}

func newCommandFluentdStart() *cobra.Command {
	dialerCfg := grpc.NewDialerConfig()
	config := grpc.NewConfig()
	cmd := &cobra.Command{
		Use:   "start [name]",
		Short: "Start the Fluentd service.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			md := metadata.New(map[string]string{"name": name})
			mdCtx := metadata.NewOutgoingContext(context.Background(), md)

			dialer, err := grpc.NewDialer(dialerCfg)
			if err != nil {
				return err
			}
			conn, err := dialer.Dial()
			if err != nil {
				return err
			}
			defer conn.Close()
			client := pb.NewFluentdClient(conn)

			var v pb.FluentdStartRequest
			fn := client.Start

			// get encoder and decoder

			// decode request

			// call grpc

			// encode response

			return config.RoundTrip(func(cfg *grpc.Config, in grpc.Decoder, out grpc.Encoder) error {
				if cfg.PrintSampleRequest {
					return out.Encode(&v)
				}
				err := in.Decode(&v)
				if err != nil {
					return err
				}
				resp, err := fn(mdCtx, &v)
				if err != nil {
					return err
				}
				return out.Encode(resp)
			})
		},
	}
	cmd.Flags().SortFlags = false
	dialerCfg.ProcessEnv()
	dialerCfg.AddFlags(cmd.Flags())
	config.AddFlags(cmd.Flags())
	return cmd
}

func newCommandFluentdStop() *cobra.Command {
	dialerCfg := grpc.NewDialerConfig()
	config := grpc.NewConfig()
	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stops the Fluentd service.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			dialer, err := grpc.NewDialer(dialerCfg)
			if err != nil {
				return err
			}
			conn, err := dialer.Dial()
			if err != nil {
				return err
			}
			defer conn.Close()
			client := pb.NewFluentdClient(conn)

			var v pb.FluentdStopRequest
			fn := client.Stop
			return config.RoundTrip(func(cfg *grpc.Config, in grpc.Decoder, out grpc.Encoder) error {
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
		},
	}
	cmd.Flags().SortFlags = false
	dialerCfg.ProcessEnv()
	dialerCfg.AddFlags(cmd.Flags())
	config.AddFlags(cmd.Flags())
	return cmd
}

func newCommandFluentdRestart() *cobra.Command {
	dialerCfg := grpc.NewDialerConfig()
	config := grpc.NewConfig()
	cmd := &cobra.Command{
		Use:   "restart",
		Short: "Restarts the Fluentd service.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			dialer, err := grpc.NewDialer(dialerCfg)
			if err != nil {
				return err
			}
			conn, err := dialer.Dial()
			if err != nil {
				return err
			}
			defer conn.Close()
			client := pb.NewFluentdClient(conn)

			var v pb.FluentdRestartRequest
			fn := client.Restart
			return config.RoundTrip(func(cfg *grpc.Config, in grpc.Decoder, out grpc.Encoder) error {
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
		},
	}
	cmd.Flags().SortFlags = false
	dialerCfg.ProcessEnv()
	dialerCfg.AddFlags(cmd.Flags())
	config.AddFlags(cmd.Flags())
	return cmd
}
