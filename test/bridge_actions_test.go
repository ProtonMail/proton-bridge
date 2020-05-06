// Copyright (c) 2020 Proton Technologies AG
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
)

func BridgeActionsFeatureContext(s *godog.Suite) {
	s.Step(`^bridge starts$`, bridgeStarts)
	s.Step(`^bridge syncs "([^"]*)"$`, bridgeSyncsUser)
	s.Step(`^"([^"]*)" logs in to bridge$`, userLogsInToBridge)
	s.Step(`^"([^"]*)" logs in to bridge with bad password$`, userLogsInToBridgeWithBadPassword)
	s.Step(`^"([^"]*)" logs out from bridge$`, userLogsOutFromBridge)
	s.Step(`^"([^"]*)" changes the address mode$`, userChangesTheAddressMode)
	s.Step(`^user deletes "([^"]*)" from bridge$`, userDeletesUserFromBridge)
	s.Step(`^user deletes "([^"]*)" from bridge with cache$`, userDeletesUserFromBridgeWithCache)
	s.Step(`^the internet connection is lost$`, theInternetConnectionIsLost)
	s.Step(`^the internet connection is restored$`, theInternetConnectionIsRestored)
	s.Step(`^(\d+) seconds pass$`, secondsPass)
	s.Step(`^"([^"]*)" swaps address "([^"]*)" with address "([^"]*)"$`, swapsAddressWithAddress)
}

func bridgeStarts() error {
	ctx.SetLastBridgeError(ctx.RestartBridge())
	return nil
}

func bridgeSyncsUser(bddUserID string) error {
	account := ctx.GetTestAccount(bddUserID)
	if account == nil {
		return godog.ErrPending
	}
	if err := ctx.WaitForSync(account.Username()); err != nil {
		return internalError(err, "waiting for sync")
	}
	ctx.SetLastBridgeError(ctx.GetTestingError())
	return nil
}

func userLogsInToBridge(bddUserID string) error {
	account := ctx.GetTestAccount(bddUserID)
	if account == nil {
		return godog.ErrPending
	}
	ctx.SetLastBridgeError(ctx.LoginUser(account.Username(), account.Password(), account.MailboxPassword()))
	return nil
}

func userLogsInToBridgeWithBadPassword(bddUserID string) error {
	account := ctx.GetTestAccount(bddUserID)
	if account == nil {
		return godog.ErrPending
	}
	ctx.SetLastBridgeError(ctx.LoginUser(account.Username(), "you shall not pass!", "123"))
	return nil
}

func userLogsOutFromBridge(bddUserID string) error {
	account := ctx.GetTestAccount(bddUserID)
	if account == nil {
		return godog.ErrPending
	}
	ctx.SetLastBridgeError(ctx.LogoutUser(account.Username()))
	return nil
}

func userChangesTheAddressMode(bddUserID string) error {
	account := ctx.GetTestAccount(bddUserID)
	if account == nil {
		return godog.ErrPending
	}
	bridgeUser, err := ctx.GetUser(account.Username())
	if err != nil {
		return internalError(err, "getting user %s", account.Username())
	}
	if err := bridgeUser.SwitchAddressMode(); err != nil {
		return err
	}

	ctx.EventuallySyncIsFinishedForUsername(account.Username())
	return nil
}

func userDeletesUserFromBridge(bddUserID string) error {
	return deleteUserFromBridge(bddUserID, false)
}

func userDeletesUserFromBridgeWithCache(bddUserID string) error {
	return deleteUserFromBridge(bddUserID, true)
}

func deleteUserFromBridge(bddUserID string, cache bool) error {
	account := ctx.GetTestAccount(bddUserID)
	if account == nil {
		return godog.ErrPending
	}
	bridgeUser, err := ctx.GetUser(account.Username())
	if err != nil {
		return internalError(err, "getting user %s", account.Username())
	}
	return ctx.GetBridge().DeleteUser(bridgeUser.ID(), cache)
}

func theInternetConnectionIsLost() error {
	ctx.GetPMAPIController().TurnInternetConnectionOff()
	return nil
}

func theInternetConnectionIsRestored() error {
	ctx.GetPMAPIController().TurnInternetConnectionOn()
	return nil
}

func secondsPass(seconds int) error {
	time.Sleep(time.Duration(seconds) * time.Second)
	return nil
}

func swapsAddressWithAddress(bddUserID, bddAddressID1, bddAddressID2 string) error {
	account := ctx.GetTestAccount(bddUserID)
	if account == nil {
		return godog.ErrPending
	}

	address1ID := account.GetAddressID(bddAddressID1)
	address2ID := account.GetAddressID(bddAddressID2)

	var address1Index, address2Index int
	var addressIDs []string
	for i, v := range *account.Addresses() {
		if v.ID == address1ID {
			address1Index = i
		}
		if v.ID == address2ID {
			address2Index = i
		}
		addressIDs = append(addressIDs, v.ID)
	}

	addressIDs[address1Index], addressIDs[address2Index] = addressIDs[address2Index], addressIDs[address1Index]

	ctx.ReorderAddresses(account.Username(), bddAddressID1, bddAddressID2)
	ctx.GetPMAPIController().ReorderAddresses(account.User(), addressIDs)

	return nil
}
