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

//go:build build_qt
// +build build_qt

package qt

import (
	"runtime"

	"github.com/ProtonMail/proton-bridge/internal/bridge"
	"github.com/ProtonMail/proton-bridge/internal/config/settings"
	dockIcon "github.com/ProtonMail/proton-bridge/internal/frontend/qt/dockicon"
	"github.com/therecipe/qt/core"
)

func init() {
	QMLBackend_QRegisterMetaType()
}

// QMLBackend connects QML frontend with Go backend.
type QMLBackend struct {
	core.QObject

	_ func() *core.QPoint `slot:"getCursorPos"`
	_ func()              `slot:"guiReady"`
	_ func()              `slot:"quit"`
	_ func()              `slot:"restart"`

	_ bool `property:showOnStartup`

	_ bool `property:dockIconVisible`

	_ QMLUserModel `property:"users"`

	// TODO copy stuff from Bridge_test.qml backend object
	_ string `property:"goos"`

	_ func(username, password string) `slot:"login"`
	_ func(username, code string)     `slot:"login2FA"`
	_ func(username, password string) `slot:"login2Password"`
	_ func(username string)           `slot:"loginAbort"`
	_ func(errorMsg string)           `signal:"loginUsernamePasswordError"`
	_ func()                          `signal:"loginFreeUserError"`
	_ func(errorMsg string)           `signal:"loginConnectionError"`
	_ func(username string)           `signal:"login2FARequested"`
	_ func(errorMsg string)           `signal:"login2FAError"`
	_ func(errorMsg string)           `signal:"login2FAErrorAbort"`
	_ func()                          `signal:"login2PasswordRequested"`
	_ func(errorMsg string)           `signal:"login2PasswordError"`
	_ func(errorMsg string)           `signal:"login2PasswordErrorAbort"`
	_ func()                          `signal:"loginFinished"`

	_ func() `signal:"internetOff"`
	_ func() `signal:"internetOn"`

	_ func(version string) `signal:"updateManualReady"`
	_ func()               `signal:"updateManualRestartNeeded"`
	_ func()               `signal:"updateManualError"`
	_ func(version string) `signal:"updateForce"`
	_ func()               `signal:"updateForceError"`
	_ func()               `signal:"updateSilentRestartNeeded"`
	_ func()               `signal:"updateSilentError"`
	_ func()               `signal:"updateIsLatestVersion"`
	_ func()               `slot:"checkUpdates"`
	_ func()               `signal:"checkUpdatesFinished"`

	_ bool                                             `property:"isDiskCacheEnabled"`
	_ string                                           `property:"diskCachePath"`
	_ func()                                           `signal:"cacheUnavailable"`
	_ func()                                           `signal:"cacheCantMove"`
	_ func()                                           `signal:"cacheLocationChangeSuccess"`
	_ func()                                           `signal:"diskFull"`
	_ func(enableDiskCache bool, diskCachePath string) `slot:"changeLocalCache"`
	_ func()                                           `signal:"changeLocalCacheFinished"`

	_ bool                    `property:"isAutomaticUpdateOn"`
	_ func(makeItActive bool) `slot:"toggleAutomaticUpdate"`

	_ bool                    `property:"isAutostartOn"`
	_ func(makeItActive bool) `slot:"toggleAutostart"`
	_ func()                  `signal:"toggleAutostartFinished"`

	_ bool                    `property:"isBetaEnabled"`
	_ func(makeItActive bool) `slot:"toggleBeta"`

	_ bool                    `property:"isDoHEnabled"`
	_ func(makeItActive bool) `slot:"toggleDoH"`

	_ bool                    `property:"useSSLforSMTP"`
	_ func(makeItActive bool) `slot:"toggleUseSSLforSMTP"`
	_ func()                  `signal:"toggleUseSSLFinished"`

	_ string                       `property:"hostname"`
	_ int                          `property:"portIMAP"`
	_ int                          `property:"portSMTP"`
	_ func(imapPort, smtpPort int) `slot:"changePorts"`
	_ func(port int) bool          `slot:"isPortFree"`
	_ func()                       `signal:"changePortFinished"`
	_ func()                       `signal:"portIssueIMAP"`
	_ func()                       `signal:"portIssueSMTP"`

	_ func() `slot:"triggerReset"`
	_ func() `signal:"resetFinished"`

	_ string `property:"version"`
	_ string `property:"logsPath"`
	_ string `property:"licensePath"`
	_ string `property:"releaseNotesLink"`
	_ string `property:"landingPageLink"`

	_ string                                                           `property:"currentEmailClient"`
	_ func()                                                           `slot:"updateCurrentMailClient"`
	_ func(description, address, emailClient string, includeLogs bool) `slot:"reportBug"`
	_ func()                                                           `signal:"reportBugFinished"`
	_ func()                                                           `signal:"bugReportSendSuccess"`
	_ func()                                                           `signal:"bugReportSendError"`

	_ []string              `property:"availableKeychain"`
	_ string                `property:"selectedKeychain"`
	_ func(keychain string) `slot:"selectKeychain"`
	_ func()                `signal:"notifyHasNoKeychain"`

	_ func(email string) `signal:noActiveKeyForRecipient`
	_ func()             `signal:showMainWindow`

	_ func(address string)  `signal:addressChanged`
	_ func(address string)  `signal:addressChangedLogout`
	_ func(username string) `signal:userDisconnected`
	_ func()                `signal:apiCertIssue`

	_ func(userID string) `signal:userChanged`
}

