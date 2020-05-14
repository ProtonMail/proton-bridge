// Copyright (c) 2020 Proton Technologies AG
//
// This file is part of ProtonMail Bridge.
//
// ProtonMail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ProtonMail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ProtonMail Bridge.  If not, see <https://www.gnu.org/licenses/>.

// Package frontend provides all interfaces of the Bridge.
package frontend

import (
	"github.com/0xAX/notificator"
	"github.com/ProtonMail/proton-bridge/internal/bridge"
	"github.com/ProtonMail/proton-bridge/internal/frontend/cli"
	cliie "github.com/ProtonMail/proton-bridge/internal/frontend/cli-ie"
	"github.com/ProtonMail/proton-bridge/internal/frontend/qt"
	"github.com/ProtonMail/proton-bridge/internal/frontend/types"
	"github.com/ProtonMail/proton-bridge/internal/importexport"
	"github.com/ProtonMail/proton-bridge/pkg/config"
	"github.com/ProtonMail/proton-bridge/pkg/listener"
	"github.com/sirupsen/logrus"
)

var (
	log = logrus.WithField("pkg", "frontend") // nolint[unused]
)

// Frontend is an interface to be implemented by each frontend type (cli, gui, html).
type Frontend interface {
	Loop(credentialsError error) error
	IsAppRestarting() bool
}

// HandlePanic handles panics which occur for users with GUI.
func HandlePanic() {
	notify := notificator.New(notificator.Options{
		DefaultIcon: "../frontend/ui/icon/icon.png",
		AppName:     "ProtonMail Bridge",
	})
	_ = notify.Push("Fatal Error", "The ProtonMail Bridge has encountered a fatal error. ", "/frontend/icon/icon.png", notificator.UR_CRITICAL)
}

// New returns initialized frontend based on `frontendType`, which can be `cli` or `qt`.
func New(
	version,
	buildVersion,
	frontendType string,
	showWindowOnStart bool,
	panicHandler types.PanicHandler,
	config *config.Config,
	preferences *config.Preferences,
	eventListener listener.Listener,
	updates types.Updater,
	bridge *bridge.Bridge,
	noEncConfirmator types.NoEncConfirmator,
) Frontend {
	bridgeWrap := types.NewBridgeWrap(bridge)
	return new(version, buildVersion, frontendType, showWindowOnStart, panicHandler, config, preferences, eventListener, updates, bridgeWrap, noEncConfirmator)
}

func new(
	version,
	buildVersion,
	frontendType string,
	showWindowOnStart bool,
	panicHandler types.PanicHandler,
	config *config.Config,
	preferences *config.Preferences,
	eventListener listener.Listener,
	updates types.Updater,
	bridge types.Bridger,
	noEncConfirmator types.NoEncConfirmator,
) Frontend {
	switch frontendType {
	case "cli":
		return cli.New(panicHandler, config, preferences, eventListener, updates, bridge)
	default:
		return qt.New(version, buildVersion, showWindowOnStart, panicHandler, config, preferences, eventListener, updates, bridge, noEncConfirmator)
	}
}

// NewImportExport returns initialized frontend based on `frontendType`, which can be `cli` or `qt`.
func NewImportExport(
	version,
	buildVersion,
	frontendType string,
	panicHandler types.PanicHandler,
	config *config.Config,
	eventListener listener.Listener,
	updates types.Updater,
	ie *importexport.ImportExport,
) Frontend {
	ieWrap := types.NewImportExportWrap(ie)
	return newImportExport(version, buildVersion, frontendType, panicHandler, config, eventListener, updates, ieWrap)
}

func newImportExport(
	version,
	buildVersion,
	frontendType string,
	panicHandler types.PanicHandler,
	config *config.Config,
	eventListener listener.Listener,
	updates types.Updater,
	ie types.ImportExporter,
) Frontend {
	switch frontendType {
	case "cli":
		return cliie.New(panicHandler, config, eventListener, updates, ie)
	default:
		return cliie.New(panicHandler, config, eventListener, updates, ie)
		//return qt.New(panicHandler, config, eventListener, updates, ie)
	}
}
