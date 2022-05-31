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

//go:build build_qt
// +build build_qt

package qt

import (
	"sync"

	"github.com/ProtonMail/proton-bridge/v2/internal/config/settings"
	"github.com/ProtonMail/proton-bridge/v2/internal/updater"
	"github.com/pkg/errors"
)

var checkingUpdates = sync.Mutex{}

func (f *FrontendQt) checkUpdates() error {
	version, err := f.updater.Check()
	if err != nil {
		return err
	}

	f.SetVersion(version)
	return nil
}

func (f *FrontendQt) checkUpdatesAndNotify(isRequestFromUser bool) {
	checkingUpdates.Lock()
	defer checkingUpdates.Unlock()
	defer f.qml.CheckUpdatesFinished()

	if err := f.checkUpdates(); err != nil {
		f.log.WithError(err).Error("An error occurred while checking updates")
		if isRequestFromUser {
			f.qml.UpdateManualError()
		}
		return
	}

	if !f.updater.IsUpdateApplicable(f.newVersionInfo) {
		f.log.Debug("No need to update")
		if isRequestFromUser {
			f.qml.UpdateIsLatestVersion()
		}
		return
	}

	if !f.updater.CanInstall(f.newVersionInfo) {
		f.log.Debug("A manual update is required")
		f.qml.UpdateManualError()
		return
	}

	if f.settings.GetBool(settings.AutoUpdateKey) {
		// NOOP will update eventually
		return
	}

	if isRequestFromUser {
		f.qml.UpdateManualReady(f.newVersionInfo.Version.String())
	}
}

func (f *FrontendQt) updateForce() {
	checkingUpdates.Lock()
	defer checkingUpdates.Unlock()

	version := ""
	if err := f.checkUpdates(); err == nil {
		version = f.newVersionInfo.Version.String()
	}

	f.qml.UpdateForce(version)
}

func (f *FrontendQt) setIsAutomaticUpdateOn() {
	f.qml.SetIsAutomaticUpdateOn(f.settings.GetBool(settings.AutoUpdateKey))
}

func (f *FrontendQt) toggleAutomaticUpdate(makeItEnabled bool) {
	f.qml.SetIsAutomaticUpdateOn(makeItEnabled)
	isEnabled := f.settings.GetBool(settings.AutoUpdateKey)
	if makeItEnabled == isEnabled {
		return
	}

	f.settings.SetBool(settings.AutoUpdateKey, makeItEnabled)

	f.checkUpdatesAndNotify(false)
}

func (f *FrontendQt) setIsBetaEnabled() {
	channel := f.bridge.GetUpdateChannel()
	f.qml.SetIsBetaEnabled(channel == updater.EarlyChannel)
}

func (f *FrontendQt) toggleBeta(makeItEnabled bool) {
	channel := updater.StableChannel
	if makeItEnabled {
		channel = updater.EarlyChannel
	}

	f.bridge.SetUpdateChannel(channel)

	f.setIsBetaEnabled()

	// Immediately check the updates to set the correct landing page link.
	f.checkUpdates()
}

func (f *FrontendQt) installUpdate() {
	checkingUpdates.Lock()
	defer checkingUpdates.Unlock()

	if !f.updater.CanInstall(f.newVersionInfo) {
		f.log.Warning("Skipping update installation, current version too old")
		f.qml.UpdateManualError()
		return
	}

	if err := f.updater.InstallUpdate(f.newVersionInfo); err != nil {
		if errors.Cause(err) == updater.ErrDownloadVerify {
			f.log.WithError(err).Warning("Skipping update installation due to temporary error")
		} else {
			f.log.WithError(err).Error("The update couldn't be installed")
			f.qml.UpdateManualError()
		}
		return
	}

	f.qml.UpdateSilentRestartNeeded()
}
