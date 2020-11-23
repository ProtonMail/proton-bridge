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

// Package ie implements the ie CLI application.
package ie

import (
	"time"

	"github.com/ProtonMail/proton-bridge/internal/api"
	"github.com/ProtonMail/proton-bridge/internal/app/base"
	"github.com/ProtonMail/proton-bridge/internal/config/settings"
	"github.com/ProtonMail/proton-bridge/internal/constants"
	"github.com/ProtonMail/proton-bridge/internal/frontend"
	"github.com/ProtonMail/proton-bridge/internal/importexport"
	"github.com/ProtonMail/proton-bridge/internal/updater"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func New(b *base.Base) *cli.App {
	return b.NewApp(run)
}

func run(b *base.Base, c *cli.Context) error {
	ie := importexport.New(b.Locations, b.Cache, b.CrashHandler, b.Listener, b.CM, b.Creds)

	go func() {
		defer b.CrashHandler.HandlePanic()
		api.NewAPIServer(b.Settings, b.Listener).ListenAndServe()
	}()

	var frontendMode string

	switch {
	case c.Bool("cli"):
		frontendMode = "cli"
	default:
		frontendMode = "qt"
	}

	f := frontend.NewImportExport(
		constants.Version,
		constants.BuildVersion,
		frontendMode,
		b.CrashHandler,
		b.Locations,
		b.Listener,
		b.Updater,
		ie,
		b,
	)

	b.Updater.Watch(
		time.Hour,
		func(update updater.VersionInfo) error {
			if !b.Settings.GetBool(settings.AutoUpdateKey) {
				return f.NotifyManualUpdate(update)
			}

			return b.Updater.InstallUpdate(update)
		},
		func(err error) {
			logrus.WithError(err).Error("An error occurred while watching for updates")
		},
	)

	return f.Loop()
}
