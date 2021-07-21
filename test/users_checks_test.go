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
	"time"

	"github.com/cucumber/godog"
	a "github.com/stretchr/testify/assert"
)

func UsersChecksFeatureContext(s *godog.Suite) {
	s.Step(`^"([^"]*)" has address mode in "([^"]*)" mode$`, userHasAddressModeInMode)
	s.Step(`^"([^"]*)" is disconnected$`, userIsDisconnected)
	s.Step(`^"([^"]*)" is connected$`, userIsConnected)
	s.Step(`^"([^"]*)" has database file$`, userHasDatabaseFile)
	s.Step(`^"([^"]*)" does not have database file$`, userDoesNotHaveDatabaseFile)
	s.Step(`^"([^"]*)" has loaded store$`, userHasLoadedStore)
	s.Step(`^"([^"]*)" does not have loaded store$`, userDoesNotHaveLoadedStore)
	s.Step(`^"([^"]*)" has running event loop$`, userHasRunningEventLoop)
	s.Step(`^"([^"]*)" does not have running event loop$`, userDoesNotHaveRunningEventLoop)
}

func userHasAddressModeInMode(bddUserID, wantAddressMode string) error {
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
	a.Equal(ctx.GetTestingT(), wantAddressMode, addressMode)
	return ctx.GetTestingError()
}

func userIsDisconnected(bddUserID string) error {
	account := ctx.GetTestAccount(bddUserID)
	if account == nil {
		return godog.ErrPending
	}
	user, err := ctx.GetUser(account.Username())
	if err != nil {
		return internalError(err, "getting user %s", account.Username())
	}
	a.Eventually(ctx.GetTestingT(), func() bool {
		return !user.IsConnected()
	}, 5*time.Second, 10*time.Millisecond)
	return ctx.GetTestingError()
}

func userIsConnected(bddUserID string) error {
	account := ctx.GetTestAccount(bddUserID)
	if account == nil {
		return godog.ErrPending
	}
	t := ctx.GetTestingT()
	user, err := ctx.GetUser(account.Username())
	if err != nil {
		return internalError(err, "getting user %s", account.Username())
	}
	a.Eventually(ctx.GetTestingT(), user.IsConnected, 5*time.Second, 10*time.Millisecond)
	a.NotEmpty(t, user.GetPrimaryAddress())
	a.NotEmpty(t, user.GetStoreAddresses())
	return ctx.GetTestingError()
}

func userHasDatabaseFile(bddUserID string) error {
	account := ctx.GetTestAccount(bddUserID)
	if account == nil {
		return godog.ErrPending
	}
	filePath := ctx.GetDatabaseFilePath(account.UserID())
	a.FileExists(ctx.GetTestingT(), filePath)
	return ctx.GetTestingError()
}

func userDoesNotHaveDatabaseFile(bddUserID string) error {
	account := ctx.GetTestAccount(bddUserID)
	if account == nil {
		return godog.ErrPending
	}
	filePath := ctx.GetDatabaseFilePath(account.UserID())
	a.NoFileExists(ctx.GetTestingT(), filePath)
	return ctx.GetTestingError()
}

func userHasLoadedStore(bddUserID string) error {
	account := ctx.GetTestAccount(bddUserID)
	if account == nil {
		return godog.ErrPending
	}
	user, err := ctx.GetUser(account.Username())
	if err != nil {
		return internalError(err, "getting user %s", account.Username())
	}
	a.NotNil(ctx.GetTestingT(), user.GetStore())
	return ctx.GetTestingError()
}

func userDoesNotHaveLoadedStore(bddUserID string) error {
	account := ctx.GetTestAccount(bddUserID)
	if account == nil {
		return godog.ErrPending
	}
	user, err := ctx.GetUser(account.Username())
	if err != nil {
		return internalError(err, "getting user %s", account.Username())
	}
	a.Nil(ctx.GetTestingT(), user.GetStore())
	return ctx.GetTestingError()
}

func userHasRunningEventLoop(bddUserID string) error {
	account := ctx.GetTestAccount(bddUserID)
	if account == nil {
		return godog.ErrPending
	}
	store, err := ctx.GetStore(account.Username())
	if err != nil {
		return internalError(err, "getting store of %s", account.Username())
	}
	a.Eventually(ctx.GetTestingT(), func() bool {
		return store.TestGetEventLoop() != nil && store.TestGetEventLoop().IsRunning()
	}, 5*time.Second, 10*time.Millisecond)
	return ctx.GetTestingError()
}

func userDoesNotHaveRunningEventLoop(bddUserID string) error {
	account := ctx.GetTestAccount(bddUserID)
	if account == nil {
		return godog.ErrPending
	}
	store, err := ctx.GetStore(account.Username())
	if err != nil {
		return internalError(err, "getting store of %s", account.Username())
	}
	if store == nil {
		return nil
	}
	a.Eventually(ctx.GetTestingT(), func() bool {
		return store.TestGetEventLoop() == nil || !store.TestGetEventLoop().IsRunning()
	}, 5*time.Second, 10*time.Millisecond)
	return ctx.GetTestingError()
}
