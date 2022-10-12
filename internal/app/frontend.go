package app

import (
	"fmt"

	"github.com/ProtonMail/proton-bridge/v2/internal/bridge"
	"github.com/ProtonMail/proton-bridge/v2/internal/crash"
	"github.com/ProtonMail/proton-bridge/v2/internal/events"
	bridgeCLI "github.com/ProtonMail/proton-bridge/v2/internal/frontend/cli"
	"github.com/ProtonMail/proton-bridge/v2/internal/frontend/grpc"
	"github.com/ProtonMail/proton-bridge/v2/internal/locations"
	"github.com/ProtonMail/proton-bridge/v2/pkg/restarter"
	"github.com/urfave/cli/v2"
)

func runFrontend(
	c *cli.Context,
	crashHandler *crash.Handler,
	restarter *restarter.Restarter,
	locations *locations.Locations,
	bridge *bridge.Bridge,
	eventCh <-chan events.Event,
) error {
	switch {
	case c.Bool(flagCLI):
		return bridgeCLI.New(bridge, eventCh).Loop()

	case c.Bool(flagNonInteractive):
		select {}

	default:
		service, err := grpc.NewService(crashHandler, restarter, locations, bridge, eventCh, !c.Bool(flagNoWindow))
		if err != nil {
			return fmt.Errorf("could not create service: %w", err)
		}

		return service.Loop()
	}
}
