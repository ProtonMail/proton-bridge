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

package focus

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/ProtonMail/proton-bridge/v2/internal/focus/proto"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

// TryRaise tries to raise the application by dialing the focus service.
// It returns true if the service is running and the application was told to raise.
func TryRaise() bool {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	cc, err := grpc.DialContext(
		ctx,
		net.JoinHostPort(Host, fmt.Sprint(Port)),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return false
	}

	defer func() {
		if err := cc.Close(); err != nil {
			logrus.WithError(err).Warn("Failed to close focus connection")
		}
	}()

	if _, err := proto.NewFocusClient(cc).Raise(ctx, &emptypb.Empty{}); err != nil {
		return false
	}

	if err := cc.Close(); err != nil {
		return false
	}

	return true
}

// TryRaise tries to raise the application by dialing the focus service.
// It returns true if the service is running and the application was told to raise.
func TryVersion() (*semver.Version, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	cc, err := grpc.DialContext(
		ctx,
		net.JoinHostPort(Host, fmt.Sprint(Port)),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, false
	}

	defer func() {
		if err := cc.Close(); err != nil {
			logrus.WithError(err).Warn("Failed to close focus connection")
		}
	}()

	raw, err := proto.NewFocusClient(cc).Version(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, false
	}

	version, err := semver.NewVersion(raw.GetVersion())
	if err != nil {
		return nil, false
	}

	return version, true
}
