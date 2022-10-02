package focus

import (
	"context"
	"fmt"
	"net"

	"github.com/ProtonMail/proton-bridge/v2/internal/focus/proto"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Host is the local host to listen on.
const Host = "127.0.0.1"

// Port is the port to listen on.
var Port = 1042

// FocusService is a gRPC service that can be used to raise the application.
type FocusService struct {
	proto.UnimplementedFocusServer

	server   *grpc.Server
	listener net.Listener
	raiseCh  chan struct{}
}

// NewService creates a new focus service.
// It listens on the local host and port 1042 (by default).
func NewService() (*FocusService, error) {
	listener, err := net.Listen("tcp", net.JoinHostPort(Host, fmt.Sprint(Port)))
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %w", err)
	}

	service := &FocusService{
		server:   grpc.NewServer(),
		listener: listener,
		raiseCh:  make(chan struct{}, 1),
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
func (service *FocusService) Raise(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	service.raiseCh <- struct{}{}
	return &emptypb.Empty{}, nil
}

// GetRaiseCh returns a channel on which events are sent when the application should be raised.
func (service *FocusService) GetRaiseCh() <-chan struct{} {
	return service.raiseCh
}

// Close closes the service.
func (service *FocusService) Close() {
	service.server.Stop()
}
