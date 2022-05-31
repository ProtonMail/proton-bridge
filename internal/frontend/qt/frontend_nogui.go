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

//go:build !build_qt
// +build !build_qt

package qt

import (
	"fmt"
	"net/http"

	"github.com/ProtonMail/proton-bridge/v2/internal/config/settings"
	"github.com/ProtonMail/proton-bridge/v2/internal/config/useragent"
	"github.com/ProtonMail/proton-bridge/v2/internal/frontend/types"
	"github.com/ProtonMail/proton-bridge/v2/internal/locations"
	"github.com/ProtonMail/proton-bridge/v2/internal/updater"
	"github.com/ProtonMail/proton-bridge/v2/pkg/listener"
	"github.com/sirupsen/logrus"
)

var log = logrus.WithField("pkg", "frontend-nogui") //nolint:gochecknoglobals

type FrontendHeadless struct{}

func New(
	version,
	buildVersion,
	programName string,
	showWindowOnStart bool,
	panicHandler types.PanicHandler,
	locations *locations.Locations,
	settings *settings.Settings,
	eventListener listener.Listener,
	updater types.Updater,
	userAgent *useragent.UserAgent,
	bridge types.Bridger,
	noEncConfirmator types.NoEncConfirmator,
	restarter types.Restarter,
) *FrontendHeadless {
	return &FrontendHeadless{}
}

func (s *FrontendHeadless) Loop() error {
	log.Info("Check status on localhost:8081")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Bridge is running")
	})
	return http.ListenAndServe(":8081", nil)
}

func (s *FrontendHeadless) NotifyManualUpdate(update updater.VersionInfo, canInstall bool) {
	// NOTE: Save the update somewhere so that it can be installed when user chooses "install now".
}

func (s *FrontendHeadless) WaitUntilFrontendIsReady() {
}

func (s *FrontendHeadless) SetVersion(update updater.VersionInfo) {
}

func (s *FrontendHeadless) NotifySilentUpdateInstalled() {
}

func (s *FrontendHeadless) NotifySilentUpdateError(err error) {
}

func (s *FrontendHeadless) InstanceExistAlert() {}
