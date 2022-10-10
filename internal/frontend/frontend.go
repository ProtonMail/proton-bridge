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
	"github.com/ProtonMail/proton-bridge/v2/internal/frontend/cli"
	"github.com/ProtonMail/proton-bridge/v2/internal/frontend/grpc"
	"github.com/ProtonMail/proton-bridge/v2/internal/frontend/types"
	"github.com/ProtonMail/proton-bridge/v2/internal/locations"
	"github.com/ProtonMail/proton-bridge/v2/internal/updater"
	"github.com/ProtonMail/proton-bridge/v2/pkg/listener"
)

// Type describes the available types of frontend.
type Type int

const (
	CLI Type = iota
	GRPC
	NonInteractive
)

type Frontend interface {
	Loop(b types.Bridger) error
	NotifyManualUpdate(update updater.VersionInfo, canInstall bool)
	SetVersion(update updater.VersionInfo)
	NotifySilentUpdateInstalled()
	NotifySilentUpdateError(error)
	WaitUntilFrontendIsReady()
}

// New returns initialized frontend based on `frontendType`, which can be `CLI` or `GRPC`.
func New(
	frontendType Type,
	showWindowOnStart bool,
	panicHandler types.PanicHandler,
	eventListener listener.Listener,
	updater types.Updater,
	restarter types.Restarter,
	locations *locations.Locations,
) Frontend {
	switch frontendType {
	case GRPC:
		return grpc.NewService(
			showWindowOnStart,
			panicHandler,
			eventListener,
			updater,
			restarter,
			locations,
		)

	case CLI:
		return cli.New(
			panicHandler,
			eventListener,
			updater,
			restarter,
		)

	case NonInteractive:
		fallthrough

	default:
		return nil
	}
}
