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

func SMTPChecksFeatureContext(s *godog.ScenarioContext) {
	s.Step(`^SMTP response is "([^"]*)"$`, smtpResponseIs)
	s.Step(`^SMTP response to "([^"]*)" is "([^"]*)"$`, smtpResponseNamedIs)
	s.Step(`^SMTP client is logged out`, smtpClientIsLoggedOut)
	s.Step(`^SMTP client "([^"]*)" is logged out`, smtpClientNamedIsLoggedOut)
}

func smtpResponseIs(expectedResponse string) error {
	return smtpResponseNamedIs("smtp", expectedResponse)
}

func smtpResponseNamedIs(clientID, expectedResponse string) error {
	res := ctx.GetSMTPLastResponse(clientID)
	if expectedResponse == "OK" {
		res.AssertOK()
	} else {
		res.AssertError(expectedResponse)
	}
	return ctx.GetTestingError()
}

func smtpClientIsLoggedOut() error {
	return smtpClientNamedIsLoggedOut("smtp")
}

func smtpClientNamedIsLoggedOut(clientName string) error {
	res := ctx.GetSMTPClient(clientName).SendCommands("HELO loggedOut.com")
	res.AssertError("read response failed:")
	return ctx.GetTestingError()
}
