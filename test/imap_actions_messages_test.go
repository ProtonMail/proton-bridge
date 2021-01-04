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
	s.Step(`^IMAP client copies message seq "([^"]*)" to "([^"]*)"$`, imapClientCopiesMessagesTo)
	s.Step(`^IMAP client moves message seq "([^"]*)" to "([^"]*)"$`, imapClientMovesMessagesTo)
	s.Step(`^IMAP clients "([^"]*)" and "([^"]*)" move message seq "([^"]*)" of "([^"]*)" from "([^"]*)" to "([^"]*)" by append and delete$`, imapClientsMoveMessageSeqOfUserFromToByAppendAndDelete)
	s.Step(`^IMAP client imports message to "([^"]*)"$`, imapClientCreatesMessage)
	s.Step(`^IMAP client imports message to "([^"]*)" with encoding "([^"]*)"$`, imapClientCreatesMessageWithEncoding)
	s.Step(`^IMAP client creates message "([^"]*)" from "([^"]*)" to "([^"]*)" with body "([^"]*)" in "([^"]*)"$`, imapClientCreatesMessageFromToWithBody)
	s.Step(`^IMAP client creates message "([^"]*)" from "([^"]*)" to address "([^"]*)" of "([^"]*)" with body "([^"]*)" in "([^"]*)"$`, imapClientCreatesMessageFromToAddressOfUserWithBody)
	s.Step(`^IMAP client creates message "([^"]*)" from address "([^"]*)" of "([^"]*)" to "([^"]*)" with body "([^"]*)" in "([^"]*)"$`, imapClientCreatesMessageFromAddressOfUserToWithBody)
	s.Step(`^IMAP client marks message seq "([^"]*)" with "([^"]*)"$`, imapClientMarksMessageSeqWithFlags)
	s.Step(`^IMAP client "([^"]*)" marks message seq "([^"]*)" with "([^"]*)"$`, imapClientNamedMarksMessageSeqWithFlags)
	s.Step(`^IMAP client marks message seq "([^"]*)" as read$`, imapClientMarksMessageSeqAsRead)
	s.Step(`^IMAP client "([^"]*)" marks message seq "([^"]*)" as read$`, imapClientNamedMarksMessageSeqAsRead)
	s.Step(`^IMAP client marks message seq "([^"]*)" as unread$`, imapClientMarksMessageSeqAsUnread)
	s.Step(`^IMAP client "([^"]*)" marks message seq "([^"]*)" as unread$`, imapClientNamedMarksMessageSeqAsUnread)
	s.Step(`^IMAP client marks message seq "([^"]*)" as starred$`, imapClientMarksMessageSeqAsStarred)
	s.Step(`^IMAP client "([^"]*)" marks message seq "([^"]*)" as starred$`, imapClientNamedMarksMessageSeqAsStarred)
	s.Step(`^IMAP client marks message seq "([^"]*)" as unstarred$`, imapClientMarksMessageSeqAsUnstarred)
	s.Step(`^IMAP client "([^"]*)" marks message seq "([^"]*)" as unstarred$`, imapClientNamedMarksMessageSeqAsUnstarred)
	s.Step(`^IMAP client marks message seq "([^"]*)" as deleted$`, imapClientMarksMessageSeqAsDeleted)
	s.Step(`^IMAP client "([^"]*)" marks message seq "([^"]*)" as deleted$`, imapClientNamedMarksMessageSeqAsDeleted)
	s.Step(`^IMAP client marks message seq "([^"]*)" as undeleted$`, imapClientMarksMessageSeqAsUndeleted)
	s.Step(`^IMAP client "([^"]*)" marks message seq "([^"]*)" as undeleted$`, imapClientNamedMarksMessageSeqAsUndeleted)
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

func imapClientCopiesMessagesTo(messageSeq, newMailboxName string) error {
	res := ctx.GetIMAPClient("imap").Copy(messageSeq, newMailboxName)
	ctx.SetIMAPLastResponse("imap", res)
	return nil
}

