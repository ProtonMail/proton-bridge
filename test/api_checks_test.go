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
	"fmt"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
	"github.com/ProtonMail/proton-bridge/v2/test/accounts"
	"github.com/cucumber/godog"
	"github.com/stretchr/testify/assert"
)

func APIChecksFeatureContext(s *godog.ScenarioContext) {
	s.Step(`^API endpoint "([^"]*)" is called$`, apiIsCalled)
	s.Step(`^API endpoint "([^"]*)" is called with$`, apiIsCalledWith)
	s.Step(`^API endpoint "([^"]*)" is not called$`, apiIsNotCalled)
	s.Step(`^API endpoint "([^"]*)" is not called with$`, apiIsNotCalledWith)
	s.Step(`^message is sent with API call$`, messageIsSentWithAPICall)
	s.Step(`^packages are sent with API call$`, packagesAreSentWithAPICall)
	s.Step(`^API mailbox "([^"]*)" for "([^"]*)" has (\d+) message(?:s)?$`, apiMailboxForUserHasNumberOfMessages)
	s.Step(`^API mailbox "([^"]*)" for address "([^"]*)" of "([^"]*)" has (\d+) message(?:s)?$`, apiMailboxForAddressOfUserHasNumberOfMessages)
	s.Step(`^API mailbox "([^"]*)" for "([^"]*)" has messages$`, apiMailboxForUserHasMessages)
	s.Step(`^API mailbox "([^"]*)" for address "([^"]*)" of "([^"]*)" has messages$`, apiMailboxForAddressOfUserHasMessages)
	s.Step(`^API user-agent is "([^"]*)"$`, userAgent)
}

func apiIsCalled(endpoint string) error {
	if !apiIsCalledWithHelper(endpoint, "") {
		return fmt.Errorf("%s was not called", endpoint)
	}
	return nil
}

func apiIsCalledWith(endpoint string, data *godog.DocString) error {
	if !apiIsCalledWithHelper(endpoint, data.Content) {
		return fmt.Errorf("%s was not called with %s", endpoint, data.Content)
	}
	return nil
}

func apiIsCalledWithRegex(endpoint string, data *godog.DocString) error {
	match, err := apiIsCalledWithHelperRegex(endpoint, data.Content)
	if err != nil {
		return err
	}
	if !match {
		return fmt.Errorf("%s was not called with %s", endpoint, data.Content)
	}
	return nil
}

func apiIsNotCalled(endpoint string) error {
	if apiIsCalledWithHelper(endpoint, "") {
		return fmt.Errorf("%s was called", endpoint)
	}
	return nil
}

func apiIsNotCalledWith(endpoint string, data *godog.DocString) error {
	if apiIsCalledWithHelper(endpoint, data.Content) {
		return fmt.Errorf("%s was called with %s", endpoint, data.Content)
	}
	return nil
}

func apiIsCalledWithHelper(endpoint string, content string) bool {
	split := strings.Split(endpoint, " ")
	method := split[0]
	path := split[1]
	request := []byte(content)
	return ctx.GetPMAPIController().WasCalled(method, path, request)
}

func apiIsCalledWithHelperRegex(endpoint string, content string) (bool, error) {
	split := strings.Split(endpoint, " ")
	method := split[0]
	path := split[1]
	request := []byte(content)
	return ctx.GetPMAPIController().WasCalledRegex(method, path, request)
}

func messageIsSentWithAPICall(data *godog.DocString) error {
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

func packagesAreSentWithAPICall(data *godog.DocString) error {
	endpoint := "POST /mail/v4/messages/.+$"
	if err := apiIsCalledWithRegex(endpoint, data); err != nil {
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

func apiMailboxForUserHasNumberOfMessages(mailboxName, bddUserID string, countOfMessages int) error {
	return apiMailboxForAddressOfUserHasNumberOfMessages(mailboxName, "", bddUserID, countOfMessages)
}

func apiMailboxForAddressOfUserHasNumberOfMessages(mailboxName, bddAddressID, bddUserID string, countOfMessages int) error {
	account := ctx.GetTestAccountWithAddress(bddUserID, bddAddressID)
	if account == nil {
		return godog.ErrPending
	}

	start := time.Now()
	for {
		afterLimit := time.Since(start) > ctx.EventLoopTimeout()
		pmapiMessages, err := getPMAPIMessages(account, mailboxName)
		if err != nil {
			return err
		}
		total := len(pmapiMessages)
		if total == countOfMessages {
			break
		}
		if afterLimit {
			return fmt.Errorf("expected %v messages, but got %v", countOfMessages, total)
		}
		time.Sleep(100 * time.Millisecond)
	}
	return nil
}

func apiMailboxForUserHasMessages(mailboxName, bddUserID string, messages *godog.Table) error {
	return apiMailboxForAddressOfUserHasMessages(mailboxName, "", bddUserID, messages)
}

func apiMailboxForAddressOfUserHasMessages(mailboxName, bddAddressID, bddUserID string, messages *godog.Table) error {
	account := ctx.GetTestAccountWithAddress(bddUserID, bddAddressID)
	if account == nil {
		return godog.ErrPending
	}

	pmapiMessages, err := getPMAPIMessages(account, mailboxName)
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

func getPMAPIMessages(account *accounts.TestAccount, mailboxName string) ([]*pmapi.Message, error) {
	labelIDs, err := ctx.GetPMAPIController().GetLabelIDs(account.Username(), []string{mailboxName})
	if err != nil {
		return nil, internalError(err, "getting label %s for %s", mailboxName, account.Username())
	}
	labelID := labelIDs[0]

	return ctx.GetPMAPIController().GetMessages(account.Username(), labelID)
}

func userAgent(expectedUserAgent string) error {
	expectedUserAgent = strings.ReplaceAll(expectedUserAgent, "[GOOS]", runtime.GOOS)

	assert.Eventually(ctx.GetTestingT(), func() bool {
		return ctx.GetUserAgent() == expectedUserAgent
	}, 5*time.Second, time.Second)

	return nil
}
