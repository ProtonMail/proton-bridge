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

func SMTPSetupFeatureContext(s *godog.ScenarioContext) {
	s.Step(`^there is SMTP client logged in as "([^"]*)"$`, thereIsSMTPClientLoggedInAs)
	s.Step(`^there is SMTP client "([^"]*)" logged in as "([^"]*)"$`, thereIsSMTPClientNamedLoggedInAs)
	s.Step(`^there is SMTP client logged in as "([^"]*)" with address "([^"]*)"$`, thereIsSMTPClientLoggedInAsWithAddress)
	s.Step(`^there is SMTP client "([^"]*)" logged in as "([^"]*)" with address "([^"]*)"$`, thereIsSMTPClientNamedLoggedInAsWithAddress)
}

func thereIsSMTPClientLoggedInAs(bddUserID string) error {
	return thereIsSMTPClientNamedLoggedInAs("smtp", bddUserID)
}

func thereIsSMTPClientNamedLoggedInAs(clientID, bddUserID string) error {
	account := ctx.GetTestAccount(bddUserID)
	if account == nil {
		return godog.ErrPending
	}
	ctx.GetSMTPClient(clientID).Login(account.Address(), account.BridgePassword()).AssertOK()
	return ctx.GetTestingError()
}

func thereIsSMTPClientLoggedInAsWithAddress(bddUserID, bddAddressID string) error {
	return thereIsSMTPClientNamedLoggedInAsWithAddress("smtp", bddUserID, bddAddressID)
}

func thereIsSMTPClientNamedLoggedInAsWithAddress(clientID, bddUserID, bddAddressID string) error {
	account := ctx.GetTestAccountWithAddress(bddUserID, bddAddressID)
	if account == nil {
		return godog.ErrPending
	}
	ctx.GetSMTPClient(clientID).Login(account.Address(), account.BridgePassword()).AssertOK()
	return ctx.GetTestingError()
}
