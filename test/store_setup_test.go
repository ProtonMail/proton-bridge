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
	"strconv"
	"strings"
	"time"

	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
	"github.com/cucumber/godog"
	"github.com/cucumber/godog/gherkin"
)

func StoreSetupFeatureContext(s *godog.Suite) {
	s.Step(`^there is "([^"]*)" with mailboxes`, thereIsUserWithMailboxes)
	s.Step(`^there is "([^"]*)" with mailbox "([^"]*)"$`, thereIsUserWithMailbox)
	s.Step(`^there are messages in mailbox(?:es)? "([^"]*)" for "([^"]*)"$`, thereAreMessagesInMailboxesForUser)
	s.Step(`^there are messages in mailbox(?:es)? "([^"]*)" for address "([^"]*)" of "([^"]*)"$`, thereAreMessagesInMailboxesForAddressOfUser)
	s.Step(`^there are (\d+) messages in mailbox(?:es)? "([^"]*)" for "([^"]*)"$`, thereAreSomeMessagesInMailboxesForUser)
	s.Step(`^there are messages for "([^"]*)" as follows$`, thereAreSomeMessagesForUserAsFollows)
	s.Step(`^there are (\d+) messages in mailbox(?:es)? "([^"]*)" for address "([^"]*)" of "([^"]*)"$`, thereAreSomeMessagesInMailboxesForAddressOfUser)
}

func thereIsUserWithMailboxes(bddUserID string, mailboxes *gherkin.DataTable) error {
	for _, row := range mailboxes.Rows {
		if err := thereIsUserWithMailbox(bddUserID, row.Cells[0].Value); err != nil {
			return err
		}
	}
	return nil
}

func thereIsUserWithMailbox(bddUserID, mailboxName string) error {
	account := ctx.GetTestAccount(bddUserID)
	if account == nil {
		return godog.ErrPending
	}
	err := ctx.GetPMAPIController().AddUserLabel(account.Username(), &pmapi.Label{
		Name: mailboxName,
		Type: pmapi.LabelTypeMailbox,
	})
	if err != nil {
		return internalError(err, "adding label %s for %s", mailboxName, account.Username())
	}
	store, err := ctx.GetStore(account.Username())
	if err != nil {
		return internalError(err, "getting store of %s", account.Username())
	}
	if store == nil {
		return nil
	}
	return internalError(store.RebuildMailboxes(), "rebuilding mailboxes")
}

func thereAreMessagesInMailboxesForUser(mailboxNames, bddUserID string, messages *gherkin.DataTable) error {
	return thereAreMessagesInMailboxesForAddressOfUser(mailboxNames, "", bddUserID, messages)
}

func thereAreMessagesInMailboxesForAddressOfUser(mailboxNames, bddAddressID, bddUserID string, messages *gherkin.DataTable) error {
	account := ctx.GetTestAccountWithAddress(bddUserID, bddAddressID)
	if account == nil {
		return godog.ErrPending
	}

	labelIDs, err := ctx.GetPMAPIController().GetLabelIDs(account.Username(), strings.Split(mailboxNames, ","))
	if err != nil {
		return internalError(err, "getting labels %s for %s", mailboxNames, account.Username())
	}

	var markMessageIDsDeleted []string

	// Inserting in the opposite order becase sync is done from newest to oldest.
	// The goal is to have simply predictable IMAP sequence numbers if possible.
	head := messages.Rows[0].Cells
	for i := len(messages.Rows) - 1; i > 0; i-- {
		row := messages.Rows[i]
		message := &pmapi.Message{
			MIMEType:  "text/plain",
			LabelIDs:  labelIDs,
			AddressID: account.AddressID(),
		}

		if message.HasLabelID(pmapi.SentLabel) {
			message.Flags |= pmapi.FlagSent
		}

		bddMessageID := ""
		hasDeletedFlag := false

		for n, cell := range row.Cells {
			switch head[n].Value {
			case "id":
				bddMessageID = cell.Value
			case "from":
				message.Sender = &mail.Address{
					Address: ctx.EnsureAddress(account.Username(), cell.Value),
				}
			case "to":
				message.AddressID = ctx.EnsureAddressID(account.Username(), cell.Value)
				message.ToList = []*mail.Address{{
					Address: ctx.EnsureAddress(account.Username(), cell.Value),
				}}
			case "cc":
				message.AddressID = ctx.EnsureAddressID(account.Username(), cell.Value)
				message.CCList = []*mail.Address{{
					Address: ctx.EnsureAddress(account.Username(), cell.Value),
				}}
			case "subject":
				message.Subject = cell.Value
			case "body":
				message.Body = cell.Value
			case "read":
				unread := 1
				if cell.Value == "true" {
					unread = 0
				}
				message.Unread = unread
			case "starred":
				if cell.Value == "true" {
					message.LabelIDs = append(message.LabelIDs, "10")
				}
			case "time": //nolint[goconst]
				date, err := time.Parse(timeFormat, cell.Value)
				if err != nil {
					return internalError(err, "parsing time")
				}
				message.Time = date.Unix()
			case "deleted":
				hasDeletedFlag = cell.Value == "true"
			default:
				return fmt.Errorf("unexpected column name: %s", head[n].Value)
			}
		}
		lastMessageID, err := ctx.GetPMAPIController().AddUserMessage(account.Username(), message)
		if err != nil {
			return internalError(err, "adding message")
		}
		ctx.PairMessageID(account.Username(), bddMessageID, lastMessageID)

		if hasDeletedFlag {
			markMessageIDsDeleted = append(markMessageIDsDeleted, lastMessageID)
		}
	}

	if err := internalError(ctx.WaitForSync(account.Username()), "waiting for sync"); err != nil {
		return err
	}

	if len(markMessageIDsDeleted) > 0 {
		for _, mailboxName := range strings.Split(mailboxNames, ",") {
			storeMailbox, err := ctx.GetStoreMailbox(account.Username(), account.AddressID(), mailboxName)
			if err != nil {
				return err
			}
			if err := storeMailbox.MarkMessagesDeleted(markMessageIDsDeleted); err != nil {
				return err
			}
		}
	}

	return nil
}

