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

// +build nogui

package qt

import (
	"fmt"
	"net/http"

	"github.com/ProtonMail/proton-bridge/internal/frontend/types"
	"github.com/ProtonMail/proton-bridge/pkg/config"
	"github.com/ProtonMail/proton-bridge/pkg/listener"
	"github.com/ProtonMail/proton-bridge/pkg/logs"
)

var log = logs.GetLogEntry("frontend-nogui") //nolint[gochecknoglobals]

type FrontendHeadless struct{}

func (s *FrontendHeadless) Loop(credentialsError error) error {
	log.Info("Check status on localhost:8081")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Bridge is running")
	})
	return http.ListenAndServe(":8081", nil)
}

func (s *FrontendHeadless) InstanceExistAlert()   {}
func (s *FrontendHeadless) IsAppRestarting() bool { return false }

func New(
	version,
	buildVersion string,
	showWindowOnStart bool,
	panicHandler types.PanicHandler,
	config *config.Config,
	preferences *config.Preferences,
	eventListener listener.Listener,
	updates types.Updater,
	bridge types.Bridger,
	noEncConfirmator types.NoEncConfirmator,
) *FrontendHeadless {
	return &FrontendHeadless{}
}
