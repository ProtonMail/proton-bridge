// Copyright (c) 2024 Proton AG
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

package focus

import (
	"context"
	"fmt"
	"net"
	"path/filepath"

	"github.com/Masterminds/semver/v3"
	"github.com/ProtonMail/proton-bridge/v3/internal/focus/proto"
	"github.com/ProtonMail/proton-bridge/v3/internal/service"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// TryRaise tries to raise the application by dialing the focus service.
// It returns true if the service is running and the application was told to raise.
func TryRaise(settingsPath string) bool {
	var raised bool

	if err := withClientConn(context.Background(), settingsPath, func(ctx context.Context, client proto.FocusClient) error {
		if _, err := client.Raise(ctx, &wrapperspb.StringValue{Value: "TryRaise"}); err != nil {
			return fmt.Errorf("failed to call client.Raise: %w", err)
		}

		raised = true

		return nil
	}); err != nil {
		logrus.WithError(err).Debug("Failed to raise application")
		return false
	}

	return raised
}

// TryVersion tries to determine the version of the running application instance.
// It returns the version and true if the version could be determined.
func TryVersion(settingsPath string) (*semver.Version, bool) {
	var version *semver.Version

	if err := withClientConn(context.Background(), settingsPath, func(ctx context.Context, client proto.FocusClient) error {
		raw, err := client.Version(ctx, &emptypb.Empty{})
		if err != nil {
			return fmt.Errorf("failed to call client.Version: %w", err)
		}

		parsed, err := semver.NewVersion(raw.GetVersion())
		if err != nil {
			return fmt.Errorf("failed to parse version: %w", err)
		}

		version = parsed

		return nil
	}); err != nil {
		logrus.WithError(err).Debug("Failed to determine version of running instance")
		return nil, false
	}

	return version, true
}

func withClientConn(ctx context.Context, settingsPath string, fn func(context.Context, proto.FocusClient) error) error {
	var config = service.Config{}
	err := config.Load(filepath.Join(settingsPath, serverConfigFileName))
	if err != nil {
		return err
	}
	cc, err := grpc.DialContext(
		ctx,
		net.JoinHostPort(Host, fmt.Sprint(config.Port)),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return err
	}

	defer func() {
		if err := cc.Close(); err != nil {
			logrus.WithError(err).Warn("Failed to close focus connection")
		}
	}()

	return fn(ctx, proto.NewFocusClient(cc))
}
