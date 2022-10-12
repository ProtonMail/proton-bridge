// Package focus provides a gRPC service for raising the application.
package focus

import (
	"context"
	"fmt"
	"net"

	"github.com/Masterminds/semver/v3"
	"github.com/ProtonMail/proton-bridge/v2/internal/focus/proto"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Host is the local host to listen on.
const Host = "127.0.0.1"

// Port is the port to listen on.
var Port = 1042 // nolint:gochecknoglobals

// Service is a gRPC service that can be used to raise the application.
type Service struct {
	proto.UnimplementedFocusServer

	server   *grpc.Server
	listener net.Listener
	raiseCh  chan struct{}
	version  *semver.Version
}

// NewService creates a new focus service.
// It listens on the local host and port 1042 (by default).
func NewService(version *semver.Version) (*Service, error) {
	listener, err := net.Listen("tcp", net.JoinHostPort(Host, fmt.Sprint(Port)))
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %w", err)
	}

	service := &Service{
		server:   grpc.NewServer(),
		listener: listener,
		raiseCh:  make(chan struct{}, 1),
		version:  version,
	}

	proto.RegisterFocusServer(service.server, service)

	go func() {
		if err := service.server.Serve(listener); err != nil {
			fmt.Printf("failed to serve: %v", err)
		}
	}()

	return service, nil
}

// Raise implements the gRPC FocusService interface; it raises the application.
func (service *Service) Raise(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	service.raiseCh <- struct{}{}
	return &emptypb.Empty{}, nil
}

// Version implements the gRPC FocusService interface; it returns the version of the service.
func (service *Service) Version(context.Context, *emptypb.Empty) (*proto.VersionResponse, error) {
	return &proto.VersionResponse{
		Version: service.version.Original(),
	}, nil
}

// GetRaiseCh returns a channel on which events are sent when the application should be raised.
func (service *Service) GetRaiseCh() <-chan struct{} {
	return service.raiseCh
}

// Close closes the service.
func (service *Service) Close() {
	service.server.Stop()
	close(service.raiseCh)
}
