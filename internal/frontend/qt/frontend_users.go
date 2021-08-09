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

package qt

import (
	"context"
	"encoding/base64"
	"github.com/ProtonMail/proton-bridge/internal/frontend/types"
	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
)

func (f *FrontendQt) loadUsers() {
	f.usersMtx.Lock()
	defer f.usersMtx.Unlock()

	f.qml.Users().clear()

	for _, user := range f.bridge.GetUsers() {
		f.qml.Users().addUser(newQMLUserFromBacked(f, user))
	}

	// If there are no active accounts.
	if f.qml.Users().Count() == 0 {
		f.log.Info("No active accounts")
	}
}

func (f *FrontendQt) userChanged(userID string) {
	f.usersMtx.Lock()
	defer f.usersMtx.Unlock()

	fUsers := f.qml.Users()

	index := fUsers.indexByID(userID)
	user, err := f.bridge.GetUser(userID)

	if user == nil || err != nil {
		if index >= 0 { // delete existing user
			fUsers.removeUser(index)
		}
		return
	}

	if index < 0 { // add non-existing user
		fUsers.addUser(newQMLUserFromBacked(f, user))
		return
	}

	// update exiting user
	fUsers.users[index].update(user)
}

func newQMLUserFromBacked(f *FrontendQt, user types.User) *QMLUser {
	qu := NewQMLUser(nil)
	qu.ID = user.ID()

	qu.update(user)

	qu.ConnectToggleSplitMode(func(activateSplitMode bool) {
		go func() {
			defer qu.ToggleSplitModeFinished()
			if activateSplitMode == user.IsCombinedAddressMode() {
				user.SwitchAddressMode()
			}
			qu.SetSplitMode(!user.IsCombinedAddressMode())
		}()
	})

	qu.ConnectLogout(func() {
		qu.SetLoggedIn(false)
		go user.Logout()
	})

	qu.ConnectConfigureAppleMail(func(address string) {
		go f.configureAppleMail(qu.ID, address)
	})

	return qu
}

func (f *FrontendQt) login(username, password string) {
	var err error
	f.password, err = base64.StdEncoding.DecodeString(password)
	if err != nil {
		f.log.WithError(err).Error("Cannot decode password")
		f.qml.LoginUsernamePasswordError("Cannot decode password")
		f.loginClean()
		return
	}

	f.authClient, f.auth, err = f.bridge.Login(username, f.password)
	if err != nil {
		f.qml.LoginUsernamePasswordError(err.Error())
		f.loginClean()
		return
	}

	if f.auth.HasTwoFactor() {
		f.qml.Login2FARequested()
		return
	}
	if f.auth.HasMailboxPassword() {
		f.qml.Login2PasswordRequested()
		return
	}

	f.finishLogin()
}

func (f *FrontendQt) login2FA(username, code string) {
	if f.auth == nil || f.authClient == nil {
		f.log.Errorf("Login 2FA: authethication incomplete %p %p", f.auth, f.authClient)
		f.qml.Login2FAErrorAbort("Missing authentication, try again.")
		f.loginClean()
		return
	}

	twoFA, err := base64.StdEncoding.DecodeString(code)
	if err != nil {
		f.log.WithError(err).Error("Cannot decode 2fa code")
		f.qml.LoginUsernamePasswordError("Cannot decode 2fa code")
		f.loginClean()
		return
	}

	err = f.authClient.Auth2FA(context.Background(), string(twoFA))
	if err == pmapi.ErrBad2FACodeTryAgain {
		f.log.Warn("Login 2FA: retry 2fa")
		f.qml.Login2FAError("")
		return
	}

	if err == pmapi.ErrBad2FACode {
		f.log.Warn("Login 2FA: abort 2fa")
		f.qml.Login2FAErrorAbort("")
		f.loginClean()
		return
	}

	if err != nil {
		f.log.WithError(err).Warn("Login 2FA: failed.")
		f.qml.Login2FAErrorAbort(err.Error())
		f.loginClean()
		return
	}

	if f.auth.HasMailboxPassword() {
		f.qml.Login2PasswordRequested()
		return
	}

	f.finishLogin()
}

func (f *FrontendQt) login2Password(username, mboxPassword string) {
	var err error
	f.password, err = base64.StdEncoding.DecodeString(mboxPassword)
	if err != nil {
		f.log.WithError(err).Error("Cannot decode mbox password")
		f.qml.LoginUsernamePasswordError("Cannot decode mbox password")
		f.loginClean()
		return
	}

	f.finishLogin()
}

func (f *FrontendQt) finishLogin() {
	defer f.loginClean()

	if f.auth == nil || f.authClient == nil {
		f.log.Errorf("Finish login: Authethication incomplete %p %p", f.auth, f.authClient)
		f.qml.Login2PasswordErrorAbort("Missing authentication, try again.")
		return
	}

	user, err := f.bridge.FinishLogin(f.authClient, f.auth, f.password)
	if err != nil {
		f.log.Errorf("Authethication incomplete %p %p", f.auth, f.authClient)
		f.qml.Login2PasswordErrorAbort("Missing authentication, try again.")
		return
	}

	index := f.qml.Users().indexByID(user.ID())
	if index < 0 {
		qu := newQMLUserFromBacked(f, user)
		qu.SetSetupGuideSeen(false)
		f.qml.Users().addUser(qu)
		return
	}

	f.qml.Users().users[index].update(user)
	f.qml.LoginFinished()
}

func (f *FrontendQt) loginAbort(username string) {
	f.loginClean()
}

func (f *FrontendQt) loginClean() {
	f.auth = nil
	f.authClient = nil
	for i := range f.password {
		f.password[i] = '\x00'
	}
	f.password = f.password[0:0]
}
