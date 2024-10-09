// Copyright (c) 2024 Proton AG
//
// This file is part of Proton Mail Bridge.
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
// along with Proton Mail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package tests

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ProtonMail/gluon/rfc822"
	"github.com/bradenaw/juniper/iterator"
	"github.com/bradenaw/juniper/xslices"
	"github.com/cucumber/godog"
	"github.com/emersion/go-imap"
	id "github.com/emersion/go-imap-id"
	"github.com/emersion/go-imap/client"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
)

func (s *scenario) userConnectsIMAPClient(username, clientID string) error {
	return s.t.newIMAPClient(s.t.getUserByName(username).getUserID(), clientID)
}

func (s *scenario) userConnectsIMAPClientOnPort(username, clientID string, port int) error {
	return s.t.newIMAPClientOnPort(s.t.getUserByName(username).getUserID(), clientID, port)
}

func (s *scenario) userConnectsAndAuthenticatesIMAPClient(username, clientID string) error {
	return s.userConnectsAndAuthenticatesIMAPClientWithAddress(username, clientID, s.t.getUserByName(username).getEmails()[0])
}

func (s *scenario) userConnectsAndAuthenticatesIMAPClientWithAddress(username, clientID, address string) error {
	if err := s.t.newIMAPClient(s.t.getUserByName(username).getUserID(), clientID); err != nil {
		return err
	}

	userID, client := s.t.getIMAPClient(clientID)

	return client.Login(address, s.t.getUserByID(userID).getBridgePass())
}

func (s *scenario) userConnectsAndCanNotAuthenticateIMAPClientWithAddress(username, clientID, address string) error {
	if err := s.t.newIMAPClient(s.t.getUserByName(username).getUserID(), clientID); err != nil {
		return err
	}

	userID, client := s.t.getIMAPClient(clientID)

	if err := client.Login(address, s.t.getUserByID(userID).getBridgePass()); err == nil {
		return fmt.Errorf("expected error, got nil")
	}

	return nil
}

func (s *scenario) imapClientCanAuthenticate(clientID string) error {
	userID, client := s.t.getIMAPClient(clientID)

	return client.Login(s.t.getUserByID(userID).getEmails()[0], s.t.getUserByID(userID).getBridgePass())
}

func (s *scenario) imapClientCanAuthenticateWithAddress(clientID string, address string) error {
	userID, client := s.t.getIMAPClient(clientID)

	return client.Login(address, s.t.getUserByID(userID).getBridgePass())
}

func (s *scenario) imapClientCannotAuthenticate(clientID string) error {
	userID, client := s.t.getIMAPClient(clientID)

	if err := client.Login(s.t.getUserByID(userID).getEmails()[0], s.t.getUserByID(userID).getBridgePass()); err == nil {
		return fmt.Errorf("expected error, got nil")
	}

	return nil
}

func (s *scenario) imapClientCannotAuthenticateWithAddress(clientID, address string) error {
	userID, client := s.t.getIMAPClient(clientID)

	if err := client.Login(address, s.t.getUserByID(userID).getBridgePass()); err == nil {
		return fmt.Errorf("expected error, got nil")
	}

	return nil
}

func (s *scenario) imapClientCannotAuthenticateWithIncorrectUsername(clientID string) error {
	userID, client := s.t.getIMAPClient(clientID)

	if err := client.Login(s.t.getUserByID(userID).getEmails()[0]+"bad", s.t.getUserByID(userID).getBridgePass()); err == nil {
		return fmt.Errorf("expected error, got nil")
	}

	return nil
}

func (s *scenario) imapClientCannotAuthenticateWithIncorrectPassword(clientID string) error {
	userID, client := s.t.getIMAPClient(clientID)
	badPass := base64.StdEncoding.EncodeToString([]byte("bad_password"))
	if err := client.Login(s.t.getUserByID(userID).getEmails()[0], badPass); err == nil {
		return fmt.Errorf("expected error, got nil")
	}

	return nil
}

func (s *scenario) imapClientCloses(clientID string) error {
	_, client := s.t.getIMAPClient(clientID)
	if err := client.Logout(); err != nil {
		return err
	}
	delete(s.t.imapClients, clientID)
	return nil
}

