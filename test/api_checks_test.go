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
	"fmt"
	"regexp"
	"strings"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/gherkin"
)

func APIChecksFeatureContext(s *godog.Suite) {
	s.Step(`^API endpoint "([^"]*)" is called with:$`, apiIsCalledWith)
	s.Step(`^message is sent with API call:$`, messageIsSentWithAPICall)
}

func apiIsCalledWith(endpoint string, data *gherkin.DocString) error {
	split := strings.Split(endpoint, " ")
	method := split[0]
	path := split[1]
	request := []byte(data.Content)
	if !ctx.GetPMAPIController().WasCalled(method, path, request) {
		return fmt.Errorf("%s was not called with %s", endpoint, request)
	}
	return nil
}

func messageIsSentWithAPICall(data *gherkin.DocString) error {
	endpoint := "POST /messages"
	if err := apiIsCalledWith(endpoint, data); err != nil {
		return err
	}
	for _, request := range ctx.GetPMAPIController().GetCalls("POST", "/messages") {
		if !checkAllRequiredFieldsForSendingMessage(request) {
			return fmt.Errorf("%s was not called with all required fields: %s", endpoint, request)
		}
	}

	return nil
}

func checkAllRequiredFieldsForSendingMessage(request []byte) bool {
	if matches := regexp.MustCompile(`"Subject":`).Match(request); !matches {
		return false
	}
	if matches := regexp.MustCompile(`"ToList":`).Match(request); !matches {
		return false
	}
	if matches := regexp.MustCompile(`"CCList":`).Match(request); !matches {
		return false
	}
	if matches := regexp.MustCompile(`"BCCList":`).Match(request); !matches {
		return false
	}
	if matches := regexp.MustCompile(`"AddressID":`).Match(request); !matches {
		return false
	}
	if matches := regexp.MustCompile(`"Body":`).Match(request); !matches {
		return false
	}
	return true
}
