package hub

import (
	"context"
	"time"

	"github.com/devodev/grpc-demo/internal/client"
	"github.com/devodev/grpc-demo/internal/feed"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Service implements Hub.
type Service struct {
	Registry     client.Registry
	ActivityFeed *feed.Feed
}

// RegisterServer resgisters itself to a grpc server.
func (s *Service) RegisterServer(server *grpc.Server) {
	RegisterHubServer(server, s)
}

// ListClients returns the list of clients currently connected to the hub, as well as the total count.
func (s *Service) ListClients(ctx context.Context, r *HubListClientsRequest) (*HubListClientsResponse, error) {
	now := time.Now()

	clientList := s.Registry.List()

	var clients []*Client
	for _, client := range clientList {
		cclient := client
		c := &Client{
			Name:           cclient.Name,
			ConnectionTime: cclient.ConnectionTime.String(),
			Uptime:         now.Sub(cclient.ConnectionTime).String(),
		}
		clients = append(clients, c)
	}
	return &HubListClientsResponse{Count: int64(len(clients)), Clients: clients}, nil
}

// StreamActivityFeed returns a stream of ActivityEvent.
func (s *Service) StreamActivityFeed(req *HubActivityFeedRequest, server Hub_StreamActivityFeedServer) error {
	quit := make(chan struct{})
	defer close(quit)

	ch := s.ActivityFeed.GetCh(quit)
	for message := range ch {
		if err := server.Send(&ActivityEvent{Message: message}); err != nil {
			return status.Errorf(codes.Aborted, "error: %v", err)
		}
	}
	return nil
}