func (s *scenario) imapClientAnnouncesItsIDWithNameAndVersion(clientID, name, version string) error {
	_, client := s.t.getIMAPClient(clientID)

	if _, err := id.NewClient(client).ID(id.ID{id.FieldName: name, id.FieldVersion: version}); err != nil {
		return fmt.Errorf("expected error, got nil")
	}

	return nil
}

func (s *scenario) imapClientCreatesMailbox(clientID, mailbox string) error {
	_, client := s.t.getIMAPClient(clientID)

	s.t.pushError(client.Create(mailbox))

	return nil
}

func (s *scenario) imapClientDeletesMailbox(clientID, mailbox string) error {
	_, client := s.t.getIMAPClient(clientID)

	s.t.pushError(client.Delete(mailbox))

	return nil
}

func (s *scenario) imapClientRenamesMailboxTo(clientID, fromName, toName string) error {
	_, client := s.t.getIMAPClient(clientID)

	s.t.pushError(client.Rename(fromName, toName))

	return nil
}

func (s *scenario) imapClientSeesTheFollowingMailboxInfo(clientID string, table *godog.Table) error {
	_, client := s.t.getIMAPClient(clientID)

	status, err := clientStatus(client)
	if err != nil {
		return err
	}

	haveMailboxes := xslices.Map(status, newMailboxFromIMAP)

	wantMailboxes, err := unmarshalTable[Mailbox](table)
	if err != nil {
		return err
	}

	return matchMailboxes(haveMailboxes, wantMailboxes)
}

func (s *scenario) imapClientEventuallySeesTheFollowingMailboxInfo(clientID string, table *godog.Table) error {
	return eventually(func() error {
		return s.imapClientSeesTheFollowingMailboxInfo(clientID, table)
	})
}

func (s *scenario) imapClientSeesTheFollowingMailboxInfoForMailbox(clientID, mailbox string, table *godog.Table) error {
	_, client := s.t.getIMAPClient(clientID)

	status, err := clientStatus(client)
	if err != nil {
		return err
	}

	status = xslices.Filter(status, func(status *imap.MailboxStatus) bool {
		return status.Name == mailbox
	})

	haveMailboxes := xslices.Map(status, newMailboxFromIMAP)

	wantMailboxes, err := unmarshalTable[Mailbox](table)
	if err != nil {
		return err
	}

	return matchMailboxes(haveMailboxes, wantMailboxes)
}

func (s *scenario) imapClientSeesMailbox(clientID, mailbox string) error {
	_, client := s.t.getIMAPClient(clientID)

	mailboxes := clientList(client)

	if !slices.Contains(xslices.Map(mailboxes, func(info *imap.MailboxInfo) string { return info.Name }), mailbox) {
		return fmt.Errorf("expected %v to contain %v but it doesn't", mailboxes, mailbox)
	}

	return nil
}

func (s *scenario) imapClientDoesNotSeeMailbox(clientID, mailbox string) error {
	_, client := s.t.getIMAPClient(clientID)

	mailboxes := clientList(client)

	if slices.Contains(xslices.Map(mailboxes, func(info *imap.MailboxInfo) string { return info.Name }), mailbox) {
		return fmt.Errorf("expected %v to not contain %v but it does", mailboxes, mailbox)
	}

	return nil
}

func (s *scenario) imapClientCountsMailboxesUnder(clientID string, count int, parent string) error {
	_, client := s.t.getIMAPClient(clientID)

	mailboxes := clientList(client)

	mailboxes = xslices.Filter(mailboxes, func(info *imap.MailboxInfo) bool {
		return strings.HasPrefix(info.Name, parent) && info.Name != parent
	})

	if len(mailboxes) != count {
		return fmt.Errorf("expected %v to have %v mailboxes, got %v", parent, count, len(mailboxes))
	}

	return nil
}

func (s *scenario) imapClientSelectsMailbox(clientID, mailbox string) error {
	_, client := s.t.getIMAPClient(clientID)

	status, err := client.Select(mailbox, false)
	if err != nil {
		s.t.pushError(err)
	} else if status.Name != mailbox {
		return fmt.Errorf("expected mailbox %v, got %v", mailbox, status.Name)
	}

	return nil
}

func (s *scenario) imapClientCopiesTheMessageWithSubjectFromTo(clientID, subject, from, to string) error {
	_, client := s.t.getIMAPClient(clientID)

	uid, err := clientGetUIDBySubject(client, from, subject)
	if err != nil {
		return err
	}

	if err := clientCopy(client, from, to, uid); err != nil {
		s.t.pushError(err)
	}

	return nil
}

