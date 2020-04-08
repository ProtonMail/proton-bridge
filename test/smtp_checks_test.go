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
	"github.com/cucumber/godog"
)

func SMTPChecksFeatureContext(s *godog.Suite) {
	s.Step(`^SMTP response is "([^"]*)"$`, smtpResponseIs)
	s.Step(`^SMTP response to "([^"]*)" is "([^"]*)"$`, smtpResponseNamedIs)
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
