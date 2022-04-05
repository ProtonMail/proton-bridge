// Copyright (c) 2022 Proton AG
//
// This file is part of Proton Mail Bridge.Bridge.
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

package tests

import (
	"github.com/cucumber/godog"
)

func IMAPActionsAuthFeatureContext(s *godog.ScenarioContext) {
	s.Step(`^IMAP client authenticates "([^"]*)"$`, imapClientAuthenticates)
	s.Step(`^IMAP client "([^"]*)" authenticates "([^"]*)"$`, imapClientNamedAuthenticates)
	s.Step(`^IMAP client authenticates "([^"]*)" with address "([^"]*)"$`, imapClientAuthenticatesWithAddress)
	s.Step(`^IMAP client "([^"]*)" authenticates "([^"]*)" with address "([^"]*)"$`, imapClientNamedAuthenticatesWithAddress)
	s.Step(`^IMAP client authenticates "([^"]*)" with bad password$`, imapClientAuthenticatesWithBadPassword)
	s.Step(`^IMAP client authenticates with username "([^"]*)" and password "([^"]*)"$`, imapClientAuthenticatesWithUsernameAndPassword)
	s.Step(`^IMAP client logs out$`, imapClientLogsOut)
}

func imapClientAuthenticates(bddUserID string) error {
	return imapClientNamedAuthenticates("imap", bddUserID)
}

func imapClientNamedAuthenticates(clientID, bddUserID string) error {
	account := ctx.GetTestAccount(bddUserID)
	if account == nil {
		return godog.ErrPending
	}
	res := ctx.GetIMAPClient(clientID).Login(account.Address(), account.BridgePassword())
	ctx.SetIMAPLastResponse(clientID, res)
	return nil
}

func imapClientAuthenticatesWithAddress(bddUserID, bddAddressID string) error {
	return imapClientNamedAuthenticatesWithAddress("imap", bddUserID, bddAddressID)
}

func imapClientNamedAuthenticatesWithAddress(clientID, bddUserID, bddAddressID string) error {
	account := ctx.GetTestAccountWithAddress(bddUserID, bddAddressID)
	if account == nil {
		return godog.ErrPending
	}
	res := ctx.GetIMAPClient(clientID).Login(account.Address(), account.BridgePassword())
	ctx.SetIMAPLastResponse(clientID, res)
	return nil
}

func imapClientAuthenticatesWithBadPassword(bddUserID string) error {
	account := ctx.GetTestAccount(bddUserID)
	if account == nil {
		return godog.ErrPending
	}
	res := ctx.GetIMAPClient("imap").Login(account.Address(), "you shall not pass!")
	ctx.SetIMAPLastResponse("imap", res)
	return nil
}

func imapClientAuthenticatesWithUsernameAndPassword(username, password string) error {
	res := ctx.GetIMAPClient("imap").Login(username, password)
	ctx.SetIMAPLastResponse("imap", res)
	return nil
}

func imapClientLogsOut() error {
	res := ctx.GetIMAPClient("imap").Logout()
	ctx.SetIMAPLastResponse("imap", res)
	return nil
}