func (s *scenario) imapClientCopiesAllMessagesFromTo(clientID, from, to string) error {
	_, client := s.t.getIMAPClient(clientID)

	if err := clientCopy(client, from, to); err != nil {
		s.t.pushError(err)
	}

	return nil
}

func (s *scenario) imapClientSeesTheFollowingMessagesInMailbox(clientID, mailbox string, table *godog.Table) error {
	_, client := s.t.getIMAPClient(clientID)

	fetch, err := clientFetch(client, mailbox)
	if err != nil {
		return err
	}

	haveMessages := xslices.Map(fetch, newMessageFromIMAP)

	wantMessages, err := unmarshalTable[Message](table)
	if err != nil {
		return err
	}
	return matchMessages(haveMessages, wantMessages)
}

func (s *scenario) imapClientMovesTheMessageWithSubjectFromTo(clientID, subject, from, to string) error {
	_, client := s.t.getIMAPClient(clientID)

	uid, err := clientGetUIDBySubject(client, from, subject)
	if err != nil {
		return err
	}

	if err := clientMove(client, from, to, uid); err != nil {
		s.t.pushError(err)
	}

	return nil
}

func (s *scenario) imapClientMovesAllMessagesFromTo(clientID, from, to string) error {
	_, client := s.t.getIMAPClient(clientID)

	if err := clientMove(client, from, to); err != nil {
		s.t.pushError(err)
	}

	return nil
}

func (s *scenario) imapClientEventuallySeesTheFollowingMessagesInMailbox(clientID, mailbox string, table *godog.Table) error {
	return eventually(func() error {
		err := s.imapClientSeesTheFollowingMessagesInMailbox(clientID, mailbox, table)
		logrus.WithError(err).Trace("Matching eventually")
		return err
	})
}

func (s *scenario) imapClientSeesMessageInMailboxWithStructure(clientID, mailbox string, message *godog.DocString) error {
	return eventually(func() error {
		_, client := s.t.getIMAPClient(clientID)

		var msgStruct MessageStruct
		if err := json.Unmarshal([]byte(message.Content), &msgStruct); err != nil {
			return err
		}

		fetch, err := clientFetch(client, mailbox)
		if err != nil {
			return err
		}

		debug := false
		for iFetch := range fetch {
			if !debug {
				continue
			}

			fmt.Printf("\n\n\n fetch %d %#v\n evenlope %+v\n",
				iFetch, fetch[iFetch],
				fetch[iFetch].Envelope,
			)

			for _, v := range fetch[iFetch].Body {
				fmt.Println("body literal", v)
			}

			fmt.Printf("\n\n\n")
		}

		haveMessages := xslices.Map(fetch, newMessageStructFromIMAP)

		return matchStructure(haveMessages, msgStruct)
	})
}

func (s *scenario) imapClientSeesMessagesInMailbox(clientID string, count int, mailbox string) error {
	_, client := s.t.getIMAPClient(clientID)

	status, err := client.Status(mailbox, []imap.StatusItem{imap.StatusMessages})
	if err != nil {
		return err
	}

	if int(status.Messages) != count {
		return fmt.Errorf("expected mailbox %v to have %v items, got %v", mailbox, count, status.Messages)
	}

	return nil
}

func (s *scenario) imapClientEventuallySeesMessagesInMailbox(clientID string, count int, mailbox string) error {
	return eventually(func() error {
		return s.imapClientSeesMessagesInMailbox(clientID, count, mailbox)
	})
}

func (s *scenario) imapClientMarksMessageAsDeleted(clientID string, seq int) error {
	_, client := s.t.getIMAPClient(clientID)

	if _, err := clientStore(client, seq, seq, false, imap.FormatFlagsOp(imap.AddFlags, true), imap.DeletedFlag); err != nil {
		s.t.pushError(err)
	}

	return nil
}

func (s *scenario) imapClientMarksTheMessageWithSubjectAsDeleted(clientID, subject string) error {
	_, client := s.t.getIMAPClient(clientID)

	uid, err := clientGetUIDBySubject(client, client.Mailbox().Name, subject)
	if err != nil {
		return err
	}

	if _, err := clientStore(client, int(uid), int(uid), true, imap.FormatFlagsOp(imap.AddFlags, true), imap.DeletedFlag); err != nil {
		s.t.pushError(err)
	}

	return nil
}

