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
