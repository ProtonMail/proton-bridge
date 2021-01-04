// Copyright (c) 2021 Proton Technologies AG
//
// This file is part of ProtonMail Bridge.Bridge.
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

package tests

import (
	"os"
	"time"

	"github.com/cucumber/godog"
	a "github.com/stretchr/testify/assert"
)

func UsersSetupFeatureContext(s *godog.Suite) {
	s.Step(`^there is user "([^"]*)"$`, thereIsUser)
	s.Step(`^there is connected user "([^"]*)"$`, thereIsConnectedUser)
	s.Step(`^there is disconnected user "([^"]*)"$`, thereIsDisconnectedUser)
	s.Step(`^there is database file for "([^"]*)"$`, thereIsDatabaseFileForUser)
	s.Step(`^there is no database file for "([^"]*)"$`, thereIsNoDatabaseFileForUser)
	s.Step(`^there is "([^"]*)" in "([^"]*)" address mode$`, thereIsUserWithAddressMode)
}

func thereIsUser(bddUserID string) error {
	account := ctx.GetTestAccount(bddUserID)
	if account == nil {
		return godog.ErrPending
	}
	err := ctx.GetPMAPIController().AddUser(account.User(), account.Addresses(), account.Password(), account.IsTwoFAEnabled())
	return internalError(err, "adding user %s", account.Username())
}

func thereIsConnectedUser(bddUserID string) error {
	account := ctx.GetTestAccount(bddUserID)
	if account == nil {
		return godog.ErrPending
	}
	err := ctx.GetPMAPIController().AddUser(account.User(), account.Addresses(), account.Password(), account.IsTwoFAEnabled())
	if err != nil {
		return internalError(err, "adding user %s", account.Username())
	}
	return ctx.LoginUser(account.Username(), account.Password(), account.MailboxPassword())
}

func thereIsDisconnectedUser(bddUserID string) error {
	account := ctx.GetTestAccount(bddUserID)
	if account == nil {
		return godog.ErrPending
	}
	err := ctx.GetPMAPIController().AddUser(account.User(), account.Addresses(), account.Password(), account.IsTwoFAEnabled())
	if err != nil {
		return internalError(err, "adding user %s", account.Username())
	}
	err = ctx.LoginUser(account.Username(), account.Password(), account.MailboxPassword())
	if err != nil {
		return internalError(err, "logging user %s in", account.Username())
	}
	user, err := ctx.GetUser(account.Username())
	if err != nil {
		return internalError(err, "getting user %s", account.Username())
	}
	err = user.Logout()
	if err != nil {
		return internalError(err, "disconnecting user %s", account.Username())
	}

	// We need to wait till event loop is stopped because when it's stopped
	// logout is also called and if we would do login at the same time, it
	// wouldn't work. 100 ms after event loop is stopped should be enough.
	a.Eventually(ctx.GetTestingT(), func() bool {
		store := user.GetStore()
		if store == nil {
			return true
		}
		return !store.TestGetEventLoop().IsRunning()
	}, 1*time.Second, 10*time.Millisecond)
	time.Sleep(100 * time.Millisecond)
	return ctx.GetTestingError()
}

func thereIsDatabaseFileForUser(bddUserID string) error {
	account := ctx.GetTestAccount(bddUserID)
	if account == nil {
		return godog.ErrPending
	}
	filePath := ctx.GetDatabaseFilePath(account.UserID())
	_, err := os.Stat(filePath)
	return internalError(err, "getting database file of %s", account.Username())
}

func thereIsNoDatabaseFileForUser(bddUserID string) error {
	account := ctx.GetTestAccount(bddUserID)
	if account == nil {
		return godog.ErrPending
	}
	filePath := ctx.GetDatabaseFilePath(account.UserID())
	if _, err := os.Stat(filePath); err != nil {
		return nil
	}
	return internalError(os.Remove(filePath), "removing database file of %s", account.Username())
}

func thereIsUserWithAddressMode(bddUserID, wantAddressMode string) error {
	account := ctx.GetTestAccount(bddUserID)
	if account == nil {
		return godog.ErrPending
	}
	user, err := ctx.GetUser(account.Username())
	if err != nil {
		return internalError(err, "getting user %s", account.Username())
	}
	addressMode := "split"
	if user.IsCombinedAddressMode() {
		addressMode = "combined"
	}
	if wantAddressMode != addressMode {
		err := user.SwitchAddressMode()
		if err != nil {
			return internalError(err, "switching mode")
		}
	}
	ctx.EventuallySyncIsFinishedForUsername(user.Username())
	return nil
}