func (s *scenario) imapClientMarksMessageAsNotDeleted(clientID string, seq int) error {
	_, client := s.t.getIMAPClient(clientID)

	_, err := clientStore(client, seq, seq, false, imap.FormatFlagsOp(imap.RemoveFlags, true), imap.DeletedFlag)
	if err != nil {
		s.t.pushError(err)
	}

	return nil
}

func (s *scenario) imapClientMarksAllMessagesAsDeleted(clientID string) error {
	_, client := s.t.getIMAPClient(clientID)

	_, err := clientStore(client, 1, int(client.Mailbox().Messages), false, imap.FormatFlagsOp(imap.AddFlags, true), imap.DeletedFlag)
	if err != nil {
		s.t.pushError(err)
	}

	return nil
}

func (s *scenario) imapClientMarksMessageAsState(clientID string, seq int, messageState string) error {
	_, client := s.t.getIMAPClient(clientID)

	err := clientChangeMessageState(client, seq, messageState, true)
	if err != nil {
		s.t.pushError(err)
	}

	return nil
}

func (s *scenario) imapClientMarksTheMessageWithSubjectAsState(clientID, subject, messageState string) error {
	_, client := s.t.getIMAPClient(clientID)

	uid, err := clientGetUIDBySubject(client, client.Mailbox().Name, subject)
	if err != nil {
		return err
	}

	if err := clientChangeMessageState(client, int(uid), messageState, true); err != nil {
		s.t.pushError(err)
	}

	return nil
}

func (s *scenario) imapClientMarksAllMessagesAsState(clientID, messageState string) error {
	_, client := s.t.getIMAPClient(clientID)

	if err := clientChangeAllMessageState(client, messageState); err != nil {
		s.t.pushError(err)
	}

	return nil
}

func (s *scenario) imapClientEventuallySeesThatMessageHasTheFlag(clientID string, seq int, flag string) error {
	return eventually(func() error {
		_, client := s.t.getIMAPClient(clientID)

		return clientIsFlagApplied(client, seq, flag, true, false)
	})
}

func (s *scenario) imapClientSeesThatMessageDoesNotHaveTheFlag(clientID string, seq int, flag string) error {
	_, client := s.t.getIMAPClient(clientID)

	return clientIsFlagApplied(client, seq, flag, false, false)
}

func (s *scenario) imapClientEventuallySeesThatTheMessageWithSubjectHasTheFlag(clientID, subject, flag string) error {
	return eventually(func() error {
		_, client := s.t.getIMAPClient(clientID)

		uid, err := clientGetUIDBySubject(client, client.Mailbox().Name, subject)
		if err != nil {
			return err
		}

		return clientIsFlagApplied(client, int(uid), flag, true, false)
	})
}

func (s *scenario) imapClientEventuallySeesThatTheMessageWithSubjectDoesNotHaveTheFlag(clientID, subject, flag string) error {
	return eventually(func() error {
		_, client := s.t.getIMAPClient(clientID)

		uid, err := clientGetUIDBySubject(client, client.Mailbox().Name, subject)
		if err != nil {
			return err
		}

		return clientIsFlagApplied(client, int(uid), flag, false, false)
	})
}

func (s *scenario) imapClientEventuallySeesThatAllTheMessagesHaveTheFlag(clientID string, flag string) error {
	return eventually(func() error {
		_, client := s.t.getIMAPClient(clientID)

		return clientIsFlagApplied(client, 1, flag, true, true)
	})
}

func (s *scenario) imapClientEventuallySeesThatAllTheMessagesDoNotHaveTheFlag(clientID string, flag string) error {
	return eventually(func() error {
		_, client := s.t.getIMAPClient(clientID)

		return clientIsFlagApplied(client, 1, flag, false, true)
	})
}

func (s *scenario) imapClientExpunges(clientID string) error {
	_, client := s.t.getIMAPClient(clientID)

	s.t.pushError(client.Expunge(nil))

	return nil
}

func (s *scenario) imapClientAppendsTheFollowingMessageToMailbox(clientID string, mailbox string, docString *godog.DocString) error {
	_, client := s.t.getIMAPClient(clientID)

	if err := clientAppend(client, mailbox, docString.Content); err != nil {
		s.t.pushError(err)
	}

	return nil
}

