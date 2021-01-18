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

// +build nogui

package qtie

import (
	"fmt"
	"net/http"

	"github.com/ProtonMail/proton-bridge/internal/config/settings"
	"github.com/ProtonMail/proton-bridge/internal/frontend/types"
	"github.com/ProtonMail/proton-bridge/internal/locations"
	"github.com/ProtonMail/proton-bridge/internal/updater"
	"github.com/ProtonMail/proton-bridge/pkg/listener"
	"github.com/sirupsen/logrus"
)

var log = logrus.WithField("pkg", "frontend-nogui") //nolint[gochecknoglobals]

type FrontendHeadless struct{}

func (s *FrontendHeadless) Loop() error {
	log.Info("Check status on localhost:8082")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "IE is running")
	})
	return http.ListenAndServe(":8082", nil)
}

func (s *FrontendHeadless) NotifyManualUpdate(update updater.VersionInfo, canInstall bool) {
	// NOTE: Save the update somewhere so that it can be installed when user chooses "install now".
}

func (s *FrontendHeadless) SetVersion(update updater.VersionInfo) {
}

func (s *FrontendHeadless) NotifySilentUpdateInstalled() {
}

func (s *FrontendHeadless) NotifySilentUpdateError(err error) {
}

func New(
	version, buildVersion, appName string,
	panicHandler types.PanicHandler,
	locations *locations.Locations,
	settings *settings.Settings,
	eventListener listener.Listener,
	updater types.Updater,
	ie types.ImportExporter,
	restarter types.Restarter,
) *FrontendHeadless {
	return &FrontendHeadless{}
}
