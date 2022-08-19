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
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.

// Package frontend provides all interfaces of the Bridge.
package frontend

import (
	"github.com/ProtonMail/proton-bridge/v2/internal/bridge"
	"github.com/ProtonMail/proton-bridge/v2/internal/config/settings"
	"github.com/ProtonMail/proton-bridge/v2/internal/frontend/cli"
	"github.com/ProtonMail/proton-bridge/v2/internal/frontend/grpc"
	"github.com/ProtonMail/proton-bridge/v2/internal/frontend/types"
	"github.com/ProtonMail/proton-bridge/v2/internal/updater"
	"github.com/ProtonMail/proton-bridge/v2/pkg/listener"
)

type Frontend interface {
	Loop() error
	NotifyManualUpdate(update updater.VersionInfo, canInstall bool)
	SetVersion(update updater.VersionInfo)
	NotifySilentUpdateInstalled()
	NotifySilentUpdateError(error)
	WaitUntilFrontendIsReady()
}

// New returns initialized frontend based on `frontendType`, which can be `cli` or `grpc`.
func New(
	frontendType string,
	showWindowOnStart bool,
	panicHandler types.PanicHandler,
	settings *settings.Settings,
	eventListener listener.Listener,
	updater types.Updater,
	bridge *bridge.Bridge,
	restarter types.Restarter,
) Frontend {
	bridgeWrap := types.NewBridgeWrap(bridge)
	switch frontendType {
	case "grpc":
		return grpc.NewService(
			showWindowOnStart,
			panicHandler,
			settings,
			eventListener,
			updater,
			bridgeWrap,
			restarter,
		)

	case "cli":
		return cli.New(
			panicHandler,
			settings,
			eventListener,
			updater,
			bridgeWrap,
			restarter,
		)

	default:
		return nil
	}
}