func (s *scenario) imapClientAppendsTheFollowingMessagesToMailbox(clientID string, mailbox string, table *godog.Table) error {
	_, client := s.t.getIMAPClient(clientID)

	messages, err := unmarshalTable[Message](table)
	if err != nil {
		return err
	}

	for idx, message := range messages {
		if len(message.Date) == 0 {
			logrus.Warnf("Appended message has no date, adding default one so it does not fail test")
		}
		messages[idx].Date = "23 Feb 80 00:00 GMT"
	}

	for _, message := range messages {
		if err := clientAppend(client, mailbox, string(message.Build())); err != nil {
			s.t.pushError(err)
		}
	}

	return nil
}

func (s *scenario) imapClientAppendsToMailbox(clientID string, file, mailbox string) error {
	_, client := s.t.getIMAPClient(clientID)

	b, err := os.ReadFile(filepath.Join("testdata", file))
	if err != nil {
		return err
	}

	if err := clientAppend(client, mailbox, string(b)); err != nil {
		s.t.pushError(err)
	}

	return nil
}

func (s *scenario) imapClientsMoveMessageWithSubjectUserFromToByOrderedOperations(sourceIMAPClient, targetIMAPClient, messageSubject, _, targetMailboxName, op1, op2, op3 string) error {
	// call NOOP to prevent unilateral updates in following FETCH
	_, sourceClient := s.t.getIMAPClient(sourceIMAPClient)
	_, targetClient := s.t.getIMAPClient(targetIMAPClient)

	if err := sourceClient.Noop(); err != nil {
		return err
	}

	if err := targetClient.Noop(); err != nil {
		return err
	}

	uid, err := clientGetUIDBySubject(sourceClient, sourceClient.Mailbox().Name, messageSubject)
	if err != nil {
		return err
	}

	// get the original message
	messages, err := clientFetchSequence(sourceClient, fmt.Sprintf("%v", uid), true)
	if err != nil {
		return err
	}

	if len(messages) != 1 {
		return fmt.Errorf("more than one message in sequence set")
	}

	bodySection, err := imap.ParseBodySectionName("BODY[]")
	if err != nil {
		return err
	}

	literal, err := io.ReadAll(messages[0].GetBody(bodySection))
	if err != nil {
		return err
	}

	var targetErr error
	var storeErr error
	var expungeErr error

	for _, op := range []string{op1, op2, op3} {
		switch op {
		case "APPEND":

			flags := messages[0].Flags
			if index := xslices.Index(flags, imap.RecentFlag); index >= 0 {
				flags = xslices.Remove(flags, index, 1)
			}

			targetErr = targetClient.Append(targetMailboxName, flags, time.Now(), bytes.NewReader(literal))
		case "DELETE":
			if _, err := clientStore(sourceClient, int(uid), int(uid), true, imap.FormatFlagsOp(imap.AddFlags, true), imap.DeletedFlag); err != nil {
				storeErr = err
			}
		case "EXPUNGE":
			expungeErr = sourceClient.Expunge(nil)
		default:
			return errors.New("unknown IMAP operation " + op)
		}
		time.Sleep(100 * time.Millisecond)
	}

	if targetErr != nil || storeErr != nil || expungeErr != nil {
		return fmt.Errorf("one or more operations failed: append=%v store=%v expunge=%v", targetErr, storeErr, expungeErr)
	}

	return nil
}

func (s *scenario) imapClientSeesHeaderInMessageWithSubject(clientID, headerString, subject, mailbox string) error {
	_, client := s.t.getIMAPClient(clientID)

	messages, err := clientFetch(client, mailbox)
	if err != nil {
		return err
	}

	section, err := imap.ParseBodySectionName("BODY[]")
	if err != nil {
		return err
	}

	for _, m := range messages {
		if m.Envelope.Subject == subject {
			literal, err := io.ReadAll(m.GetBody(section))
			if err != nil {
				return err
			}

			header, _ := rfc822.Split(literal)

			if !bytes.Contains(header, []byte(headerString)) {
				return fmt.Errorf("message header does not contain '%v'", headerString)
			}

			return nil
		}
	}

	return fmt.Errorf("could not find message with given subject '%v'", subject)
}

