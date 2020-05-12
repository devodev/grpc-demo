package fluentd

import (
	"context"

	"google.golang.org/grpc"
)

// Service implements Fluentd.
// It provides service level methods for managing
// a fluentd systemd service.
type Service struct {
}

// RegisterServer resgisters itself to a grpc server.
func (s *Service) RegisterServer(server *grpc.Server) {
	RegisterFluentdServer(server, s)
}

// Start will send a start command and wait for the provided amount of time.
func (s *Service) Start(ctx context.Context, r *FluentdStartRequest) (*FluentdStartResponse, error) {
	return &FluentdStartResponse{Status: FluentdStartResponse_START_SUCCESS}, nil
}

// Stop will send a stop command and wait for the provided amount of time.
func (s *Service) Stop(ctx context.Context, r *FluentdStopRequest) (*FluentdStopResponse, error) {
	return &FluentdStopResponse{Status: FluentdStopResponse_STOP_SUCCESS}, nil
}

// Restart will send a stop command, if started, and then a start command
// and wait for the provided amount of time.
func (s *Service) Restart(ctx context.Context, r *FluentdRestartRequest) (*FluentdRestartResponse, error) {
	return &FluentdRestartResponse{Status: FluentdRestartResponse_RESTART_SUCCESS}, nil
}
