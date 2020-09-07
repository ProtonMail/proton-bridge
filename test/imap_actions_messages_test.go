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
	"strconv"
	"sync"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/gherkin"
	"golang.org/x/net/html/charset"
)

func IMAPActionsMessagesFeatureContext(s *godog.Suite) {
	s.Step(`^IMAP client sends command "([^"]*)"$`, imapClientSendsCommand)
	s.Step(`^IMAP client fetches "([^"]*)"$`, imapClientFetches)
	s.Step(`^IMAP client fetches by UID "([^"]*)"$`, imapClientFetchesByUID)
	s.Step(`^IMAP client searches for "([^"]*)"$`, imapClientSearchesFor)
	s.Step(`^IMAP client copies messages "([^"]*)" to "([^"]*)"$`, imapClientCopiesMessagesTo)
	s.Step(`^IMAP client moves messages "([^"]*)" to "([^"]*)"$`, imapClientMovesMessagesTo)
	s.Step(`^IMAP clients "([^"]*)" and "([^"]*)" move message "([^"]*)" of "([^"]*)" from "([^"]*)" to "([^"]*)" by append and delete$`, imapClientsMoveMessageOfUserFromToByAppendAndDelete)
	s.Step(`^IMAP client imports message to "([^"]*)"$`, imapClientCreatesMessage)
	s.Step(`^IMAP client imports message to "([^"]*)" with encoding "([^"]*)"$`, imapClientCreatesMessageWithEncoding)
	s.Step(`^IMAP client creates message "([^"]*)" from "([^"]*)" to "([^"]*)" with body "([^"]*)" in "([^"]*)"$`, imapClientCreatesMessageFromToWithBody)
	s.Step(`^IMAP client creates message "([^"]*)" from "([^"]*)" to address "([^"]*)" of "([^"]*)" with body "([^"]*)" in "([^"]*)"$`, imapClientCreatesMessageFromToAddressOfUserWithBody)
	s.Step(`^IMAP client creates message "([^"]*)" from address "([^"]*)" of "([^"]*)" to "([^"]*)" with body "([^"]*)" in "([^"]*)"$`, imapClientCreatesMessageFromAddressOfUserToWithBody)
	s.Step(`^IMAP client marks message "([^"]*)" with "([^"]*)"$`, imapClientMarksMessageWithFlags)
	s.Step(`^IMAP client "([^"]*)" marks message "([^"]*)" with "([^"]*)"$`, imapClientNamedMarksMessageWithFlags)
	s.Step(`^IMAP client marks message "([^"]*)" as read$`, imapClientMarksMessageAsRead)
	s.Step(`^IMAP client "([^"]*)" marks message "([^"]*)" as read$`, imapClientNamedMarksMessageAsRead)
	s.Step(`^IMAP client marks message "([^"]*)" as unread$`, imapClientMarksMessageAsUnread)
	s.Step(`^IMAP client "([^"]*)" marks message "([^"]*)" as unread$`, imapClientNamedMarksMessageAsUnread)
	s.Step(`^IMAP client marks message "([^"]*)" as starred$`, imapClientMarksMessageAsStarred)
	s.Step(`^IMAP client "([^"]*)" marks message "([^"]*)" as starred$`, imapClientNamedMarksMessageAsStarred)
	s.Step(`^IMAP client marks message "([^"]*)" as unstarred$`, imapClientMarksMessageAsUnstarred)
	s.Step(`^IMAP client "([^"]*)" marks message "([^"]*)" as unstarred$`, imapClientNamedMarksMessageAsUnstarred)
	s.Step(`^IMAP client marks message "([^"]*)" as deleted$`, imapClientMarksMessageAsDeleted)
	s.Step(`^IMAP client "([^"]*)" marks message "([^"]*)" as deleted$`, imapClientNamedMarksMessageAsDeleted)
	s.Step(`^IMAP client marks message "([^"]*)" as undeleted$`, imapClientMarksMessageAsUndeleted)
	s.Step(`^IMAP client "([^"]*)" marks message "([^"]*)" as undeleted$`, imapClientNamedMarksMessageAsUndeleted)
	s.Step(`^IMAP client starts IDLE-ing$`, imapClientStartsIDLEing)
	s.Step(`^IMAP client "([^"]*)" starts IDLE-ing$`, imapClientNamedStartsIDLEing)
	s.Step(`^IMAP client sends expunge$`, imapClientExpunge)
	s.Step(`^IMAP client "([^"]*)" sends expunge$`, imapClientNamedExpunge)
}

