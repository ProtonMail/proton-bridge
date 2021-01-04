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
	"github.com/cucumber/godog"
)

func IMAPActionsMailboxFeatureContext(s *godog.Suite) {
	s.Step(`^IMAP client creates mailbox "([^"]*)"$`, imapClientCreatesMailbox)
	s.Step(`^IMAP client renames mailbox "([^"]*)" to "([^"]*)"$`, imapClientRenamesMailboxTo)
	s.Step(`^IMAP client deletes mailbox "([^"]*)"$`, imapClientDeletesMailbox)
	s.Step(`^IMAP client lists mailboxes$`, imapClientListsMailboxes)
	s.Step(`^IMAP client selects "([^"]*)"$`, imapClientSelects)
	s.Step(`^IMAP client gets info of "([^"]*)"$`, imapClientGetsInfoOf)
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
	res := ctx.GetIMAPClient("imap").DeleteMailbox(mailboxName)
	ctx.SetIMAPLastResponse("imap", res)
	return nil
}

func imapClientListsMailboxes() error {
	res := ctx.GetIMAPClient("imap").ListMailboxes()
	ctx.SetIMAPLastResponse("imap", res)
	return nil
}

func imapClientSelects(mailboxName string) error {
	res := ctx.GetIMAPClient("imap").Select(mailboxName)
	ctx.SetIMAPLastResponse("imap", res)
	return nil
}

func imapClientGetsInfoOf(mailboxName string) error {
	res := ctx.GetIMAPClient("imap").GetMailboxInfo(mailboxName)
	ctx.SetIMAPLastResponse("imap", res)
	return nil
}

func imapClientGetsStatusOf(mailboxName string) error {
	res := ctx.GetIMAPClient("imap").GetMailboxStatus(mailboxName)
	ctx.SetIMAPLastResponse("imap", res)
	return nil
}
