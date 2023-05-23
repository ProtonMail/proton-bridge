// Copyright (c) 2023 Proton AG
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
	"github.com/ProtonMail/proton-bridge/v3/internal/events"
	"github.com/ProtonMail/proton-bridge/v3/internal/updater"
	"github.com/abiosoft/ishell"
)

func (f *frontendCLI) checkUpdates(_ *ishell.Context) {
	updateCh, done := f.bridge.GetEvents(events.UpdateAvailable{}, events.UpdateNotAvailable{})
	defer done()

	f.bridge.CheckForUpdates()

	switch (<-updateCh).(type) {
	case events.UpdateAvailable:
		// ... this is handled by the main event loop

	case events.UpdateNotAvailable:
		f.Println("Bridge is already up to date.")
	}
}

func (f *frontendCLI) enableAutoUpdates(_ *ishell.Context) {
	if f.bridge.GetAutoUpdate() {
		f.Println("Bridge is already set to automatically install updates.")
		return
	}

	f.Println("Bridge is currently set to NOT automatically install updates.")

	if f.yesNoQuestion("Are you sure you want to allow bridge to do this") {
		if err := f.bridge.SetAutoUpdate(true); err != nil {
			f.printAndLogError(err)
			return
		}
	}
}

func (f *frontendCLI) disableAutoUpdates(_ *ishell.Context) {
	if !f.bridge.GetAutoUpdate() {
		f.Println("Bridge is already set to NOT automatically install updates.")
		return
	}

	f.Println("Bridge is currently set to automatically install updates.")

	if f.yesNoQuestion("Are you sure you want to stop bridge from doing this") {
		if err := f.bridge.SetAutoUpdate(false); err != nil {
			f.printAndLogError(err)
			return
		}
	}
}

func (f *frontendCLI) selectEarlyChannel(_ *ishell.Context) {
	if f.bridge.GetUpdateChannel() == updater.EarlyChannel {
		f.Println("Bridge is already on the early-access update channel.")
		return
	}

	f.Println("Bridge is currently on the stable update channel.")

	if f.yesNoQuestion("Are you sure you want to switch to the early-access update channel") {
		if err := f.bridge.SetUpdateChannel(updater.EarlyChannel); err != nil {
			f.printAndLogError(err)
			return
		}
	}
}

func (f *frontendCLI) selectStableChannel(_ *ishell.Context) {
	if f.bridge.GetUpdateChannel() == updater.StableChannel {
		f.Println("Bridge is already on the stable update channel.")
		return
	}

	f.Println("Bridge is currently on the early-access update channel.")
	f.Println("Switching to the stable channel may reset all data!")

	if f.yesNoQuestion("Are you sure you want to switch to the stable update channel") {
		if err := f.bridge.SetUpdateChannel(updater.StableChannel); err != nil {
			f.printAndLogError(err)
			return
		}
	}
}
