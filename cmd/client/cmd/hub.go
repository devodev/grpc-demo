package cmd

import (
	"context"
	"io"

	"github.com/devodev/grpc-demo/cmd/client/grpc"
	pb "github.com/devodev/grpc-demo/internal/pb/local"

	"github.com/spf13/cobra"
)

func newCommandHub() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hub",
		Short: "Interact with the Hub service.",
	}
	cmd.AddCommand(
		newCommandHubListClients(),
		newCommandHubActivityFeed(),
	)
	return cmd
}

func newCommandHubListClients() *cobra.Command {
	dialerCfg := grpc.NewDialerConfig()
	config := grpc.NewConfig()
	cmd := &cobra.Command{
		Use:   "list-clients",
		Short: "List clients connected to hub.",
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
			hubClient := pb.NewHubClient(conn)

			var v pb.HubListClientsRequest
			fn := hubClient.ListClients

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

func newCommandHubActivityFeed() *cobra.Command {
	dialerCfg := grpc.NewDialerConfig()
	config := grpc.NewConfig()
	cmd := &cobra.Command{
		Use:   "activity-feed",
		Short: "Stream the hub activity feed.",
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
			hubClient := pb.NewHubClient(conn)

			var v pb.HubActivityFeedRequest
			fn := hubClient.StreamActivityFeed

			return config.RoundTrip(func(cfg *grpc.Config, in grpc.Decoder, out grpc.Encoder) error {
				if cfg.PrintSampleRequest {
					return out.Encode(&v)
				}
				if err := in.Decode(&v); err != nil {
					return err
				}
				feed, err := fn(context.Background(), &v)
				if err != nil {
					return err
				}
				defer feed.CloseSend()

				var message pb.ActivityEvent
				for {
					if err := feed.RecvMsg(&message); err != nil {
						if err == io.EOF {
							break
						}
						return err
					}
					if err := out.Encode(&message); err != nil {
						return err
					}
				}
				return nil
			})
		},
	}
	cmd.Flags().SortFlags = false
	dialerCfg.ProcessEnv()
	dialerCfg.AddFlags(cmd.Flags())
	config.AddFlags(cmd.Flags())
	return cmd
}