func (s *scenario) imapClientDoesNotSeeHeaderInMessageWithSubject(clientID, headerString, subject, mailbox string) error {
	err := s.imapClientSeesHeaderInMessageWithSubject(clientID, headerString, subject, mailbox)
	if err == nil {
		return fmt.Errorf("message header contains '%v'", headerString)
	}
	return nil
}

func clientList(client *client.Client) []*imap.MailboxInfo {
	resCh := make(chan *imap.MailboxInfo)

	go func() {
		if err := client.List("", "*", resCh); err != nil {
			panic(err)
		}
	}()

	return iterator.Collect(iterator.Chan(resCh))
}

func clientStatus(client *client.Client) ([]*imap.MailboxStatus, error) {
	list := clientList(client)

	status := make([]*imap.MailboxStatus, 0, len(list))

	for _, info := range list {
		res, err := client.Status(info.Name, []imap.StatusItem{imap.StatusMessages, imap.StatusRecent, imap.StatusUidNext, imap.StatusUidValidity, imap.StatusUnseen})
		if err != nil {
			return nil, err
		}

		status = append(status, res)
	}

	return status, nil
}

func clientGetUIDBySubject(client *client.Client, mailbox, subject string) (uint32, error) {
	fetch, err := clientFetch(client, mailbox)
	if err != nil {
		return 0, err
	}

	for _, msg := range fetch {
		if msg.Envelope.Subject == subject {
			return msg.Uid, nil
		}
	}

	return 0, fmt.Errorf("could not find message with subject %v", subject)
}

func clientFetch(client *client.Client, mailbox string) ([]*imap.Message, error) {
	status, err := client.Select(mailbox, false)
	if err != nil {
		return nil, err
	}

	if status.Messages == 0 {
		return nil, nil
	}

	resCh := make(chan *imap.Message)

	go func() {
		if err := client.Fetch(
			&imap.SeqSet{Set: []imap.Seq{{Start: 1, Stop: status.Messages}}},
			[]imap.FetchItem{imap.FetchFlags, imap.FetchEnvelope, imap.FetchUid, "BODY.PEEK[]"},
			resCh,
		); err != nil {
			panic(err)
		}
	}()

	return iterator.Collect(iterator.Chan(resCh)), nil
}

func clientFetchSequence(client *client.Client, sequenceSet string, isUID bool) ([]*imap.Message, error) {
	seqSet, err := imap.ParseSeqSet(sequenceSet)
	if err != nil {
		return nil, err
	}

	resCh := make(chan *imap.Message)

	go func() {
		if isUID {
			if err := client.UidFetch(
				seqSet,
				[]imap.FetchItem{imap.FetchFlags, imap.FetchEnvelope, imap.FetchUid, "BODY.PEEK[]"},
				resCh,
			); err != nil {
				panic(err)
			}
		} else {
			if err := client.Fetch(
				seqSet,
				[]imap.FetchItem{imap.FetchFlags, imap.FetchEnvelope, imap.FetchUid, "BODY.PEEK[]"},
				resCh,
			); err != nil {
				panic(err)
			}
		}
	}()

	return iterator.Collect(iterator.Chan(resCh)), nil
}

func clientCopy(client *client.Client, from, to string, uid ...uint32) error {
	status, err := client.Select(from, false)
	if err != nil {
		return err
	}

	if status.Messages == 0 {
		return fmt.Errorf("expected %v to have messages, but it doesn't", from)
	}

	var seqset *imap.SeqSet

	if len(uid) == 0 {
		seqset = &imap.SeqSet{Set: []imap.Seq{{Start: 1, Stop: status.Messages}}}
	} else {
		seqset = &imap.SeqSet{}

		for _, uid := range uid {
			seqset.AddNum(uid)
		}
	}

	return client.UidCopy(seqset, to)
}

func clientMove(client *client.Client, from, to string, uid ...uint32) error {
	status, err := client.Select(from, false)
	if err != nil {
		return err
	}

	if status.Messages == 0 {
		return fmt.Errorf("expected %v to have messages, but it doesn't", from)
	}

	var seqset *imap.SeqSet

	if len(uid) == 0 {
		seqset = &imap.SeqSet{Set: []imap.Seq{{Start: 1, Stop: status.Messages}}}
	} else {
		seqset = &imap.SeqSet{}

		for _, uid := range uid {
			seqset.AddNum(uid)
		}
	}

	return client.UidMove(seqset, to)
}

