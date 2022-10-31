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
	"strings"
	"time"

	"github.com/ProtonMail/go-rfc5322"
	"github.com/ProtonMail/proton-bridge/v2/internal/store"
	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
	"github.com/ProtonMail/proton-bridge/v2/test/accounts"
	"github.com/cucumber/godog"
	"github.com/cucumber/messages-go/v16"
	"github.com/hashicorp/go-multierror"
)

func StoreChecksFeatureContext(s *godog.ScenarioContext) {
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
	s.Step(`^message "([^"]*)" in "([^"]*)" for "([^"]*)" is marked as deleted$`, messagesInMailboxForUserIsMarkedAsDeleted)
	s.Step(`^message "([^"]*)" in "([^"]*)" for "([^"]*)" is marked as undeleted$`, messagesInMailboxForUserIsMarkedAsUndeleted)
	s.Step(`^header is not cached for message "([^"]*)" in "([^"]*)" for "([^"]*)"$`, messageHeaderIsNotCached)
	s.Step(`^header is cached for message "([^"]*)" in "([^"]*)" for "([^"]*)"$`, messageHeaderIsCached)
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

func userHasFollowingMessages(bddUserID string, structure *godog.Table) error {
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
		total, _, _, _ := mailbox.GetCounts()
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

func mailboxForUserHasMessages(mailboxName, bddUserID string, messages *godog.Table) error {
	return mailboxForAddressOfUserHasMessages(mailboxName, "", bddUserID, messages)
}

func mailboxForUserHasNoMessages(mailboxName, bddUserID string) error {
	return mailboxForUserHasNumberOfMessages(mailboxName, bddUserID, 0)
}

func mailboxForAddressOfUserHasMessages(mailboxName, bddAddressID, bddUserID string, messages *godog.Table) error {
	account := ctx.GetTestAccountWithAddress(bddUserID, bddAddressID)
	if account == nil {
		return godog.ErrPending
	}
	mailbox, err := ctx.GetStoreMailbox(account.Username(), account.AddressID(), mailboxName)
	if err != nil {
		return internalError(err, "getting store mailbox")
	}
	apiIDs, err := mailbox.GetAPIIDsFromSequenceRange(1, 1000)
	if err != nil {
		return internalError(err, "getting API IDs from sequence range")
	}
	allMessages := []*store.Message{}
	for _, apiID := range apiIDs {
		message, err := mailbox.GetMessage(apiID)
		if err != nil {
			return internalError(err, "getting message by ID")
		}
		allMessages = append(allMessages, message)
	}

	head := messages.Rows[0].Cells
	start := time.Now()
	for {
		afterLimit := time.Since(start) > ctx.EventLoopTimeout()
		allFound := true
		for _, row := range messages.Rows[1:] {
			found, err := storeMessagesContainsMessageRow(account, allMessages, head, row)
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

func pmapiMessagesContainsMessageRow(account *accounts.TestAccount, pmapiMessages []*pmapi.Message, head []*messages.PickleTableCell, row *messages.PickleTableRow) (bool, error) {
	messages := make([]interface{}, len(pmapiMessages))
	for i := range pmapiMessages {
		messages[i] = pmapiMessages[i]
	}
	return messagesContainsMessageRow(account, messages, head, row)
}

func storeMessagesContainsMessageRow(account *accounts.TestAccount, storeMessages []*store.Message, head []*messages.PickleTableCell, row *messages.PickleTableRow) (bool, error) {
	messages := make([]interface{}, len(storeMessages))
	for i := range storeMessages {
		messages[i] = storeMessages[i]
	}
	return messagesContainsMessageRow(account, messages, head, row)
}

func messagesContainsMessageRow(account *accounts.TestAccount, allMessages []interface{}, head []*messages.PickleTableCell, row *messages.PickleTableRow) (bool, error) { //nolint:funlen,gocyclo
	found := false
	for _, someMessage := range allMessages {
		var message *pmapi.Message
		var storeMessage *store.Message

		switch v := someMessage.(type) {
		case *pmapi.Message:
			message = v
		case *store.Message:
			message = v.Message()
			storeMessage = v
		}

		matches := true
		for n, cell := range row.Cells {
			switch head[n].Value {
			case "id":
				id, err := ctx.GetAPIMessageID(account.Username(), cell.Value)
				if err != nil {
					return false, fmt.Errorf("unknown BDD message ID: %s", cell.Value)
				}
				if message.ID != id {
					matches = false
				}
			case "externalid":
				if message.ExternalID != cell.Value {
					matches = false
				}
			case "from": //nolint:goconst
				address := ctx.EnsureAddress(account.Username(), cell.Value)
				if !areAddressesSame(message.Sender.Address, address) {
					matches = false
				}
			case "to": //nolint:goconst
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
			case "subject": //nolint:goconst
				expectedSubject := cell.Value
				if expectedSubject == "" {
					expectedSubject = "(No Subject)"
				}
				if message.Subject != expectedSubject {
					matches = false
				}
			case "body": //nolint:goconst
				if message.Body != cell.Value {
					matches = false
				}
			case "read":
				var unread pmapi.Boolean
				if cell.Value == "true" { //nolint:goconst
					unread = false
				} else {
					unread = true
				}

				if message.Unread != unread {
					matches = false
				}
			case "deleted": //nolint:goconst
				if storeMessage == nil {
					return false, fmt.Errorf("deleted column not supported for pmapi message object")
				}

				expectedDeleted := cell.Value == "true"
				matches = storeMessage.IsMarkedDeleted() == expectedDeleted
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
	firstAddress, err := rfc5322.ParseAddressList(first)
	if err != nil {
		return false
	}
	secondAddress, err := rfc5322.ParseAddressList(second)
	if err != nil {
		return false
	}
	return firstAddress[0].Address == secondAddress[0].Address
}

func messagesInMailboxForUserIsMarkedAsRead(bddMessageIDs, mailboxName, bddUserID string) error {
	return checkMessages(bddUserID, mailboxName, bddMessageIDs, func(message *store.Message) error {
		if !message.Message().Unread {
			return nil
		}
		return fmt.Errorf("message %s \"%s\" is expected to be read but is not", message.ID(), message.Message().Subject)
	})
}

func messagesInMailboxForUserIsMarkedAsUnread(bddMessageIDs, mailboxName, bddUserID string) error {
	return checkMessages(bddUserID, mailboxName, bddMessageIDs, func(message *store.Message) error {
		if message.Message().Unread {
			return nil
		}
		return fmt.Errorf("message %s \"%s\" is expected to not be read but is", message.ID(), message.Message().Subject)
	})
}

func messagesInMailboxForUserIsMarkedAsStarred(bddMessageIDs, mailboxName, bddUserID string) error {
	return checkMessages(bddUserID, mailboxName, bddMessageIDs, func(message *store.Message) error {
		if hasItem(message.Message().LabelIDs, "10") {
			return nil
		}
		return fmt.Errorf("message %s \"%s\" is expected to be starred but is not", message.ID(), message.Message().Subject)
	})
}

func messagesInMailboxForUserIsMarkedAsUnstarred(bddMessageIDs, mailboxName, bddUserID string) error {
	return checkMessages(bddUserID, mailboxName, bddMessageIDs, func(message *store.Message) error {
		if !hasItem(message.Message().LabelIDs, "10") {
			return nil
		}
		return fmt.Errorf("message %s \"%s\" is expected to not be starred but is", message.ID(), message.Message().Subject)
	})
}

func messagesInMailboxForUserIsMarkedAsDeleted(bddMessageIDs, mailboxName, bddUserID string) error {
	return checkMessages(bddUserID, mailboxName, bddMessageIDs, func(message *store.Message) error {
		if message.IsMarkedDeleted() {
			return nil
		}
		return fmt.Errorf("message %s \"%s\" is expected to be deleted but is not", message.ID(), message.Message().Subject)
	})
}

func messagesInMailboxForUserIsMarkedAsUndeleted(bddMessageIDs, mailboxName, bddUserID string) error {
	return checkMessages(bddUserID, mailboxName, bddMessageIDs, func(message *store.Message) error {
		if !message.IsMarkedDeleted() {
			return nil
		}
		return fmt.Errorf("message %s \"%s\" is expected to not be deleted but is", message.ID(), message.Message().Subject)
	})
}

func messageHeaderIsNotCached(bddMessageIDs, mailboxName, bddUserID string) error {
	return checkMessages(bddUserID, mailboxName, bddMessageIDs, func(message *store.Message) error {
		if !message.IsFullHeaderCached() {
			return nil
		}
		return fmt.Errorf("message %s \"%s\" is expected to not have cached header but has one", message.ID(), message.Message().Subject)
	})
}

func messageHeaderIsCached(bddMessageIDs, mailboxName, bddUserID string) error {
	return checkMessages(bddUserID, mailboxName, bddMessageIDs, func(message *store.Message) error {
		if message.IsFullHeaderCached() {
			return nil
		}
		return fmt.Errorf("message %s \"%s\" is expected to have cached header but it doesn't have", message.ID(), message.Message().Subject)
	})
}

func checkMessages(bddUserID, mailboxName, bddMessageIDs string, callback func(*store.Message) error) error {
	account := ctx.GetTestAccount(bddUserID)
	if account == nil {
		return godog.ErrPending
	}
	messages, err := getMessages(account.Username(), account.AddressID(), mailboxName, bddMessageIDs)
	if err != nil {
		return internalError(err, "getting messages %s", bddMessageIDs)
	}
	for _, message := range messages {
		if err := callback(message); err != nil {
			return err
		}
	}
	return nil
}

func getMessages(username, addressID, mailboxName, bddMessageIDs string) ([]*store.Message, error) {
	msgs := []*store.Message{}
	var allErrs *multierror.Error
	iterateOverSeqSet(bddMessageIDs, func(bddMessageID string) {
		messageID, err := ctx.GetAPIMessageID(username, bddMessageID)
		if err != nil {
			allErrs = multierror.Append(allErrs, err)
			return
		}
		msg, err := getMessage(username, addressID, mailboxName, messageID)
		if err != nil {
			allErrs = multierror.Append(allErrs, err)
			return
		}
		msgs = append(msgs, msg)
	})
	return msgs, allErrs.ErrorOrNil()
}

func getMessage(username, addressID, mailboxName, messageID string) (*store.Message, error) {
	mailbox, err := ctx.GetStoreMailbox(username, addressID, mailboxName)
	if err != nil {
		return nil, err
	}
	return mailbox.GetMessage(messageID)
}

func hasItem(items []string, value string) bool {
	for _, item := range items {
		if item == value {
			return true
		}
	}
	return false
}
