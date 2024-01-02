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

package app

import (
	"fmt"

	"github.com/ProtonMail/proton-bridge/v3/internal/bridge"
	"github.com/ProtonMail/proton-bridge/v3/internal/crash"
	"github.com/ProtonMail/proton-bridge/v3/internal/events"
	bridgeCLI "github.com/ProtonMail/proton-bridge/v3/internal/frontend/cli"
	"github.com/ProtonMail/proton-bridge/v3/internal/frontend/grpc"
	"github.com/ProtonMail/proton-bridge/v3/internal/locations"
	"github.com/ProtonMail/proton-bridge/v3/pkg/restarter"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func runFrontend(
	c *cli.Context,
	crashHandler *crash.Handler,
	restarter *restarter.Restarter,
	locations *locations.Locations,
	bridge *bridge.Bridge,
	eventCh <-chan events.Event,
	quitCh <-chan struct{},
	parentPID int,
) error {
	logrus.Debug("Running frontend")
	defer logrus.Debug("Frontend stopped")

	switch {
	case c.Bool(flagCLI):
		return bridgeCLI.New(bridge, restarter, eventCh, crashHandler, quitCh).Loop()

	case c.Bool(flagNonInteractive):
		<-quitCh
		return nil

	case c.Bool(flagGRPC):
		service, err := grpc.NewService(crashHandler, restarter, locations, bridge, eventCh, quitCh, !c.Bool(flagNoWindow), parentPID)
		if err != nil {
			return fmt.Errorf("could not create service: %w", err)
		}

		return service.Loop()

	default:
		if err := cli.ShowAppHelp(c); err != nil {
			logrus.WithError(err).Error("Failed to show app help")
		}

		return fmt.Errorf("no frontend specified, use --cli, --grpc or --noninteractive")
	}
}
