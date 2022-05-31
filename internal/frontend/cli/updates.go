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

package cli

import (
	"strings"

	"github.com/ProtonMail/proton-bridge/v2/internal/bridge"
	"github.com/ProtonMail/proton-bridge/v2/internal/config/settings"
	"github.com/ProtonMail/proton-bridge/v2/internal/updater"
	"github.com/abiosoft/ishell"
)

func (f *frontendCLI) checkUpdates(c *ishell.Context) {
	version, err := f.updater.Check()
	if err != nil {
		f.Println("An error occurred while checking for updates.")
		return
	}

	if f.updater.IsUpdateApplicable(version) {
		f.Println("An update is available.")
	} else {
		f.Println("Your version is up to date.")
	}
}

func (f *frontendCLI) printCredits(c *ishell.Context) {
	for _, pkg := range strings.Split(bridge.Credits, ";") {
		f.Println(pkg)
	}
}

func (f *frontendCLI) enableAutoUpdates(c *ishell.Context) {
	if f.settings.GetBool(settings.AutoUpdateKey) {
		f.Println("Bridge is already set to automatically install updates.")
		return
	}

	f.Println("Bridge is currently set to NOT automatically install updates.")

	if f.yesNoQuestion("Are you sure you want to allow bridge to do this") {
		f.settings.SetBool(settings.AutoUpdateKey, true)
	}
}

func (f *frontendCLI) disableAutoUpdates(c *ishell.Context) {
	if !f.settings.GetBool(settings.AutoUpdateKey) {
		f.Println("Bridge is already set to NOT automatically install updates.")
		return
	}

	f.Println("Bridge is currently set to automatically install updates.")

	if f.yesNoQuestion("Are you sure you want to stop bridge from doing this") {
		f.settings.SetBool(settings.AutoUpdateKey, false)
	}
}

func (f *frontendCLI) selectEarlyChannel(c *ishell.Context) {
	if f.bridge.GetUpdateChannel() == updater.EarlyChannel {
		f.Println("Bridge is already on the early-access update channel.")
		return
	}

	f.Println("Bridge is currently on the stable update channel.")

	if f.yesNoQuestion("Are you sure you want to switch to the early-access update channel") {
		f.bridge.SetUpdateChannel(updater.EarlyChannel)
	}
}

func (f *frontendCLI) selectStableChannel(c *ishell.Context) {
	if f.bridge.GetUpdateChannel() == updater.StableChannel {
		f.Println("Bridge is already on the stable update channel.")
		return
	}

	f.Println("Bridge is currently on the early-access update channel.")
	f.Println("Switching to the stable channel may reset all data!")

	if f.yesNoQuestion("Are you sure you want to switch to the stable update channel") {
		f.bridge.SetUpdateChannel(updater.StableChannel)
	}
}
