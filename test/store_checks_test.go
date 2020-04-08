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
	"net/mail"
	"strings"
	"time"

	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
	"github.com/ProtonMail/proton-bridge/test/accounts"
	"github.com/cucumber/godog"
	"github.com/cucumber/godog/gherkin"
)

func StoreChecksFeatureContext(s *godog.Suite) {
	s.Step(`^"([^"]*)" has mailbox "([^"]*)"$`, userHasMailbox)
	s.Step(`^"([^"]*)" does not have mailbox "([^"]*)"$`, userDoesNotHaveMailbox)
	s.Step(`^"([^"]*)" has the following messages$`, userHasFollowingMessages)
	s.Step(`^mailbox "([^"]*)" for "([^"]*)" has (\d+) message(?:s)?$`, mailboxForUserHasNumberOfMessages)
	s.Step(`^mailbox "([^"]*)" for "([^"]*)" has no messages$`, mailboxForUserHasNoMessages)
	s.Step(`^mailbox "([^"]*)" for address "([^"]*)" of "([^"]*)" has (\d+) message(?:s)?$`, mailboxForAddressOfUserHasNumberOfMessages)
	s.Step(`^mailbox "([^"]*)" for "([^"]*)" has messages$`, mailboxForUserHasMessages)
	s.Step(`^mailbox "([^"]*)" for address "([^"]*)" of "([^"]*)" has message(?:s)?$`, mailboxForAddressOfUserHasMessages)
	s.Step(`^message "([^"]*)" in "([^"]*)" for "([^"]*)" is marked as read$`, messagesInMailboxForUserIsMarkedAsRead)
	s.Step(`^message "([^"]*)" in "([^"]*)" for "([^"]*)" is marked as unread$`, messagesInMailboxForUserIsMarkedAsUnread)
	s.Step(`^message "([^"]*)" in "([^"]*)" for "([^"]*)" is marked as starred$`, messagesInMailboxForUserIsMarkedAsStarred)
	s.Step(`^message "([^"]*)" in "([^"]*)" for "([^"]*)" is marked as unstarred$`, messagesInMailboxForUserIsMarkedAsUnstarred)
}

func userHasMailbox(bddUserID, mailboxName string) error {
	account := ctx.GetTestAccount(bddUserID)
	if account == nil {
		return godog.ErrPending
	}
	if mailbox, err := ctx.GetStoreMailbox(account.Username(), account.AddressID(), mailboxName); mailbox == nil || mailbox.Name() != mailboxName {
		return fmt.Errorf("user %s does not have mailbox %s: %s", account.Username(), mailboxName, err)
	}
	return nil
}

func userDoesNotHaveMailbox(bddUserID, mailboxName string) error {
	account := ctx.GetTestAccount(bddUserID)
	if account == nil {
		return godog.ErrPending
	}
	if mailbox, err := ctx.GetStoreMailbox(account.Username(), account.AddressID(), mailboxName); mailbox != nil {
		return fmt.Errorf("user %s has unexpected mailbox %s: %s", account.Username(), mailboxName, err)
	}
	return nil
}

func userHasFollowingMessages(bddUserID string, structure *gherkin.DataTable) error {
	return processMailboxStructureDataTable(structure, func(bddAddressID string, mailboxNames string, numberOfMessages int) error {
		for _, mailboxName := range strings.Split(mailboxNames, ",") {
			if err := mailboxForAddressOfUserHasNumberOfMessages(mailboxName, bddAddressID, bddUserID, numberOfMessages); err != nil {
				return err
			}
		}
		return nil
	})
}

func mailboxForUserHasNumberOfMessages(mailboxName, bddUserID string, countOfMessages int) error {
	return mailboxForAddressOfUserHasNumberOfMessages(mailboxName, "", bddUserID, countOfMessages)
}

