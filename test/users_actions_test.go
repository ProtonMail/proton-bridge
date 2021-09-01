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
	"github.com/cucumber/godog"
)

func UsersActionsFeatureContext(s *godog.Suite) {
	s.Step(`^"([^"]*)" logs in$`, userLogsIn)
	s.Step(`^"([^"]*)" logs in with bad password$`, userLogsInWithBadPassword)
	s.Step(`^"([^"]*)" logs out$`, userLogsOut)
	s.Step(`^"([^"]*)" changes the address mode$`, userChangesTheAddressMode)
	s.Step(`^user deletes "([^"]*)"$`, userDeletesUser)
	s.Step(`^user deletes "([^"]*)" with cache$`, userDeletesUserWithCache)
	s.Step(`^"([^"]*)" swaps address "([^"]*)" with address "([^"]*)"$`, swapsAddressWithAddress)
}

func userLogsIn(bddUserID string) error {
	account := ctx.GetTestAccount(bddUserID)
	if account == nil {
		return godog.ErrPending
	}
	ctx.SetLastError(ctx.LoginUser(account.Username(), account.Password(), account.MailboxPassword()))
	return nil
}

func userLogsInWithBadPassword(bddUserID string) error {
	account := ctx.GetTestAccount(bddUserID)
	if account == nil {
		return godog.ErrPending
	}
	ctx.SetLastError(ctx.LoginUser(account.Username(), []byte("you shall not pass!"), []byte("123")))
	return nil
}

func userLogsOut(bddUserID string) error {
	account := ctx.GetTestAccount(bddUserID)
	if account == nil {
		return godog.ErrPending
	}
	ctx.SetLastError(ctx.LogoutUser(account.Username()))
	return nil
}

func userChangesTheAddressMode(bddUserID string) error {
	account := ctx.GetTestAccount(bddUserID)
	if account == nil {
		return godog.ErrPending
	}
	user, err := ctx.GetUser(account.Username())
	if err != nil {
		return internalError(err, "getting user %s", account.Username())
	}
	if err := user.SwitchAddressMode(); err != nil {
		return err
	}

	ctx.EventuallySyncIsFinishedForUsername(account.Username())
	return nil
}

func userDeletesUser(bddUserID string) error {
	return deleteUser(bddUserID, false)
}

func userDeletesUserWithCache(bddUserID string) error {
	return deleteUser(bddUserID, true)
}

func deleteUser(bddUserID string, cache bool) error {
	account := ctx.GetTestAccount(bddUserID)
	if account == nil {
		return godog.ErrPending
	}
	user, err := ctx.GetUser(account.Username())
	if err != nil {
		return internalError(err, "getting user %s", account.Username())
	}
	ctx.SetLastError(ctx.GetUsers().DeleteUser(user.ID(), cache))
	return nil
}

func swapsAddressWithAddress(bddUserID, bddAddressID1, bddAddressID2 string) error {
	account := ctx.GetTestAccount(bddUserID)
	if account == nil {
		return godog.ErrPending
	}

	address1ID := account.GetAddressID(bddAddressID1)
	address2ID := account.GetAddressID(bddAddressID2)
	addressIDs := make([]string, len(*account.Addresses()))

	var address1Index, address2Index int
	for i, v := range *account.Addresses() {
		if v.ID == address1ID {
			address1Index = i
		}
		if v.ID == address2ID {
			address2Index = i
		}
		addressIDs[i] = v.ID
	}

	addressIDs[address1Index], addressIDs[address2Index] = addressIDs[address2Index], addressIDs[address1Index]

	ctx.ReorderAddresses(account.Username(), bddAddressID1, bddAddressID2)

	return ctx.GetPMAPIController().ReorderAddresses(account.User(), addressIDs)
}
