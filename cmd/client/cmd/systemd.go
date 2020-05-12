package cmd

import (
	"context"

	"github.com/devodev/grpc-demo/cmd/client/grpc"
	"github.com/devodev/grpc-demo/internal/server/pb/systemd"

	"github.com/spf13/cobra"
	"google.golang.org/grpc/metadata"
)

func newCommandSystemd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "systemd",
		Short: "Interact with local SystemD daemon.",
	}
	cmd.AddCommand(
		newCommandSystemdListUnits(),
	)
	return cmd
}

func newCommandSystemdListUnits() *cobra.Command {
	dialerCfg := grpc.NewDialerConfig()
	config := grpc.NewConfig()
	cmd := &cobra.Command{
		Use:   "list-units [name]",
		Short: "List unit files.",
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
			client := systemd.NewSystemdClient(conn)

			var v systemd.SystemdListUnitsRequest
			fn := client.ListUnits

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
