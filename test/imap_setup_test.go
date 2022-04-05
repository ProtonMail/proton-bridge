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

func IMAPSetupFeatureContext(s *godog.ScenarioContext) {
	s.Step(`^there is IMAP client logged in as "([^"]*)"$`, thereIsIMAPClientLoggedInAs)
	s.Step(`^there is IMAP client "([^"]*)" logged in as "([^"]*)"$`, thereIsIMAPClientNamedLoggedInAs)
	s.Step(`^there is IMAP client logged in as "([^"]*)" with address "([^"]*)"$`, thereIsIMAPClientLoggedInAsWithAddress)
	s.Step(`^there is IMAP client "([^"]*)" logged in as "([^"]*)" with address "([^"]*)"$`, thereIsIMAPClientNamedLoggedInAsWithAddress)
	s.Step(`^there is IMAP client selected in "([^"]*)"$`, thereIsIMAPClientSelectedIn)
	s.Step(`^there is IMAP client "([^"]*)" selected in "([^"]*)"$`, thereIsIMAPClientNamedSelectedIn)
}

func thereIsIMAPClientLoggedInAs(bddUserID string) error {
	return thereIsIMAPClientNamedLoggedInAs("imap", bddUserID)
}

func thereIsIMAPClientNamedLoggedInAs(clientID, bddUserID string) (err error) {
	account := ctx.GetTestAccount(bddUserID)
	if account == nil {
		return godog.ErrPending
	}
	ctx.GetIMAPClient(clientID).Login(account.Address(), account.BridgePassword()).AssertOK()
	return ctx.GetTestingError()
}

func thereIsIMAPClientLoggedInAsWithAddress(bddUserID, bddAddressID string) error {
	return thereIsIMAPClientNamedLoggedInAsWithAddress("imap", bddUserID, bddAddressID)
}

func thereIsIMAPClientNamedLoggedInAsWithAddress(clientID, bddUserID, bddAddressID string) error {
	account := ctx.GetTestAccountWithAddress(bddUserID, bddAddressID)
	if account == nil {
		return godog.ErrPending
	}
	ctx.GetIMAPClient(clientID).Login(account.Address(), account.BridgePassword()).AssertOK()
	return nil
}

func thereIsIMAPClientSelectedIn(mailboxName string) error {
	return thereIsIMAPClientNamedSelectedIn("imap", mailboxName)
}

func thereIsIMAPClientNamedSelectedIn(clientID, mailboxName string) (err error) {
	ctx.GetIMAPClient(clientID).Select(mailboxName).AssertOK()
	return ctx.GetTestingError()
}
