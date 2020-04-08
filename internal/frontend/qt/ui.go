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

// +build !nogui

package qt

import (
	"runtime"

	"github.com/therecipe/qt/core"
)

// Interface between go and qml.
//
// Here we implement all the signals / methods.
type GoQMLInterface struct {
	core.QObject

	_ func() `constructor:"init"`

	_ bool   `property:"isAutoStart"`
	_ bool   `property:"isProxyAllowed"`
	_ string `property:"currentAddress"`
	_ string `property:"goos"`
	_ string `property:"credits"`
	_ bool   `property:"isShownOnStart"`
	_ bool   `property:"isFirstStart"`
	_ bool   `property:"isFreshVersion"`
	_ bool   `property:"isRestarting"`
	_ bool   `property:"isConnectionOK"`
	_ bool   `property:"isDefaultPort"`

	_ string `property:"programTitle"`
	_ string `property:"newversion"`
	_ string `property:"fullversion"`
	_ string `property:"downloadLink"`
	_ string `property:"landingPage"`
	_ string `property:"changelog"`
	_ string `property:"bugfixes"`

	// Translations.
	_ string `property:"wrongCredentials"`
	_ string `property:"wrongMailboxPassword"`
	_ string `property:"canNotReachAPI"`
	_ string `property:"credentialsNotRemoved"`
	_ string `property:"versionCheckFailed"`
	_ string `property:"failedAutostartPerm"`
	_ string `property:"failedAutostart"`
	_ string `property:"genericErrSeeLogs"`

	_ float32 `property:"progress"`
	_ int     `property:"progressDescription"`

	_ func(isAvailable bool)   `signal:"setConnectionStatus"`
	_ func(updateState string) `signal:"setUpdateState"`
	_ func()                   `slot:"checkInternet"`

	_ func(systX, systY, systW, systH int) `signal:"toggleMainWin"`

	_ func()                 `signal:"processFinished"`
	_ func()                 `signal:"openManual"`
	_ func(showMessage bool) `signal:"runCheckVersion"`
	_ func()                 `signal:"toggleMainWin"`

	_ func() `signal:"showWindow"`
	_ func() `signal:"showHelp"`
	_ func() `signal:"showQuit"`

	_ func() `slot:"toggleAutoStart"`
	_ func() `slot:"toggleAllowProxy"`
	_ func() `slot:"loadAccounts"`
	_ func() `slot:"openLogs"`
	_ func() `slot:"clearCache"`
	_ func() `slot:"clearKeychain"`
	_ func() `slot:"highlightSystray"`
	_ func() `slot:"errorSystray"`
	_ func() `slot:"normalSystray"`

	_ func()                                                   `slot:"getLocalVersionInfo"`
	_ func(showMessage bool)                                   `slot:"isNewVersionAvailable"`
	_ func() string                                            `slot:"getBackendVersion"`
	_ func() string                                            `slot:"getIMAPPort"`
	_ func() string                                            `slot:"getSMTPPort"`
	_ func() string                                            `slot:"getLastMailClient"`
	_ func(portStr string) int                                 `slot:"isPortOpen"`
	_ func(imapPort, smtpPort string, useSTARTTLSforSMTP bool) `slot:"setPortsAndSecurity"`
	_ func() bool                                              `slot:"isSMTPSTARTTLS"`

	_ func(description, client, address string) bool `slot:"sendBug"`

	_ func(tabIndex int, message string) `signal:"notifyBubble"`
	_ func(tabIndex int, message string) `signal:"silentBubble"`
	_ func()                             `signal:"bubbleClosed"`

	_ func(iAccount int, removePreferences bool) `slot:"deleteAccount"`
	_ func(iAccount int)                         `slot:"logoutAccount"`
	_ func(iAccount int, iAddress int)           `slot:"configureAppleMail"`
	_ func(iAccount int)                         `signal:"switchAddressMode"`

	_ func(login, password string) int      `slot:"login"`
	_ func(twoFacAuth string) int           `slot:"auth2FA"`
	_ func(mailboxPassword string) int      `slot:"addAccount"`
	_ func(message string, changeIndex int) `signal:"setAddAccountWarning"`

	_ func()                                `signal:"notifyVersionIsTheLatest"`
	_ func()                                `signal:"notifyKeychainRebuild"`
	_ func()                                `signal:"notifyHasNoKeychain"`
	_ func()                                `signal:"notifyUpdate"`
	_ func(accname string)                  `signal:"notifyLogout"`
	_ func(accname string)                  `signal:"notifyAddressChanged"`
	_ func(accname string)                  `signal:"notifyAddressChangedLogout"`
	_ func(busyPortIMAP, busyPortSMTP bool) `signal:"notifyPortIssue"`
	_ func(code string)                     `signal:"failedAutostartCode"`

	_ bool                                    `property:"isReportingOutgoingNoEnc"`
	_ func()                                  `slot:"toggleIsReportingOutgoingNoEnc"`
	_ func(messageID string, shouldSend bool) `slot:"shouldSendAnswer"`
	_ func(messageID, subject string)         `signal:"showOutgoingNoEncPopup"`
	_ func(x, y float32)                      `signal:"setOutgoingNoEncPopupCoord"`
	_ func(x, y float32)                      `slot:"saveOutgoingNoEncPopupCoord"`
	_ func(recipient string)                  `signal:"showNoActiveKeyForRecipient"`
	_ func()                                  `signal:"showCertIssue"`

	_ func()              `slot:"startUpdate"`
	_ func(hasError bool) `signal:"updateFinished"`
}