func imapClientSendsCommand(command string) error {
	res := ctx.GetIMAPClient("imap").SendCommand(command)
	ctx.SetIMAPLastResponse("imap", res)
	return nil
}

func imapClientFetches(fetchRange string) error {
	res := ctx.GetIMAPClient("imap").Fetch(fetchRange, "UID")
	ctx.SetIMAPLastResponse("imap", res)
	return nil
}

func imapClientFetchesByUID(fetchRange string) error {
	res := ctx.GetIMAPClient("imap").FetchUID(fetchRange, "UID")
	ctx.SetIMAPLastResponse("imap", res)
	return nil
}

func imapClientSearchesFor(query string) error {
	res := ctx.GetIMAPClient("imap").Search(query)
	ctx.SetIMAPLastResponse("imap", res)
	return nil
}

func imapClientCopiesMessagesTo(messageRange, newMailboxName string) error {
	res := ctx.GetIMAPClient("imap").Copy(messageRange, newMailboxName)
	ctx.SetIMAPLastResponse("imap", res)
	return nil
}

func imapClientMovesMessagesTo(messageRange, newMailboxName string) error {
	res := ctx.GetIMAPClient("imap").Move(messageRange, newMailboxName)
	ctx.SetIMAPLastResponse("imap", res)
	return nil
}

func imapClientsMoveMessageOfUserFromToByAppendAndDelete(sourceIMAPClient, targetIMAPClient, messageUID, bddUserID, sourceMailboxName, targetMailboxName string) error {
	account := ctx.GetTestAccount(bddUserID)
	if account == nil {
		return godog.ErrPending
	}
	sourceMailbox, err := ctx.GetStoreMailbox(account.Username(), account.AddressID(), sourceMailboxName)
	if err != nil {
		return internalError(err, "getting store mailbox")
	}
	uid, err := strconv.ParseUint(messageUID, 10, 32)
	if err != nil {
		return internalError(err, "parsing message UID")
	}
	apiIDs, err := sourceMailbox.GetAPIIDsFromUIDRange(uint32(uid), uint32(uid))
	if err != nil {
		return internalError(err, "getting API IDs from sequence range")
	}
	message, err := sourceMailbox.GetMessage(apiIDs[0])
	if err != nil {
		return internalError(err, "getting message by ID")
	}

	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		msg := message.Message()
		_ = imapClientNamedCreatesMessageFromToWithBody(
			targetIMAPClient,
			msg.Subject,
			msg.Sender.String(),
			msg.ToList[0].String(),
			msg.Body,
			targetMailboxName,
		)
	}()

	go func() {
		defer wg.Done()
		_ = imapClientNamedMarksMessageAsDeleted(sourceIMAPClient, messageUID)
	}()

	wg.Wait()
	return nil
}

func imapClientCreatesMessage(mailboxName string, message *gherkin.DocString) error {
	return imapClientCreatesMessageWithEncoding(mailboxName, "utf8", message)
}

func imapClientCreatesMessageWithEncoding(mailboxName, encodingName string, message *gherkin.DocString) error {
	encoding, _ := charset.Lookup(encodingName)

	msg := message.Content
	if encodingName != "utf8" {
		if encoding == nil {
			return fmt.Errorf("unsupported encoding %s", encodingName)
		}

		var err error
		msg, err = encoding.NewEncoder().String(message.Content)
		if err != nil {
			return internalError(err, "encoding message content")
		}
	}

	res := ctx.GetIMAPClient("imap").Append(mailboxName, msg)
	ctx.SetIMAPLastResponse("imap", res)
	return nil
}

func imapClientCreatesMessageFromToWithBody(subject, from, to, body, mailboxName string) error {
	return imapClientNamedCreatesMessageFromToWithBody("imap", subject, from, to, body, mailboxName)
}

func imapClientNamedCreatesMessageFromToWithBody(imapClient, subject, from, to, body, mailboxName string) error {
	res := ctx.GetIMAPClient(imapClient).AppendBody(mailboxName, subject, from, to, body)
	ctx.SetIMAPLastResponse(imapClient, res)
	return nil
}

