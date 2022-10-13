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
		return bridgeCLI.New(bridge, restarter, eventCh).Loop()

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
