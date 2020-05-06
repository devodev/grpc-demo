package local

import (
	"context"
	"time"

	"github.com/devodev/grpc-demo/internal/client"
	"github.com/devodev/grpc-demo/internal/feed"
	pb "github.com/devodev/grpc-demo/internal/pb/local"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// HubService implements pb.Hub.
type HubService struct {
	Registry     client.Registry
	ActivityFeed *feed.Feed
}

// RegisterServer resgisters itself to a grpc server.
func (s *HubService) RegisterServer(server *grpc.Server) {
	pb.RegisterHubServer(server, s)
}

// ListClients returns the list of clients currently connected to the hub, as well as the total count.
func (s *HubService) ListClients(ctx context.Context, r *pb.HubListClientsRequest) (*pb.HubListClientsResponse, error) {
	now := time.Now()

	clientList := s.Registry.List()

	var clients []*pb.Client
	for _, client := range clientList {
		cclient := client
		c := &pb.Client{
			Name:           cclient.Name,
			ConnectionTime: cclient.ConnectionTime.String(),
			Uptime:         now.Sub(cclient.ConnectionTime).String(),
		}
		clients = append(clients, c)
	}
	return &pb.HubListClientsResponse{Count: int64(len(clients)), Clients: clients}, nil
}

// StreamActivityFeed returns a stream of ActivityEvent.
func (s *HubService) StreamActivityFeed(req *pb.HubActivityFeedRequest, server pb.Hub_StreamActivityFeedServer) error {
	quit := make(chan struct{})
	defer close(quit)

	ch := s.ActivityFeed.GetCh(quit)
	for message := range ch {
		if err := server.Send(&pb.ActivityEvent{Message: message}); err != nil {
			return status.Errorf(codes.Aborted, "error: %v", err)
		}
	}
	return nil
}
