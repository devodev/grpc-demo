package remote

import (
	"context"

	pb "github.com/devodev/grpc-demo/internal/pb/remote"
	"google.golang.org/grpc"
)

// FluentdService implements pb.Fluentd.
// It provides service level methods for managing
// a fluentd systemd service.
type FluentdService struct {
}

// RegisterServer resgisters itself to a grpc server.
func (s *FluentdService) RegisterServer(server *grpc.Server) {
	pb.RegisterFluentdServer(server, s)
}

// Start will send a start command and wait for the provided amount of time.
func (s *FluentdService) Start(ctx context.Context, r *pb.FluentdStartRequest) (*pb.FluentdStartResponse, error) {
	return &pb.FluentdStartResponse{Status: pb.FluentdStartResponse_START_SUCCESS}, nil
}

// Stop will send a stop command and wait for the provided amount of time.
func (s *FluentdService) Stop(ctx context.Context, r *pb.FluentdStopRequest) (*pb.FluentdStopResponse, error) {
	return &pb.FluentdStopResponse{Status: pb.FluentdStopResponse_STOP_SUCCESS}, nil
}

// Restart will send a stop command, if started, and then a start command
// and wait for the provided amount of time.
func (s *FluentdService) Restart(ctx context.Context, r *pb.FluentdRestartRequest) (*pb.FluentdRestartResponse, error) {
	return &pb.FluentdRestartResponse{Status: pb.FluentdRestartResponse_RESTART_SUCCESS}, nil
}
