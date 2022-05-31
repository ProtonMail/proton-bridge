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

// Package qt provides communication between Qt/QML frontend and Go backend
package qt

import (
	"strings"

	"github.com/ProtonMail/proton-bridge/v2/internal/bridge"
	"github.com/ProtonMail/proton-bridge/v2/internal/events"
	"github.com/ProtonMail/proton-bridge/v2/pkg/keychain"
)

func (f *FrontendQt) watchEvents() {
	f.WaitUntilFrontendIsReady()

	// First we check bridge global errors for any error that should be shown on GUI.
	if f.bridge.HasError(bridge.ErrLocalCacheUnavailable) {
		f.qml.CacheUnavailable()
	}

	errorCh := f.eventListener.ProvideChannel(events.ErrorEvent)
	credentialsErrorCh := f.eventListener.ProvideChannel(events.CredentialsErrorEvent)
	noActiveKeyForRecipientCh := f.eventListener.ProvideChannel(events.NoActiveKeyForRecipientEvent)
	internetConnChangedCh := f.eventListener.ProvideChannel(events.InternetConnChangedEvent)
	secondInstanceCh := f.eventListener.ProvideChannel(events.SecondInstanceEvent)
	restartBridgeCh := f.eventListener.ProvideChannel(events.RestartBridgeEvent)
	addressChangedCh := f.eventListener.ProvideChannel(events.AddressChangedEvent)
	addressChangedLogoutCh := f.eventListener.ProvideChannel(events.AddressChangedLogoutEvent)
	logoutCh := f.eventListener.ProvideChannel(events.LogoutEvent)
	updateApplicationCh := f.eventListener.ProvideChannel(events.UpgradeApplicationEvent)
	userChangedCh := f.eventListener.ProvideChannel(events.UserRefreshEvent)
	certIssue := f.eventListener.ProvideChannel(events.TLSCertIssue)

	// This loop is executed outside main Qt application thread. In order
	// to make sure that all signals are propagated correctly to QML we
	// must call QMLBackend signals to apply any changes to GUI. The
	// signals will make sure the changes are executed in main Qt app
	// thread.
	for {
		select {
		case errorDetails := <-errorCh:
			if strings.Contains(errorDetails, "IMAP failed") {
				f.qml.PortIssueIMAP()
			}
			if strings.Contains(errorDetails, "SMTP failed") {
				f.qml.PortIssueSMTP()
			}
		case reason := <-credentialsErrorCh:
			if reason == keychain.ErrMacKeychainRebuild.Error() {
				f.qml.NotifyRebuildKeychain()
				continue
			}
			f.qml.NotifyHasNoKeychain()
		case email := <-noActiveKeyForRecipientCh:
			f.qml.NoActiveKeyForRecipient(email)
		case stat := <-internetConnChangedCh:
			if stat == events.InternetOff {
				f.qml.InternetOff()
			}
			if stat == events.InternetOn {
				f.qml.InternetOn()
			}
		case <-secondInstanceCh:
			f.qml.ShowMainWindow()
		case <-restartBridgeCh:
			f.restart()
		case address := <-addressChangedCh:
			f.qml.AddressChanged(address)
		case address := <-addressChangedLogoutCh:
			f.qml.AddressChangedLogout(address)
		case userID := <-logoutCh:
			user, err := f.bridge.GetUser(userID)
			if err != nil {
				return
			}
			f.qml.UserDisconnected(user.Username())
		case <-updateApplicationCh:
			f.updateForce()
		case userID := <-userChangedCh:
			f.qml.UserChanged(userID)
		case <-certIssue:
			f.qml.ApiCertIssue()
		}
	}
}
