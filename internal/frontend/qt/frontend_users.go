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
	"context"
	"encoding/base64"
	"time"

	"github.com/ProtonMail/proton-bridge/v2/internal/events"
	"github.com/ProtonMail/proton-bridge/v2/internal/users"
	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
)

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
		if err == pmapi.ErrPasswordWrong {
			// Remove error message since it is hardcodded in QML.
			f.qml.LoginUsernamePasswordError("")
			f.loginClean()
			return
		}
		if err == pmapi.ErrPaidPlanRequired {
			f.qml.LoginFreeUserError()
			f.loginClean()
			return
		}
		f.qml.LoginUsernamePasswordError(err.Error())
		f.loginClean()
		return
	}

	if f.auth.HasTwoFactor() {
		f.qml.Login2FARequested(username)
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

	if len(f.password) == 0 || f.auth == nil || f.authClient == nil {
		f.log.
			WithField("hasPass", len(f.password) != 0).
			WithField("hasAuth", f.auth != nil).
			WithField("hasClient", f.authClient != nil).
			Error("Finish login: authethication incomplete")
		f.qml.Login2PasswordErrorAbort("Missing authentication, try again.")
		return
	}

	done := make(chan string)
	f.eventListener.Add(events.UserChangeDone, done)
	defer f.eventListener.Remove(events.UserChangeDone, done)

	user, err := f.bridge.FinishLogin(f.authClient, f.auth, f.password)

	if err != nil && err != users.ErrUserAlreadyConnected {
		f.log.WithError(err).Errorf("Finish login failed")
		f.qml.Login2PasswordErrorAbort(err.Error())
		return
	}

	// The user changed should be triggerd by FinishLogin but it is not
	// guaranteed when this is going to happen. Therefore we should wait
	// until we receive the signal from userChanged function.
	f.waitForUserChangeDone(done, user.ID())

	index := f.qml.Users().indexByID(user.ID())
	f.log.WithField("index", index).Debug("Login finished")
	defer f.qml.LoginFinished(index)

	if err == users.ErrUserAlreadyConnected {
		f.log.WithError(err).Error("User already logged in")
		f.qml.LoginAlreadyLoggedIn(index)
	}
}

func (f *FrontendQt) waitForUserChangeDone(done <-chan string, userID string) {
	for {
		select {
		case changedID := <-done:
			if changedID == userID {
				return
			}
		case <-time.After(2 * time.Second):
			f.log.WithField("ID", userID).Warning("Login finished but user not added within 2 seconds")
			return
		}
	}
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