func clientStore(client *client.Client, from, to int, isUID bool, item imap.StoreItem, flags ...string) ([]*imap.Message, error) { //nolint:unparam
	resCh := make(chan *imap.Message)

	go func() {
		var storeFunc func(seqset *imap.SeqSet, item imap.StoreItem, value interface{}, ch chan *imap.Message) error

		if isUID {
			storeFunc = client.UidStore
		} else {
			storeFunc = client.Store
		}

		if err := storeFunc(
			&imap.SeqSet{Set: []imap.Seq{{Start: uint32(from), Stop: uint32(to)}}},
			item,
			xslices.Map(flags, func(flag string) interface{} { return flag }),
			resCh,
		); err != nil {
			panic(err)
		}
	}()

	return iterator.Collect(iterator.Chan(resCh)), nil
}

func clientAppend(client *client.Client, mailbox string, literal string) error {
	return client.Append(mailbox, []string{}, time.Now(), strings.NewReader(literal))
}

func clientIsFlagApplied(client *client.Client, seq int, flag string, applied bool, wholeMailbox bool) error {
	fetch, err := clientFetch(client, client.Mailbox().Name)
	if err != nil {
		return err
	}

	idx := xslices.IndexFunc(fetch, func(msg *imap.Message) bool {
		return msg.SeqNum == uint32(seq)
	})

	if slices.Contains(fetch[idx].Flags, flag) != applied {
		return fmt.Errorf("expected message %v to have flag %v set to %v, got %v", seq, flag, applied, fetch[idx].Flags)
	}

	if wholeMailbox {
		for i := seq; i <= int(client.Mailbox().Messages); i++ {
			idx := xslices.IndexFunc(fetch, func(msg *imap.Message) bool {
				return msg.SeqNum == uint32(i)
			})

			if slices.Contains(fetch[idx].Flags, flag) != applied {
				return fmt.Errorf("expected message %v to have flag %v set to %v, got %v", seq, flag, applied, fetch[idx].Flags)
			}
		}
	}

	return nil
}

func clientChangeMessageState(client *client.Client, seq int, messageState string, isUID bool) error {
	switch {
	case messageState == "read":
		_, err := clientStore(client, seq, seq, isUID, imap.FormatFlagsOp(imap.AddFlags, true), imap.SeenFlag)
		if err != nil {
			return err
		}

	case messageState == "unread":
		_, err := clientStore(client, seq, seq, isUID, imap.FormatFlagsOp(imap.RemoveFlags, true), imap.SeenFlag)
		if err != nil {
			return err
		}

	case messageState == "starred":
		_, err := clientStore(client, seq, seq, isUID, imap.FormatFlagsOp(imap.AddFlags, true), imap.FlaggedFlag)
		if err != nil {
			return err
		}

	case messageState == "unstarred":
		_, err := clientStore(client, seq, seq, isUID, imap.FormatFlagsOp(imap.RemoveFlags, true), imap.FlaggedFlag)
		if err != nil {
			return err
		}

	case messageState == "forwarded":
		_, err := clientStore(client, seq, seq, isUID, imap.FormatFlagsOp(imap.AddFlags, true), "Forwarded")
		if err != nil {
			return err
		}

	case messageState == "unforwarded":
		_, err := clientStore(client, seq, seq, isUID, imap.FormatFlagsOp(imap.RemoveFlags, true), "Forwarded")
		if err != nil {
			return err
		}

	case messageState == "replied":
		_, err := clientStore(client, seq, seq, isUID, imap.FormatFlagsOp(imap.AddFlags, true), imap.AnsweredFlag)
		if err != nil {
			return err
		}

	case messageState == "unreplied":
		_, err := clientStore(client, seq, seq, isUID, imap.FormatFlagsOp(imap.RemoveFlags, true), imap.AnsweredFlag)
		if err != nil {
			return err
		}
	}

	return nil
}

func clientChangeAllMessageState(client *client.Client, messageState string) error {
	if messageState == "read" {
		_, err := clientStore(client, 1, int(client.Mailbox().Messages), false, imap.FormatFlagsOp(imap.AddFlags, true), imap.SeenFlag)
		if err != nil {
			return err
		}
	} else if messageState == "unread" {
		_, err := clientStore(client, 1, int(client.Mailbox().Messages), false, imap.FormatFlagsOp(imap.RemoveFlags, true), imap.SeenFlag)
		if err != nil {
			return err
		}
	}

	return nil
}