func thereAreSomeMessagesInMailboxesForUser(numberOfMessages int, mailboxNames, bddUserID string) error {
	return thereAreSomeMessagesInMailboxesForAddressOfUser(numberOfMessages, mailboxNames, "", bddUserID)
}

func thereAreSomeMessagesForUserAsFollows(bddUserID string, structure *gherkin.DataTable) error {
	return processMailboxStructureDataTable(structure, func(bddAddressID string, mailboxNames string, numberOfMessages int) error {
		return thereAreSomeMessagesInMailboxesForAddressOfUser(numberOfMessages, mailboxNames, bddAddressID, bddUserID)
	})
}

func processMailboxStructureDataTable(structure *gherkin.DataTable, callback func(string, string, int) error) error {
	head := structure.Rows[0].Cells
	for i := 1; i < len(structure.Rows); i++ {
		bddAddressID := ""
		mailboxNames := "INBOX"
		numberOfMessages := 0
		for n, cell := range structure.Rows[i].Cells {
			switch head[n].Value {
			case "address":
				bddAddressID = cell.Value
			case "mailboxes":
				mailboxNames = cell.Value
			case "messages":
				number, err := strconv.Atoi(cell.Value)
				if err != nil {
					return internalError(err, "converting number of messages to integer")
				}
				numberOfMessages = number
			default:
				return fmt.Errorf("unexpected column name: %s", head[n].Value)
			}
		}

		if err := callback(bddAddressID, mailboxNames, numberOfMessages); err != nil {
			return err
		}
	}
	return nil
}

func thereAreSomeMessagesInMailboxesForAddressOfUser(numberOfMessages int, mailboxNames, bddAddressID, bddUserID string) error {
	account := ctx.GetTestAccountWithAddress(bddUserID, bddAddressID)
	if account == nil {
		return godog.ErrPending
	}

	for i := 1; i <= numberOfMessages; i++ {
		labelIDs, err := ctx.GetPMAPIController().GetLabelIDs(account.Username(), strings.Split(mailboxNames, ","))
		if err != nil {
			return internalError(err, "getting labels %s for %s", mailboxNames, account.Username())
		}
		lastMessageID, err := ctx.GetPMAPIController().AddUserMessage(account.Username(), &pmapi.Message{
			MIMEType:  "text/plain",
			LabelIDs:  labelIDs,
			AddressID: account.AddressID(),
			Subject:   fmt.Sprintf("Test message #%d", i),
			Sender:    &mail.Address{Address: "anyone@example.com"},
			ToList:    []*mail.Address{{Address: account.Address()}},
		})
		if err != nil {
			return internalError(err, "adding message")
		}

		// Generating IDs in the opposite order becase sync is done from newest to oldest.
		// The goal is to have simply predictable IMAP sequence numbers if possible.
		bddMessageID := fmt.Sprintf("%d", numberOfMessages-i+1)
		ctx.PairMessageID(account.Username(), bddMessageID, lastMessageID)
	}
	return internalError(ctx.WaitForSync(account.Username()), "waiting for sync")
}
