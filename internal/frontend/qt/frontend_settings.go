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
	"runtime"
	"time"

	"github.com/ProtonMail/proton-bridge/v2/internal/config/settings"
	"github.com/ProtonMail/proton-bridge/v2/internal/frontend/clientconfig"
	"github.com/ProtonMail/proton-bridge/v2/internal/frontend/theme"
	"github.com/ProtonMail/proton-bridge/v2/pkg/keychain"
	"github.com/ProtonMail/proton-bridge/v2/pkg/ports"
	"github.com/therecipe/qt/core"
)

func (f *FrontendQt) setIsDiskCacheEnabled() {
	f.qml.SetIsDiskCacheEnabled(f.settings.GetBool(settings.CacheEnabledKey))
}

func (f *FrontendQt) setDiskCachePath() {
	f.qml.SetDiskCachePath(core.QUrl_FromLocalFile(f.settings.Get(settings.CacheLocationKey)))
}

func (f *FrontendQt) changeLocalCache(enableDiskCache bool, diskCachePath *core.QUrl) {
	defer f.qml.ChangeLocalCacheFinished()
	defer f.setIsDiskCacheEnabled()
	defer f.setDiskCachePath()

	if enableDiskCache != f.settings.GetBool(settings.CacheEnabledKey) {
		if enableDiskCache {
			if err := f.bridge.EnableCache(); err != nil {
				f.log.WithError(err).Error("Cannot enable disk cache")
			}
		} else {
			if err := f.bridge.DisableCache(); err != nil {
				f.log.WithError(err).Error("Cannot disable disk cache")
			}
		}
	}

	_diskCachePath := diskCachePath.Path(core.QUrl__FullyDecoded)
	if (runtime.GOOS == "windows") && (_diskCachePath[0] == '/') {
		_diskCachePath = _diskCachePath[1:]
	}

	if enableDiskCache && _diskCachePath != f.settings.Get(settings.CacheLocationKey) {
		if err := f.bridge.MigrateCache(f.settings.Get(settings.CacheLocationKey), _diskCachePath); err != nil {
			f.log.WithError(err).Error("The local cache location could not be changed.")
			f.qml.CacheCantMove()
			return
		}
		f.settings.Set(settings.CacheLocationKey, _diskCachePath)
	}

	f.qml.CacheLocationChangeSuccess()
	f.restart()
}

func (f *FrontendQt) setIsAutostartOn() {
	// GODT-1507 Windows: autostart needs to be created after Qt is initialized.
	// GODT-1206: if preferences file says it should be on enable it here.
	f.firstTimeAutostart.Do(func() {
		shouldAutostartBeOn := f.settings.GetBool(settings.AutostartKey)
		if f.bridge.IsFirstStart() || shouldAutostartBeOn {
			if err := f.bridge.EnableAutostart(); err != nil {
				f.log.WithField("prefs", shouldAutostartBeOn).WithError(err).Error("Failed to enable first autostart")
			}
			return
		}
	})
	f.qml.SetIsAutostartOn(f.bridge.IsAutostartEnabled())
}

func (f *FrontendQt) toggleAutostart(makeItEnabled bool) {
	defer f.qml.ToggleAutostartFinished()
	if makeItEnabled == f.bridge.IsAutostartEnabled() {
		f.setIsAutostartOn()
		return
	}

	var err error
	if makeItEnabled {
		err = f.bridge.EnableAutostart()
	} else {
		err = f.bridge.DisableAutostart()
	}
	f.setIsAutostartOn()

	if err != nil {
		f.log.
			WithField("makeItEnabled", makeItEnabled).
			WithField("isEnabled", f.qml.IsAutostartOn()).
			WithError(err).
			Error("Autostart change failed")
	}
}

func (f *FrontendQt) toggleDoH(makeItEnabled bool) {
	f.bridge.SetProxyAllowed(makeItEnabled)
	f.qml.SetIsDoHEnabled(f.bridge.GetProxyAllowed())
}

func (f *FrontendQt) toggleUseSSLforSMTP(makeItEnabled bool) {
	if f.settings.GetBool(settings.SMTPSSLKey) == makeItEnabled {
		f.qml.SetUseSSLforSMTP(makeItEnabled)
		return
	}
	f.settings.SetBool(settings.SMTPSSLKey, makeItEnabled)
	f.restart()
}

func (f *FrontendQt) changePorts(imapPort, smtpPort int) {
	defer f.qml.ChangePortFinished()
	f.settings.SetInt(settings.IMAPPortKey, imapPort)
	f.settings.SetInt(settings.SMTPPortKey, smtpPort)
	f.restart()
}

func (f *FrontendQt) isPortFree(port int) bool {
	return ports.IsPortFree(port)
}

func (f *FrontendQt) configureAppleMail(userID, address string) {
	user, err := f.bridge.GetUser(userID)
	if err != nil {
		f.log.WithField("userID", userID).Error("Cannot configure AppleMail for user")
		return
	}

	needRestart, err := clientconfig.ConfigureAppleMail(user, address, f.settings)
	if err != nil {
		f.log.WithError(err).Error("Apple Mail config failed")
	}

	if needRestart {
		// There is delay needed for external window to open
		time.Sleep(2 * time.Second)
		f.restart()
	}
}

func (f *FrontendQt) triggerReset() {
	defer f.qml.ResetFinished()
	f.bridge.FactoryReset()
	f.restart()
}

func (f *FrontendQt) setKeychain() {
	availableKeychain := []string{}
	for chain := range keychain.Helpers {
		availableKeychain = append(availableKeychain, chain)
	}
	f.qml.SetAvailableKeychain(availableKeychain)
	f.qml.SetCurrentKeychain(f.bridge.GetKeychainApp())
}

func (f *FrontendQt) changeKeychain(wantKeychain string) {
	defer f.qml.ChangeKeychainFinished()

	if f.bridge.GetKeychainApp() == wantKeychain {
		return
	}

	f.bridge.SetKeychainApp(wantKeychain)
	f.restart()
}

func (f *FrontendQt) restart() {
	f.log.Info("Restarting bridge")
	f.restarter.SetToRestart()
	f.app.Exit(0)
}

func (f *FrontendQt) quit() {
	f.log.Warn("Your wish is my command.. I quit!")
	f.app.Exit(0)
}

func (f *FrontendQt) guiReady() {
	f.initializationDone.Do(f.initializing.Done)
}

func (f *FrontendQt) setColorScheme() {
	current := f.settings.Get(settings.ColorScheme)
	if !theme.IsAvailable(theme.Theme(current)) {
		current = string(theme.DefaultTheme())
		f.settings.Set(settings.ColorScheme, current)
	}
	f.qml.SetColorSchemeName(current)
}

func (f *FrontendQt) changeColorScheme(newScheme string) {
	if !theme.IsAvailable(theme.Theme(newScheme)) {
		f.log.WithField("scheme", newScheme).Warn("Color scheme not available")
		return
	}
	f.settings.Set(settings.ColorScheme, newScheme)
	f.setColorScheme()
}
