package focus

import (
	"context"
	"fmt"
	"net"

	"github.com/ProtonMail/proton-bridge/v2/internal/focus/proto"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	Host = "127.0.0.1"
	Port = 1042
)

type FocusService struct {
	proto.UnimplementedFocusServer

	server   *grpc.Server
	listener net.Listener
	raiseCh  chan struct{}
}

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

func (service *FocusService) Raise(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	service.raiseCh <- struct{}{}
	return &emptypb.Empty{}, nil
}

func (service *FocusService) GetRaiseCh() <-chan struct{} {
	return service.raiseCh
}

func (service *FocusService) Close() {
	service.server.Stop()
}