func imapClientMovesMessagesTo(messageSeq, newMailboxName string) error {
	res := ctx.GetIMAPClient("imap").Move(messageSeq, newMailboxName)
	ctx.SetIMAPLastResponse("imap", res)
	return nil
}

func imapClientsMoveMessageSeqOfUserFromToByAppendAndDelete(sourceIMAPClient, targetIMAPClient, messageSeq, bddUserID, sourceMailboxName, targetMailboxName string) error {
	account := ctx.GetTestAccount(bddUserID)
	if account == nil {
		return godog.ErrPending
	}
	sourceMailbox, err := ctx.GetStoreMailbox(account.Username(), account.AddressID(), sourceMailboxName)
	if err != nil {
		return internalError(err, "getting store mailbox")
	}
	uid, err := strconv.ParseUint(messageSeq, 10, 32)
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
		_ = imapClientNamedMarksMessageSeqAsDeleted(sourceIMAPClient, messageSeq)
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

func imapClientMarksMessageSeqWithFlags(messageSeq, flags string) error {
	return imapClientNamedMarksMessageSeqWithFlags("imap", messageSeq, flags)
}

func imapClientNamedMarksMessageSeqWithFlags(imapClient, messageSeq, flags string) error {
	res := ctx.GetIMAPClient(imapClient).SetFlags(messageSeq, flags)
	ctx.SetIMAPLastResponse(imapClient, res)
	return nil
}

func imapClientMarksMessageSeqAsRead(messageSeq string) error {
	return imapClientNamedMarksMessageSeqAsRead("imap", messageSeq)
}

func imapClientNamedMarksMessageSeqAsRead(imapClient, messageSeq string) error {
	res := ctx.GetIMAPClient(imapClient).MarkAsRead(messageSeq)
	ctx.SetIMAPLastResponse(imapClient, res)
	return nil
}

func imapClientMarksMessageSeqAsUnread(messageSeq string) error {
	return imapClientNamedMarksMessageSeqAsUnread("imap", messageSeq)
}

func imapClientNamedMarksMessageSeqAsUnread(imapClient, messageSeq string) error {
	res := ctx.GetIMAPClient(imapClient).MarkAsUnread(messageSeq)
	ctx.SetIMAPLastResponse(imapClient, res)
	return nil
}

func imapClientMarksMessageSeqAsStarred(messageSeq string) error {
	return imapClientNamedMarksMessageSeqAsStarred("imap", messageSeq)
}

func imapClientNamedMarksMessageSeqAsStarred(imapClient, messageSeq string) error {
	res := ctx.GetIMAPClient(imapClient).MarkAsStarred(messageSeq)
	ctx.SetIMAPLastResponse(imapClient, res)
	return nil
}

func imapClientMarksMessageSeqAsUnstarred(messageSeq string) error {
	return imapClientNamedMarksMessageSeqAsUnstarred("imap", messageSeq)
}

func imapClientNamedMarksMessageSeqAsUnstarred(imapClient, messageSeq string) error {
	res := ctx.GetIMAPClient(imapClient).MarkAsUnstarred(messageSeq)
	ctx.SetIMAPLastResponse(imapClient, res)
	return nil
}

func imapClientMarksMessageSeqAsDeleted(messageSeq string) error {
	return imapClientNamedMarksMessageSeqAsDeleted("imap", messageSeq)
}

func imapClientNamedMarksMessageSeqAsDeleted(imapClient, messageSeq string) error {
	res := ctx.GetIMAPClient(imapClient).MarkAsDeleted(messageSeq)
	ctx.SetIMAPLastResponse(imapClient, res)
	return nil
}

func imapClientMarksMessageSeqAsUndeleted(messageSeq string) error {
	return imapClientNamedMarksMessageSeqAsUndeleted("imap", messageSeq)
}

func imapClientNamedMarksMessageSeqAsUndeleted(imapClient, messageSeq string) error {
	res := ctx.GetIMAPClient(imapClient).MarkAsUndeleted(messageSeq)
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
