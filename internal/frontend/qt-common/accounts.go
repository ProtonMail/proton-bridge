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

// +build build_qt

package qtcommon

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/ProtonMail/proton-bridge/internal/bridge"
	"github.com/ProtonMail/proton-bridge/internal/config/settings"
	"github.com/ProtonMail/proton-bridge/internal/events"
	"github.com/ProtonMail/proton-bridge/internal/frontend/types"
	"github.com/ProtonMail/proton-bridge/pkg/keychain"
	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
)

// QMLer sends signals to GUI
type QMLer interface {
	ProcessFinished()
	NotifyHasNoKeychain()
	SetConnectionStatus(bool)
	SetAddAccountWarning(string, int)
	NotifyBubble(int, string)
	EmitEvent(string, string)
	Quit()

	CanNotReachAPI() string
	WrongMailboxPassword() string
}

// Accounts holds functionality of users
type Accounts struct {
	Model    *AccountsModel
	qml      QMLer
	um       types.UserManager
	settings *settings.Settings

	authClient pmapi.Client
	auth       *pmapi.Auth

	LatestUserID string
	accountMutex sync.Mutex
	restarter    types.Restarter
}

// SetupAccounts will create Model and set QMLer and UserManager
func (a *Accounts) SetupAccounts(qml QMLer, um types.UserManager, restarter types.Restarter) {
	a.Model = NewAccountsModel(nil)
	a.qml = qml
	a.um = um
	a.restarter = restarter
}

// LoadAccounts refreshes the current account list in GUI
func (a *Accounts) LoadAccounts() {
	a.accountMutex.Lock()
	defer a.accountMutex.Unlock()

	a.Model.Clear()

	users := a.um.GetUsers()

	// If there are no active accounts.
	if len(users) == 0 {
		log.Info("No active accounts")
		return
	}
	for _, user := range users {
		accInfo := NewAccountInfo(nil)
		username := user.Username()
		if username == "" {
			username = user.ID()
		}
		accInfo.SetAccount(username)

		// Set status.
		if user.IsConnected() {
			accInfo.SetStatus("connected")
		} else {
			accInfo.SetStatus("disconnected")
		}

		// Set login info.
		accInfo.SetUserID(user.ID())
		accInfo.SetHostname(bridge.Host)
		accInfo.SetPassword(user.GetBridgePassword())
		if a.settings != nil {
			accInfo.SetPortIMAP(a.settings.GetInt(settings.IMAPPortKey))
			accInfo.SetPortSMTP(a.settings.GetInt(settings.SMTPPortKey))
		}

		// Set aliases.
		accInfo.SetAliases(strings.Join(user.GetAddresses(), ";"))
		accInfo.SetIsExpanded(user.ID() == a.LatestUserID)
		accInfo.SetIsCombinedAddressMode(user.IsCombinedAddressMode())

		a.Model.addAccount(accInfo)
	}

	// Updated can clear.
	a.LatestUserID = ""
}

// ClearCache signal to remove all DB files
func (a *Accounts) ClearCache() {
	defer a.qml.ProcessFinished()
	if err := a.um.ClearData(); err != nil {
		log.Error("While clearing cache: ", err)
	}
	// Clearing data removes everything (db, preferences, ...)
	// so everything has to be stopped and started again.
	a.restarter.SetToRestart()
	a.qml.Quit()
}

// ClearKeychain signal remove all accounts from keychains
func (a *Accounts) ClearKeychain() {
	defer a.qml.ProcessFinished()
	for _, user := range a.um.GetUsers() {
		if err := a.um.DeleteUser(user.ID(), false); err != nil {
			log.Error("While deleting user: ", err)
			if err == keychain.ErrNoKeychain { // Probably not needed anymore.
				a.qml.NotifyHasNoKeychain()
			}
		}
	}
}

// LogoutAccount signal to remove account
func (a *Accounts) LogoutAccount(iAccount int) {
	defer a.qml.ProcessFinished()
	userID := a.Model.get(iAccount).UserID()
	user, err := a.um.GetUser(userID)
	if err != nil {
		log.Error("While logging out ", userID, ": ", err)
		return
	}
	if err := user.Logout(); err != nil {
		log.Error("While logging out ", userID, ": ", err)
	}
}

func (a *Accounts) showLoginError(err error, scope string) bool {
	if err == nil {
		a.qml.SetConnectionStatus(true) // If we are here connection is ok.
		return false
	}
	log.Warnf("%s: %v", scope, err)
	if err == pmapi.ErrNoConnection {
		a.qml.SetConnectionStatus(false)
		SendNotification(a.qml, TabAccount, a.qml.CanNotReachAPI())
		a.qml.ProcessFinished()
		return true
	}
	a.qml.SetConnectionStatus(true) // If we are here connection is ok.
	if err == pmapi.ErrUpgradeApplication {
		return true
	}
	a.qml.SetAddAccountWarning(err.Error(), -1)
	return true
}

// Login signal returns:
// -1: when error occurred
//  0: when no 2FA and no MBOX
//  1: when has 2FA
//  2: when has no 2FA but have MBOX
func (a *Accounts) Login(login, password string) int {
	var err error
	a.authClient, a.auth, err = a.um.Login(login, password)
	if a.showLoginError(err, "login") {
		return -1
	}
	if a.auth.HasTwoFactor() {
		return 1
	}
	if a.auth.HasMailboxPassword() {
		return 2
	}
	return 0 // No 2FA, no mailbox password.
}

// Auth2FA returns:
//  -1 : error (use SetAddAccountWarning to show message)
//   0 : single password mode
//   1 : two password mode
func (a *Accounts) Auth2FA(twoFacAuth string) int {
	var err error
	if a.auth == nil || a.authClient == nil {
		err = fmt.Errorf("missing authentication in auth2FA %p %p", a.auth, a.authClient)
	} else {
		err = a.authClient.Auth2FA(context.Background(), twoFacAuth)
	}

	if a.showLoginError(err, "auth2FA") {
		return -1
	}

	if a.auth.HasMailboxPassword() {
		return 1 // Ask for mailbox password.
	}
	return 0 // One password.
}

// AddAccount signal to add an account. It should close login modal
// ProcessFinished if ok.
func (a *Accounts) AddAccount(mailboxPassword string) int {
	if a.auth == nil || a.authClient == nil {
		log.Errorf("Missing authentication in addAccount %p %p", a.auth, a.authClient)
		a.qml.SetAddAccountWarning(a.qml.WrongMailboxPassword(), -2)
		return -1
	}

	user, err := a.um.FinishLogin(a.authClient, a.auth, mailboxPassword)
	if err != nil {
		log.WithError(err).Error("Login was unsuccessful")
		a.qml.SetAddAccountWarning("Failure: "+err.Error(), -2)
		return -1
	}

	a.LatestUserID = user.ID()
	a.qml.EmitEvent(events.UserRefreshEvent, user.ID())
	a.qml.ProcessFinished()
	return 0
}

// DeleteAccount by index in Model
func (a *Accounts) DeleteAccount(iAccount int, removePreferences bool) {
	defer a.qml.ProcessFinished()
	userID := a.Model.get(iAccount).UserID()
	if err := a.um.DeleteUser(userID, removePreferences); err != nil {
		log.Warn("deleteUser: cannot remove user: ", err)
		if err == keychain.ErrNoKeychain {
			a.qml.NotifyHasNoKeychain()
			return
		}
		SendNotification(a.qml, TabSettings, err.Error())
		return
	}
}
