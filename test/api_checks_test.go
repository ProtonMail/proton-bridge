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
	"fmt"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/gherkin"
	"github.com/stretchr/testify/assert"
)

func APIChecksFeatureContext(s *godog.Suite) {
	s.Step(`^API endpoint "([^"]*)" is called with:$`, apiIsCalledWith)
	s.Step(`^message is sent with API call$`, messageIsSentWithAPICall)
	s.Step(`^API mailbox "([^"]*)" for "([^"]*)" has messages$`, apiMailboxForUserHasMessages)
	s.Step(`^API mailbox "([^"]*)" for address "([^"]*)" of "([^"]*)" has messages$`, apiMailboxForAddressOfUserHasMessages)
	s.Step(`^API client manager user-agent is "([^"]*)"$`, clientManagerUserAgent)
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
	endpoint := "POST /mail/v4/messages"
	if err := apiIsCalledWith(endpoint, data); err != nil {
		return err
	}
	for _, request := range ctx.GetPMAPIController().GetCalls("POST", "/mail/v4/messages") {
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

func apiMailboxForUserHasMessages(mailboxName, bddUserID string, messages *gherkin.DataTable) error {
	return apiMailboxForAddressOfUserHasMessages(mailboxName, "", bddUserID, messages)
}

func apiMailboxForAddressOfUserHasMessages(mailboxName, bddAddressID, bddUserID string, messages *gherkin.DataTable) error {
	account := ctx.GetTestAccountWithAddress(bddUserID, bddAddressID)
	if account == nil {
		return godog.ErrPending
	}

	labelIDs, err := ctx.GetPMAPIController().GetLabelIDs(account.Username(), []string{mailboxName})
	if err != nil {
		return internalError(err, "getting label %s for %s", mailboxName, account.Username())
	}
	labelID := labelIDs[0]

	pmapiMessages, err := ctx.GetPMAPIController().GetMessages(account.Username(), labelID)
	if err != nil {
		return err
	}

	head := messages.Rows[0].Cells
	for _, row := range messages.Rows[1:] {
		found, err := pmapiMessagesContainsMessageRow(account, pmapiMessages, head, row)
		if err != nil {
			return err
		}
		if !found {
			rowMap := map[string]string{}
			for idx, cell := range row.Cells {
				rowMap[head[idx].Value] = cell.Value
			}
			return fmt.Errorf("message %v not found", rowMap)
		}
	}
	return nil
}

func clientManagerUserAgent(expectedUserAgent string) error {
	expectedUserAgent = strings.ReplaceAll(expectedUserAgent, "[GOOS]", runtime.GOOS)

	assert.Eventually(ctx.GetTestingT(), func() bool {
		userAgent := ctx.GetClientManager().GetUserAgent()
		return userAgent == expectedUserAgent
	}, 5*time.Second, time.Second)

	return nil
}
