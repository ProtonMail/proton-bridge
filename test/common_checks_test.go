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
	a "github.com/stretchr/testify/assert"
)

func CommonChecksFeatureContext(s *godog.ScenarioContext) {
	s.Step(`^last response is "([^"]*)"$`, lastResponseIs)
}

func lastResponseIs(expectedResponse string) error {
	err := ctx.GetLastError()
	if expectedResponse == "OK" {
		a.NoError(ctx.GetTestingT(), err)
	} else {
		a.EqualError(ctx.GetTestingT(), err, expectedResponse)
	}
	return ctx.GetTestingError()
}