func imapClientCreatesMessageFromToAddressOfUserWithBody(subject, from, bddAddressID, bddUserID, body, mailboxName string) error {
	account := ctx.GetTestAccountWithAddress(bddUserID, bddAddressID)
	if account == nil {
		return godog.ErrPending
	}
	return imapClientCreatesMessageFromToWithBody(subject, from, account.Address(), body, mailboxName)
}

func imapClientCreatesMessageFromAddressOfUserToWithBody(subject, bddAddressID, bddUserID, to, body, mailboxName string) error {
	account := ctx.GetTestAccountWithAddress(bddUserID, bddAddressID)
	if account == nil {
		return godog.ErrPending
	}

	return imapClientCreatesMessageFromToWithBody(subject, account.Address(), to, body, mailboxName)
}

func imapClientMarksMessageWithFlags(messageRange, flags string) error {
	return imapClientNamedMarksMessageWithFlags("imap", messageRange, flags)
}

func imapClientNamedMarksMessageWithFlags(imapClient, messageRange, flags string) error {
	res := ctx.GetIMAPClient(imapClient).SetFlags(messageRange, flags)
	ctx.SetIMAPLastResponse(imapClient, res)
	return nil
}

func imapClientMarksMessageAsRead(messageRange string) error {
	return imapClientNamedMarksMessageAsRead("imap", messageRange)
}

func imapClientNamedMarksMessageAsRead(imapClient, messageRange string) error {
	res := ctx.GetIMAPClient(imapClient).MarkAsRead(messageRange)
	ctx.SetIMAPLastResponse(imapClient, res)
	return nil
}

func imapClientMarksMessageAsUnread(messageRange string) error {
	return imapClientNamedMarksMessageAsUnread("imap", messageRange)
}

func imapClientNamedMarksMessageAsUnread(imapClient, messageRange string) error {
	res := ctx.GetIMAPClient(imapClient).MarkAsUnread(messageRange)
	ctx.SetIMAPLastResponse(imapClient, res)
	return nil
}

func imapClientMarksMessageAsStarred(messageRange string) error {
	return imapClientNamedMarksMessageAsStarred("imap", messageRange)
}

func imapClientNamedMarksMessageAsStarred(imapClient, messageRange string) error {
	res := ctx.GetIMAPClient(imapClient).MarkAsStarred(messageRange)
	ctx.SetIMAPLastResponse(imapClient, res)
	return nil
}

func imapClientMarksMessageAsUnstarred(messageRange string) error {
	return imapClientNamedMarksMessageAsUnstarred("imap", messageRange)
}

func imapClientNamedMarksMessageAsUnstarred(imapClient, messageRange string) error {
	res := ctx.GetIMAPClient(imapClient).MarkAsUnstarred(messageRange)
	ctx.SetIMAPLastResponse(imapClient, res)
	return nil
}

func imapClientMarksMessageAsDeleted(messageRange string) error {
	return imapClientNamedMarksMessageAsDeleted("imap", messageRange)
}

func imapClientNamedMarksMessageAsDeleted(imapClient, messageRange string) error {
	res := ctx.GetIMAPClient(imapClient).MarkAsDeleted(messageRange)
	ctx.SetIMAPLastResponse(imapClient, res)
	return nil
}

func imapClientMarksMessageAsUndeleted(messageRange string) error {
	return imapClientNamedMarksMessageAsUndeleted("imap", messageRange)
}

func imapClientNamedMarksMessageAsUndeleted(imapClient, messageRange string) error {
	res := ctx.GetIMAPClient(imapClient).MarkAsUndeleted(messageRange)
	ctx.SetIMAPLastResponse(imapClient, res)
	return nil
}

func imapClientStartsIDLEing() error {
	return imapClientNamedStartsIDLEing("imap")
}

func imapClientNamedStartsIDLEing(imapClient string) error {
	res := ctx.GetIMAPClient(imapClient).StartIDLE()
	ctx.SetIMAPLastResponse(imapClient, res)
	return nil
}

func imapClientExpunge() error {
	return imapClientNamedExpunge("imap")
}

func imapClientNamedExpunge(imapClient string) error {
	res := ctx.GetIMAPClient(imapClient).Expunge()
	ctx.SetIMAPLastResponse(imapClient, res)
	return nil
}
