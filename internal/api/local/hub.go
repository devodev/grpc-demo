package local

import (
	"context"

	"github.com/devodev/grpc-demo/internal/client"
	pb "github.com/devodev/grpc-demo/internal/pb/local"

	"google.golang.org/grpc"
)

// ClientService implements pb.Hub.
type ClientService struct {
	Registry client.Registry
}

// RegisterServer resgisters itself to a grpc server.
func (s *ClientService) RegisterServer(server *grpc.Server) {
	pb.RegisterClientServer(server, s)
}

// List returns the list of clients currently connected to the hub, as well as the total count.
func (s *ClientService) List(ctx context.Context, r *pb.ClientListRequest) (*pb.ClientListResponse, error) {
	nameList := s.Registry.List()
	return &pb.ClientListResponse{Count: int64(len(nameList)), Names: nameList}, nil
}
