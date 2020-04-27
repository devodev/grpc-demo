package main

import (
	"context"
	"math/rand"
	"time"

	pb "github.com/devodev/grpc-demo/internal/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

func validateTimeoutSec(value int32) error {
	if value < 1 {
		return status.Errorf(codes.InvalidArgument, "timeout_sec must be greater than 0")
	}
	if value > 10 {
		return status.Errorf(codes.InvalidArgument, "timeout_sec must be less than or equal to 10")
	}
	return nil
}

// Start will send a start command and wait for the provided amount of time.
func (s *FluentdService) Start(ctx context.Context, r *pb.FluentdStartRequest) (*pb.FluentdStartResponse, error) {
	if err := validateTimeoutSec(r.TimeoutSec); err != nil {
		return nil, err
	}
	ra := rand.New(rand.NewSource(time.Now().UnixNano()))
	time.Sleep(time.Duration(time.Duration(ra.Intn(int(r.TimeoutSec))) * time.Second))
	return &pb.FluentdStartResponse{Status: pb.FluentdStartResponse_START_SUCCESS}, nil
}

// Stop will send a stop command and wait for the provided amount of time.
func (s *FluentdService) Stop(ctx context.Context, r *pb.FluentdStopRequest) (*pb.FluentdStopResponse, error) {
	if err := validateTimeoutSec(r.TimeoutSec); err != nil {
		return nil, err
	}
	ra := rand.New(rand.NewSource(time.Now().UnixNano()))
	time.Sleep(time.Duration(time.Duration(ra.Intn(int(r.TimeoutSec))) * time.Second))
	return &pb.FluentdStopResponse{Status: pb.FluentdStopResponse_STOP_SUCCESS}, nil
}

// Restart will send a stop command, if started, and then a start command
// and wait for the provided amount of time.
func (s *FluentdService) Restart(ctx context.Context, r *pb.FluentdRestartRequest) (*pb.FluentdRestartResponse, error) {
	if err := validateTimeoutSec(r.TimeoutSec); err != nil {
		return nil, err
	}
	ra := rand.New(rand.NewSource(time.Now().UnixNano()))
	time.Sleep(time.Duration(time.Duration(ra.Intn(int(r.TimeoutSec))) * time.Second))
	return &pb.FluentdRestartResponse{Status: pb.FluentdRestartResponse_RESTART_SUCCESS}, nil
}