func (q *QMLBackend) setup(f *FrontendQt) {
	q.ConnectGetCursorPos(getCursorPos)
	q.ConnectQuit(f.quit)
	q.ConnectRestart(f.restart)
	q.ConnectGuiReady(f.guiReady)

	q.ConnectIsShowOnStartup(func() bool {
		return f.showOnStartup
	})

	q.ConnectIsDockIconVisible(dockIcon.GetDockIconVisibleState)
	q.ConnectSetDockIconVisible(dockIcon.SetDockIconVisibleState)

	um := NewQMLUserModel(q)
	um.f = f
	q.SetUsers(um)
	um.load()

	q.ConnectUserChanged(um.userChanged)

	q.SetGoos(runtime.GOOS)

	q.ConnectLogin(func(u, p string) {
		go func() {
			defer f.panicHandler.HandlePanic()
			f.login(u, p)
		}()
	})
	q.ConnectLogin2FA(func(u, p string) {
		go func() {
			defer f.panicHandler.HandlePanic()
			f.login2FA(u, p)
		}()
	})
	q.ConnectLogin2Password(func(u, p string) {
		go func() {
			defer f.panicHandler.HandlePanic()
			f.login2Password(u, p)
		}()
	})
	q.ConnectLoginAbort(func(u string) {
		go func() {
			defer f.panicHandler.HandlePanic()
			f.loginAbort(u)
		}()
	})

	go func() {
		defer f.panicHandler.HandlePanic()
		f.checkUpdatesAndNotify(false)
	}()
	q.ConnectCheckUpdates(func() {
		go func() {
			defer f.panicHandler.HandlePanic()
			f.checkUpdatesAndNotify(true)
		}()
	})

	f.setIsDiskCacheEnabled()
	f.setDiskCachePath()
	q.ConnectChangeLocalCache(func(e bool, d string) {
		go func() {
			defer f.panicHandler.HandlePanic()
			f.changeLocalCache(e, d)
		}()
	})

	f.setIsAutomaticUpdateOn()
	q.ConnectToggleAutomaticUpdate(func(m bool) {
		go func() {
			defer f.panicHandler.HandlePanic()
			f.toggleAutomaticUpdate(m)
		}()
	})

	f.setIsAutostartOn()
	q.ConnectToggleAutostart(f.toggleAutostart)

	f.setIsBetaEnabled()
	q.ConnectToggleBeta(func(m bool) {
		go func() {
			defer f.panicHandler.HandlePanic()
			f.toggleBeta(m)
		}()
	})

	q.SetIsDoHEnabled(f.bridge.GetProxyAllowed())
	q.ConnectToggleDoH(f.toggleDoH)

	q.SetUseSSLforSMTP(f.settings.GetBool(settings.SMTPSSLKey))
	q.ConnectToggleUseSSLforSMTP(f.toggleUseSSLforSMTP)

	q.SetHostname(bridge.Host)
	q.SetPortIMAP(f.settings.GetInt(settings.IMAPPortKey))
	q.SetPortSMTP(f.settings.GetInt(settings.SMTPPortKey))
	q.ConnectChangePorts(f.changePorts)
	q.ConnectIsPortFree(f.isPortFree)

	q.ConnectTriggerReset(func() {
		go func() {
			defer f.panicHandler.HandlePanic()
			f.triggerReset()
		}()
	})

	f.setVersion()
	f.setLogsPath()
	// release notes link is set by update
	f.setLicensePath()

	f.setCurrentEmailClient()
	q.ConnectUpdateCurrentMailClient(func() {
		go func() {
			defer f.panicHandler.HandlePanic()
			f.setCurrentEmailClient()
		}()
	})
	q.ConnectReportBug(func(d, a, e string, i bool) {
		go func() {
			defer f.panicHandler.HandlePanic()
			f.reportBug(d, a, e, i)
		}()
	})

	f.setKeychain()
	q.ConnectSelectKeychain(func(k string) {
		go func() {
			defer f.panicHandler.HandlePanic()
			f.selectKeychain(k)
		}()
	})
}
