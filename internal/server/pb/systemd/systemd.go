package systemd

import (
	context "context"

	"github.com/coreos/go-systemd/v22/dbus"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Service implements Systemd.
type Service struct {
}

// RegisterServer resgisters itself to a grpc server.
func (s *Service) RegisterServer(server *grpc.Server) {
	RegisterSystemdServer(server, s)
}

// ListUnits .
func (s *Service) ListUnits(ctx context.Context, r *SystemdListUnitsRequest) (*SystemdListUnitsResponse, error) {
	conn, err := dbus.NewSystemConnection()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer conn.Close()

	statuses, err := conn.ListUnitFiles()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	var units []*UnitFile
	for _, s := range statuses {
		units = append(units, &UnitFile{
			Path: s.Path,
			Type: s.Type,
		})
	}
	return &SystemdListUnitsResponse{Units: units}, nil
}

// Status .
func (s *Service) Status(ctx context.Context, r *SystemdStatusRequest) (*SystemdStatusResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

// ListStatus .
func (s *Service) ListStatus(ctx context.Context, r *SystemdListStatusRequest) (*SystemdListStatusResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

// Enable .
func (s *Service) Enable(ctx context.Context, r *SystemdEnableRequest) (*SystemdEnableResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

// Disable .
func (s *Service) Disable(ctx context.Context, r *SystemdDisableRequest) (*SystemdDisableResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

// Start .
func (s *Service) Start(ctx context.Context, r *SystemdStartRequest) (*SystemdStartResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

// Stop .
func (s *Service) Stop(ctx context.Context, r *SystemdStopRequest) (*SystemdStopResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

// Restart .
func (s *Service) Restart(ctx context.Context, r *SystemdRestartRequest) (*SystemdRestartResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

// Reload .
func (s *Service) Reload(ctx context.Context, r *SystemdReloadRequest) (*SystemdReloadResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

// Kill .
func (s *Service) Kill(ctx context.Context, r *SystemdKillRequest) (*SystemdKillResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}
