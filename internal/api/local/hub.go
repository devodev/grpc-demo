package local

import (
	"context"
	"time"

	"github.com/devodev/grpc-demo/internal/client"
	pb "github.com/devodev/grpc-demo/internal/pb/local"

	"google.golang.org/grpc"
)

// HubService implements pb.Hub.
type HubService struct {
	Registry client.Registry
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