func mailboxForAddressOfUserHasNumberOfMessages(mailboxName, bddAddressID, bddUserID string, countOfMessages int) error {
	account := ctx.GetTestAccountWithAddress(bddUserID, bddAddressID)
	if account == nil {
		return godog.ErrPending
	}
	mailbox, err := ctx.GetStoreMailbox(account.Username(), account.AddressID(), mailboxName)
	if err != nil {
		return internalError(err, "getting store mailbox")
	}
	start := time.Now()
	for {
		afterLimit := time.Since(start) > ctx.EventLoopTimeout()
		total, _, _ := mailbox.GetCounts()
		if total == uint(countOfMessages) {
			break
		}
		if afterLimit {
			return fmt.Errorf("expected %v messages, but got %v", countOfMessages, total)
		}
		time.Sleep(100 * time.Millisecond)
	}
	return nil
}

func mailboxForUserHasMessages(mailboxName, bddUserID string, messages *gherkin.DataTable) error {
	return mailboxForAddressOfUserHasMessages(mailboxName, "", bddUserID, messages)
}

func mailboxForUserHasNoMessages(mailboxName, bddUserID string) error {
	return mailboxForUserHasNumberOfMessages(mailboxName, bddUserID, 0)
}

func mailboxForAddressOfUserHasMessages(mailboxName, bddAddressID, bddUserID string, messages *gherkin.DataTable) error {
	account := ctx.GetTestAccountWithAddress(bddUserID, bddAddressID)
	if account == nil {
		return godog.ErrPending
	}
	mailbox, err := ctx.GetStoreMailbox(account.Username(), account.AddressID(), mailboxName)
	if err != nil {
		return internalError(err, "getting store mailbox")
	}
	apiIDs, err := mailbox.GetAPIIDsFromSequenceRange(0, 1000)
	if err != nil {
		return internalError(err, "getting API IDs from sequence range")
	}
	allMessages := []*pmapi.Message{}
	for _, apiID := range apiIDs {
		message, err := mailbox.GetMessage(apiID)
		if err != nil {
			return internalError(err, "getting message by ID")
		}
		allMessages = append(allMessages, message.Message())
	}

	head := messages.Rows[0].Cells
	start := time.Now()
	for {
		afterLimit := time.Since(start) > ctx.EventLoopTimeout()
		allFound := true
		for _, row := range messages.Rows[1:] {
			found, err := messagesContainsMessageRow(account, allMessages, head, row)
			if err != nil {
				return err
			}
			if !found {
				if !afterLimit {
					allFound = false
					break
				}
				rowMap := map[string]string{}
				for idx, cell := range row.Cells {
					rowMap[head[idx].Value] = cell.Value
				}
				return fmt.Errorf("message %v not found", rowMap)
			}
		}
		if allFound {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	return nil
}

func messagesContainsMessageRow(account *accounts.TestAccount, allMessages []*pmapi.Message, head []*gherkin.TableCell, row *gherkin.TableRow) (bool, error) { //nolint[funlen]
	found := false
	for _, message := range allMessages {
		matches := true
		for n, cell := range row.Cells {
			switch head[n].Value {
			case "time":
				switch cell.Value {
				case "now":
					if (time.Now().Unix() - message.Time) > 5 {
						matches = false
					}
				default:
					return false, fmt.Errorf("unexpected time value: %s", cell.Value)
				}
			case "from":
				address := ctx.EnsureAddress(account.Username(), cell.Value)
				if !areAddressesSame(message.Sender.Address, address) {
					matches = false
				}
			case "to":
				for _, address := range strings.Split(cell.Value, ",") {
					address = ctx.EnsureAddress(account.Username(), address)
					for _, to := range message.ToList {
						if !areAddressesSame(to.Address, address) {
							matches = false
							break
						}
					}
				}
			case "cc":
				for _, address := range strings.Split(cell.Value, ",") {
					address = ctx.EnsureAddress(account.Username(), address)
					for _, to := range message.CCList {
						if !areAddressesSame(to.Address, address) {
							matches = false
							break
						}
					}
				}
			case "subject":
				expectedSubject := cell.Value
				if expectedSubject == "" {
					expectedSubject = "(No Subject)"
				}
				if message.Subject != expectedSubject {
					matches = false
				}
			case "body":
				if message.Body != cell.Value {
					matches = false
				}
			case "read":
				unread := 1
				if cell.Value == "true" { //nolint[goconst]
					unread = 0
				}
				if message.Unread != unread {
					matches = false
				}
			default:
				return false, fmt.Errorf("unexpected column name: %s", head[n].Value)
			}
		}
		if matches {
			found = true
			break
		}
	}
	return found, nil
}

func areAddressesSame(first, second string) bool {
	firstAddress, err := mail.ParseAddress(first)
	if err != nil {
		return false
	}
	secondAddress, err := mail.ParseAddress(second)
	if err != nil {
		return false
	}
	return firstAddress.String() == secondAddress.String()
}

func messagesInMailboxForUserIsMarkedAsRead(messageIDs, mailboxName, bddUserID string) error {
	return checkMessages(bddUserID, mailboxName, messageIDs, func(message *pmapi.Message) error {
		if message.Unread == 0 {
			return nil
		}
		return fmt.Errorf("message %s \"%s\" is expected to be read but is not", message.ID, message.Subject)
	})
}

func messagesInMailboxForUserIsMarkedAsUnread(messageIDs, mailboxName, bddUserID string) error {
	return checkMessages(bddUserID, mailboxName, messageIDs, func(message *pmapi.Message) error {
		if message.Unread == 1 {
			return nil
		}
		return fmt.Errorf("message %s \"%s\" is expected to not be read but is", message.ID, message.Subject)
	})
}

func messagesInMailboxForUserIsMarkedAsStarred(messageIDs, mailboxName, bddUserID string) error {
	return checkMessages(bddUserID, mailboxName, messageIDs, func(message *pmapi.Message) error {
		if hasItem(message.LabelIDs, "10") {
			return nil
		}
		return fmt.Errorf("message %s \"%s\" is expected to be starred but is not", message.ID, message.Subject)
	})
}

func messagesInMailboxForUserIsMarkedAsUnstarred(messageIDs, mailboxName, bddUserID string) error {
	return checkMessages(bddUserID, mailboxName, messageIDs, func(message *pmapi.Message) error {
		if !hasItem(message.LabelIDs, "10") {
			return nil
		}
		return fmt.Errorf("message %s \"%s\" is expected to not be starred but is", message.ID, message.Subject)
	})
}

func checkMessages(bddUserID, mailboxName, messageIDs string, callback func(*pmapi.Message) error) error {
	account := ctx.GetTestAccount(bddUserID)
	if account == nil {
		return godog.ErrPending
	}
	messages, err := getMessages(account.Username(), account.AddressID(), mailboxName, messageIDs)
	if err != nil {
		return internalError(err, "getting messages %s", messageIDs)
	}
	for _, message := range messages {
		if err := callback(message); err != nil {
			return err
		}
	}
	return nil
}

func getMessages(username, addressID, mailboxName, messageIDs string) ([]*pmapi.Message, error) {
	msgs := []*pmapi.Message{}
	var msg *pmapi.Message
	var err error
	iterateOverSeqSet(messageIDs, func(messageID string) {
		messageID = ctx.GetPMAPIController().GetMessageID(username, messageID)
		msg, err = getMessage(username, addressID, mailboxName, messageID)
		if err == nil {
			msgs = append(msgs, msg)
		}
	})
	return msgs, err
}

func getMessage(username, addressID, mailboxName, messageID string) (*pmapi.Message, error) {
	mailbox, err := ctx.GetStoreMailbox(username, addressID, mailboxName)
	if err != nil {
		return nil, err
	}
	message, err := mailbox.GetMessage(messageID)
	if err != nil {
		return nil, err
	}
	return message.Message(), nil
}

func hasItem(items []string, value string) bool {
	for _, item := range items {
		if item == value {
			return true
		}
	}
	return false
}
