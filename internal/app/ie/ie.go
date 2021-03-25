// Copyright (c) 2021 Proton Technologies AG
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
	"github.com/ProtonMail/proton-bridge/internal/frontend/types"
	"github.com/ProtonMail/proton-bridge/internal/importexport"
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
	case c.Bool(base.FlagCLI):
		frontendMode = "cli"
	default:
		frontendMode = "qt"
	}

	// We want to remove old versions if the app exits successfully.
	b.AddTeardownAction(b.Versioner.RemoveOldVersions)

	// We want cookies to be saved to disk so they are loaded the next time.
	b.AddTeardownAction(b.CookieJar.PersistCookies)

	f := frontend.NewImportExport(
		constants.Version,
		constants.BuildVersion,
		b.Name,
		frontendMode,
		b.CrashHandler,
		b.Locations,
		b.Settings,
		b.Listener,
		b.Updater,
		ie,
		b,
	)

	// Watch for updates routine
	go func() {
		ticker := time.NewTicker(time.Hour)

		for {
			checkAndHandleUpdate(b.Updater, f, b.Settings.GetBool(settings.AutoUpdateKey))
			<-ticker.C
		}
	}()

	return f.Loop()
}

func checkAndHandleUpdate(u types.Updater, f frontend.Frontend, autoUpdate bool) { //nolint[unparam]
	log := logrus.WithField("pkg", "app/ie")
	version, err := u.Check()
	if err != nil {
		log.WithError(err).Error("An error occurred while checking for updates")
		return
	}

	f.WaitUntilFrontendIsReady()

	// Update links in UI
	f.SetVersion(version)

	if !u.IsUpdateApplicable(version) {
		log.Info("No need to update")
		return
	}

	log.WithField("version", version.Version).Info("An update is available")

	f.NotifyManualUpdate(version, u.CanInstall(version))
}