// init is basically the constructor.
func (s *GoQMLInterface) init() {}

// SetFrontend connects all slots and signals from Go to QML.
func (s *GoQMLInterface) SetFrontend(f *FrontendQt) {
	s.ConnectToggleAutoStart(f.toggleAutoStart)
	s.ConnectToggleAllowProxy(f.toggleAllowProxy)
	s.ConnectLoadAccounts(f.loadAccounts)
	s.ConnectOpenLogs(f.openLogs)
	s.ConnectClearCache(f.clearCache)
	s.ConnectClearKeychain(f.clearKeychain)

	s.ConnectGetLocalVersionInfo(f.getLocalVersionInfo)
	s.ConnectIsNewVersionAvailable(f.isNewVersionAvailable)
	s.ConnectGetIMAPPort(f.getIMAPPort)
	s.ConnectGetSMTPPort(f.getSMTPPort)
	s.ConnectGetLastMailClient(f.getLastMailClient)
	s.ConnectIsPortOpen(f.isPortOpen)
	s.ConnectIsSMTPSTARTTLS(f.isSMTPSTARTTLS)

	s.ConnectSendBug(f.sendBug)

	s.ConnectDeleteAccount(f.deleteAccount)
	s.ConnectLogoutAccount(f.logoutAccount)
	s.ConnectConfigureAppleMail(f.configureAppleMail)
	s.ConnectLogin(f.login)
	s.ConnectAuth2FA(f.auth2FA)
	s.ConnectAddAccount(f.addAccount)
	s.ConnectSetPortsAndSecurity(f.setPortsAndSecurity)

	s.ConnectHighlightSystray(HighlightSystray)
	s.ConnectErrorSystray(ErrorSystray)
	s.ConnectNormalSystray(NormalSystray)
	s.ConnectSwitchAddressMode(f.switchAddressModeUser)

	s.SetGoos(runtime.GOOS)
	s.SetIsRestarting(false)
	s.SetProgramTitle(f.programName)

	s.ConnectGetBackendVersion(func() string {
		return f.programVer
	})

	s.ConnectCheckInternet(f.checkInternet)

	s.ConnectToggleIsReportingOutgoingNoEnc(f.toggleIsReportingOutgoingNoEnc)
	s.ConnectShouldSendAnswer(f.shouldSendAnswer)
	s.ConnectSaveOutgoingNoEncPopupCoord(f.saveOutgoingNoEncPopupCoord)
	s.ConnectStartUpdate(f.StartUpdate)
}
