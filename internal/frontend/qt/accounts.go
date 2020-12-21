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

// +build !nogui

package qt

import (
	"fmt"
	"strings"

	"github.com/ProtonMail/proton-bridge/internal/bridge"
	"github.com/ProtonMail/proton-bridge/internal/config/settings"
	"github.com/ProtonMail/proton-bridge/internal/events"
	"github.com/ProtonMail/proton-bridge/pkg/keychain"
	pmapi "github.com/ProtonMail/proton-bridge/pkg/pmapi"
)

func (s *FrontendQt) loadAccounts() {
	accountMutex.Lock()
	defer accountMutex.Unlock()

	// Update users.
	s.Accounts.Clear()

	users := s.bridge.GetUsers()

	// If there are no active accounts.
	if len(users) == 0 {
		log.Info("No active accounts")
		return
	}
	for _, user := range users {
		acc_info := NewAccountInfo(nil)
		username := user.Username()
		if username == "" {
			username = user.ID()
		}
		acc_info.SetAccount(username)

		// Set status.
		if user.IsConnected() {
			acc_info.SetStatus("connected")
		} else {
			acc_info.SetStatus("disconnected")
		}

		// Set login info.
		acc_info.SetUserID(user.ID())
		acc_info.SetHostname(bridge.Host)
		acc_info.SetPassword(user.GetBridgePassword())
		acc_info.SetPortIMAP(s.settings.GetInt(settings.IMAPPortKey))
		acc_info.SetPortSMTP(s.settings.GetInt(settings.SMTPPortKey))

		// Set aliases.
		acc_info.SetAliases(strings.Join(user.GetAddresses(), ";"))
		acc_info.SetIsExpanded(user.ID() == s.userIDAdded)
		acc_info.SetIsCombinedAddressMode(user.IsCombinedAddressMode())

		s.Accounts.addAccount(acc_info)
	}

	// Updated can clear.
	s.userIDAdded = ""
}

func (s *FrontendQt) clearCache() {
	defer s.Qml.ProcessFinished()
	if err := s.bridge.ClearData(); err != nil {
		log.Error("While clearing cache: ", err)
	}
	// Clearing data removes everything (db, preferences, ...)
	// so everything has to be stopped and started again.
	s.restarter.SetToRestart()
	s.App.Quit()
}

func (s *FrontendQt) clearKeychain() {
	defer s.Qml.ProcessFinished()
	for _, user := range s.bridge.GetUsers() {
		if err := s.bridge.DeleteUser(user.ID(), false); err != nil {
			log.Error("While deleting user: ", err)
			if err == keychain.ErrNoKeychain { // Probably not needed anymore.
				s.Qml.NotifyHasNoKeychain()
			}
		}
	}
}

func (s *FrontendQt) logoutAccount(iAccount int) {
	defer s.Qml.ProcessFinished()
	userID := s.Accounts.get(iAccount).UserID()
	user, err := s.bridge.GetUser(userID)
	if err != nil {
		log.Error("While logging out ", userID, ": ", err)
		return
	}
	if err := user.Logout(); err != nil {
		log.Error("While logging out ", userID, ": ", err)
	}
}

func (s *FrontendQt) showLoginError(err error, scope string) bool {
	if err == nil {
		s.Qml.SetConnectionStatus(true) // If we are here connection is ok.
		return false
	}
	log.Warnf("%s: %v", scope, err)
	if err == pmapi.ErrAPINotReachable {
		s.Qml.SetConnectionStatus(false)
		s.SendNotification(TabAccount, s.Qml.CanNotReachAPI())
		s.Qml.ProcessFinished()
		return true
	}
	s.Qml.SetConnectionStatus(true) // If we are here connection is ok.
	if err == pmapi.ErrUpgradeApplication {
		s.eventListener.Emit(events.UpgradeApplicationEvent, "")
		return true
	}
	s.Qml.SetAddAccountWarning(err.Error(), -1)
	return true
}

// login returns:
// -1: when error occurred
//  0: when no 2FA and no MBOX
//  1: when has 2FA
//  2: when has no 2FA but have MBOX
func (s *FrontendQt) login(login, password string) int {
	var err error
	s.authClient, s.auth, err = s.bridge.Login(login, password)
	if s.showLoginError(err, "login") {
		return -1
	}
	if s.auth.HasTwoFactor() {
		return 1
	}
	if s.auth.HasMailboxPassword() {
		return 2
	}
	return 0 // No 2FA, no mailbox password.
}

// auth2FA returns:
//  -1 : error (use SetAddAccountWarning to show message)
//   0 : single password mode
//   1 : two password mode
func (s *FrontendQt) auth2FA(twoFacAuth string) int {
	var err error
	if s.auth == nil || s.authClient == nil {
		err = fmt.Errorf("missing authentication in auth2FA %p %p", s.auth, s.authClient)
	} else {
		err = s.authClient.Auth2FA(twoFacAuth, s.auth)
	}

	if s.showLoginError(err, "auth2FA") {
		return -1
	}

	if s.auth.HasMailboxPassword() {
		return 1 // Ask for mailbox password.
	}
	return 0 // One password.
}

// addAccount adds an account. It should close login modal ProcessFinished if ok.
func (s *FrontendQt) addAccount(mailboxPassword string) int {
	if s.auth == nil || s.authClient == nil {
		log.Errorf("Missing authentication in addAccount %p %p", s.auth, s.authClient)
		s.Qml.SetAddAccountWarning(s.Qml.WrongMailboxPassword(), -2)
		return -1
	}

	user, err := s.bridge.FinishLogin(s.authClient, s.auth, mailboxPassword)
	if err != nil {
		log.WithError(err).Error("Login was unsuccessful")
		s.Qml.SetAddAccountWarning("Failure: "+err.Error(), -2)
		return -1
	}

	s.userIDAdded = user.ID()
	s.eventListener.Emit(events.UserRefreshEvent, user.ID())
	s.Qml.ProcessFinished()
	return 0
}

func (s *FrontendQt) deleteAccount(iAccount int, removePreferences bool) {
	defer s.Qml.ProcessFinished()
	userID := s.Accounts.get(iAccount).UserID()
	if err := s.bridge.DeleteUser(userID, removePreferences); err != nil {
		log.Warn("deleteUser: cannot remove user: ", err)
		if err == keychain.ErrNoKeychain {
			s.Qml.NotifyHasNoKeychain()
			return
		}
		s.SendNotification(TabSettings, err.Error())
		return
	}
}
