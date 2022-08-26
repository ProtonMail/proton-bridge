package focus

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/ProtonMail/proton-bridge/v2/internal/focus/proto"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TryRaise() bool {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	cc, err := grpc.DialContext(ctx, net.JoinHostPort(Host, fmt.Sprint(Port)), grpc.WithInsecure())
	if err != nil {
		return false
	}

	if _, err := proto.NewFocusClient(cc).Raise(ctx, &emptypb.Empty{}); err != nil {
		return false
	}

	if err := cc.Close(); err != nil {
		return false
	}

	return true
}
