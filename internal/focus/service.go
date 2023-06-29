// Copyright (c) 2023 Proton AG
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
	"github.com/ProtonMail/gluon/async"
	"github.com/ProtonMail/proton-bridge/v3/internal/focus/proto"
	"github.com/ProtonMail/proton-bridge/v3/internal/service"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

const (
	Host                 = "127.0.0.1"
	serverConfigFileName = "grpcFocusServerConfig.json"
)

// Service is a gRPC service that can be used to raise the application.
type Service struct {
	proto.UnimplementedFocusServer

	server  *grpc.Server
	raiseCh chan struct{}
	version *semver.Version

	log          *logrus.Entry
	panicHandler async.PanicHandler
}

// NewService creates a new focus service.
// It listens on the local host and port 1042 (by default).
func NewService(locator service.Locator, version *semver.Version, panicHandler async.PanicHandler) (*Service, error) {
	serv := &Service{
		server:       grpc.NewServer(),
		raiseCh:      make(chan struct{}, 1),
		version:      version,
		log:          logrus.WithField("pkg", "focus/service"),
		panicHandler: panicHandler,
	}

	proto.RegisterFocusServer(serv.server, serv)

	if listener, err := net.Listen("tcp", net.JoinHostPort(Host, fmt.Sprint(0))); err != nil {
		serv.log.WithError(err).Warn("Failed to start focus service")
	} else {
		config := service.Config{}
		// retrieve the port assigned by the system, so that we can put it in the config file.
		address, ok := listener.Addr().(*net.TCPAddr)
		if !ok {
			return nil, fmt.Errorf("could not retrieve gRPC service listener address")
		}
		config.Port = address.Port
		if path, err := service.SaveGRPCServerConfigFile(locator, &config, serverConfigFileName); err != nil {
			serv.log.WithError(err).WithField("path", path).Warn("Could not write focus gRPC service config file")
		} else {
			serv.log.WithField("path", path).Info("Successfully saved gRPC Focus service config file")
		}

		go func() {
			defer async.HandlePanic(serv.panicHandler)

			if err := serv.server.Serve(listener); err != nil {
				fmt.Printf("failed to serve: %v", err)
			}
		}()
	}

	return serv, nil
}

// Raise implements the gRPC FocusService interface; it raises the application.
func (service *Service) Raise(_ context.Context, reason *wrapperspb.StringValue) (*emptypb.Empty, error) {
	service.log.WithField("Reason", reason.Value).Debug("Raise")
	service.raiseCh <- struct{}{}
	return &emptypb.Empty{}, nil
}

// Version implements the gRPC FocusService interface; it returns the version of the service.
func (service *Service) Version(context.Context, *emptypb.Empty) (*proto.VersionResponse, error) {
	service.log.Debug("Version")
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
		defer async.HandlePanic(service.panicHandler)

		// we do this in a goroutine, as on Windows, the gRPC shutdown may take minutes if something tries to
		// interact with it in an invalid way (e.g. HTTP GET request from a Qt QNetworkManager instance).
		service.server.Stop()
		close(service.raiseCh)
	}()
}
