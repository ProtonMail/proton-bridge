// Copyright (c) 2022 Proton AG
//
// This file is part of Proton Mail Bridge.
//
// Proton Mail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Proton Mail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Proton Mail Bridge.  If not, see <https://www.gnu.org/licenses/>.

// Package focus provides a gRPC service for raising the application.
package focus

import (
	"context"
	"fmt"
	"net"

	"github.com/Masterminds/semver/v3"
	"github.com/ProtonMail/proton-bridge/v2/internal/focus/proto"
	"github.com/sirupsen/logrus"
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

	server  *grpc.Server
	raiseCh chan struct{}
	version *semver.Version
}

// NewService creates a new focus service.
// It listens on the local host and port 1042 (by default).
func NewService(version *semver.Version) (*Service, error) {
	service := &Service{
		server:  grpc.NewServer(),
		raiseCh: make(chan struct{}, 1),
		version: version,
	}

	proto.RegisterFocusServer(service.server, service)

	if listener, err := net.Listen("tcp", net.JoinHostPort(Host, fmt.Sprint(Port))); err != nil {
		logrus.WithError(err).Warn("Failed to start focus service")
	} else {
		go func() {
			if err := service.server.Serve(listener); err != nil {
				fmt.Printf("failed to serve: %v", err)
			}
		}()
	}

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
	go func() {
		// we do this in a goroutine, as on Windows, the gRPC shutdown may take minutes if something tries to
		// interact with it in an invalid way (e.g. HTTP GET request from a Qt QNetworkManager instance).
		service.server.Stop()
		close(service.raiseCh)
	}()
}
