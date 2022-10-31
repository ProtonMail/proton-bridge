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
	"time"

	"github.com/cucumber/godog"
)

func IMAPActionsMailboxFeatureContext(s *godog.ScenarioContext) {
	s.Step(`^IMAP client creates mailbox "([^"]*)"$`, imapClientCreatesMailbox)
	s.Step(`^IMAP client renames mailbox "([^"]*)" to "([^"]*)"$`, imapClientRenamesMailboxTo)
	s.Step(`^IMAP client deletes mailbox "([^"]*)"$`, imapClientDeletesMailbox)
	s.Step(`^IMAP client lists mailboxes$`, imapClientListsMailboxes)
	s.Step(`^IMAP client "([^"]*)" lists mailboxes$`, imapClientNamedListsMailboxes)
	s.Step(`^IMAP client selects "([^"]*)"$`, imapClientSelects)
	s.Step(`^IMAP client gets info of "([^"]*)"$`, imapClientGetsInfoOf)
	s.Step(`^IMAP client "([^"]*)" gets info of "([^"]*)"$`, imapClientNamedGetsInfoOf)
	s.Step(`^IMAP client gets status of "([^"]*)"$`, imapClientGetsStatusOf)
}

func imapClientCreatesMailbox(mailboxName string) error {
	res := ctx.GetIMAPClient("imap").CreateMailbox(mailboxName)
	ctx.SetIMAPLastResponse("imap", res)
	return nil
}

func imapClientRenamesMailboxTo(mailboxName, newMailboxName string) error {
	res := ctx.GetIMAPClient("imap").RenameMailbox(mailboxName, newMailboxName)
	ctx.SetIMAPLastResponse("imap", res)
	return nil
}

func imapClientDeletesMailbox(mailboxName string) error {
	if mailboxName == "Trash" {
		// Delete of Trash mailbox calls empty label on API.
		// Empty label means delete all messages in that label with time
		// creation before time of execution. But creation time is in
		// seconds, not miliseconds. That's why message created at the
		// same second as emptying label is called is not deleted.
		// Tests might be that fast and therefore we need to sleep for
		// a second to make sure test doesn't produce fake failure.
		time.Sleep(time.Second)
	}
	res := ctx.GetIMAPClient("imap").DeleteMailbox(mailboxName)
	ctx.SetIMAPLastResponse("imap", res)
	return nil
}

func imapClientListsMailboxes() error {
	return imapClientNamedListsMailboxes("imap")
}

func imapClientNamedListsMailboxes(clientName string) error {
	res := ctx.GetIMAPClient(clientName).ListMailboxes()
	ctx.SetIMAPLastResponse(clientName, res)
	return nil
}

func imapClientSelects(mailboxName string) error {
	res := ctx.GetIMAPClient("imap").Select(mailboxName)
	ctx.SetIMAPLastResponse("imap", res)
	return nil
}

func imapClientGetsInfoOf(mailboxName string) error {
	return imapClientNamedGetsInfoOf("imap", mailboxName)
}

func imapClientNamedGetsInfoOf(clientName, mailboxName string) error {
	res := ctx.GetIMAPClient(clientName).GetMailboxInfo(mailboxName)
	ctx.SetIMAPLastResponse(clientName, res)
	return nil
}

func imapClientGetsStatusOf(mailboxName string) error {
	res := ctx.GetIMAPClient("imap").GetMailboxStatus(mailboxName)
	ctx.SetIMAPLastResponse("imap", res)
	return nil
}
