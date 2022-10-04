package tests

import (
	"fmt"
	"strings"
	"time"

	"github.com/bradenaw/juniper/iterator"
	"github.com/bradenaw/juniper/xslices"
	"github.com/cucumber/godog"
	"github.com/cucumber/messages-go/v16"
	"github.com/emersion/go-imap"
	id "github.com/emersion/go-imap-id"
	"github.com/emersion/go-imap/client"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"golang.org/x/exp/slices"
)

func (s *scenario) userConnectsIMAPClient(username, clientID string) error {
	return s.t.newIMAPClient(s.t.getUserID(username), clientID)
}

func (s *scenario) userConnectsIMAPClientOnPort(username, clientID string, port int) error {
	return s.t.newIMAPClientOnPort(s.t.getUserID(username), clientID, port)
}

func (s *scenario) userConnectsAndAuthenticatesIMAPClient(username, clientID string) error {
	return s.userConnectsAndAuthenticatesIMAPClientWithAddress(username, clientID, s.t.getUserAddrs(s.t.getUserID(username))[0])
}

func (s *scenario) userConnectsAndAuthenticatesIMAPClientWithAddress(username, clientID, address string) error {
	if err := s.t.newIMAPClient(s.t.getUserID(username), clientID); err != nil {
		return err
	}

	userID, client := s.t.getIMAPClient(clientID)

	return client.Login(address, s.t.getUserBridgePass(userID))
}

func (s *scenario) imapClientCanAuthenticate(clientID string) error {
	userID, client := s.t.getIMAPClient(clientID)

	return client.Login(s.t.getUserAddrs(userID)[0], s.t.getUserBridgePass(userID))
}

func (s *scenario) imapClientCannotAuthenticate(clientID string) error {
	userID, client := s.t.getIMAPClient(clientID)

	if err := client.Login(s.t.getUserAddrs(userID)[0], s.t.getUserBridgePass(userID)); err == nil {
		return fmt.Errorf("expected error, got nil")
	}

	return nil
}

func (s *scenario) imapClientCannotAuthenticateWithAddress(clientID, address string) error {
	userID, client := s.t.getIMAPClient(clientID)

	if err := client.Login(address, s.t.getUserBridgePass(userID)); err == nil {
		return fmt.Errorf("expected error, got nil")
	}

	return nil
}

func (s *scenario) imapClientCannotAuthenticateWithIncorrectUsername(clientID string) error {
	userID, client := s.t.getIMAPClient(clientID)

	if err := client.Login(s.t.getUserAddrs(userID)[0]+"bad", s.t.getUserBridgePass(userID)); err == nil {
		return fmt.Errorf("expected error, got nil")
	}

	return nil
}

func (s *scenario) imapClientCannotAuthenticateWithIncorrectPassword(clientID string) error {
	userID, client := s.t.getIMAPClient(clientID)

	if err := client.Login(s.t.getUserAddrs(userID)[0], s.t.getUserBridgePass(userID)+"bad"); err == nil {
		return fmt.Errorf("expected error, got nil")
	}

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

	haveMailboxes := xslices.Map(status, func(status *imap.MailboxStatus) Mailbox {
		return newMailboxFromIMAP(status)
	})

	wantMailboxes, err := unmarshalTable[Mailbox](table)
	if err != nil {
		return err
	}

	return matchMailboxes(haveMailboxes, wantMailboxes)
}

func (s *scenario) imapClientEventuallySeesTheFollowingMailboxInfo(clientID string, table *godog.Table) error {
	return eventually(
		func() error { return s.imapClientSeesTheFollowingMailboxInfo(clientID, table) },
		5*time.Second,
		100*time.Millisecond,
	)
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

	haveMailboxes := xslices.Map(status, func(info *imap.MailboxStatus) Mailbox {
		return newMailboxFromIMAP(info)
	})

	wantMailboxes, err := unmarshalTable[Mailbox](table)
	if err != nil {
		return err
	}

	return matchMailboxes(haveMailboxes, wantMailboxes)
}

func (s *scenario) imapClientSeesTheFollowingMailboxes(clientID string, table *godog.Table) error {
	_, client := s.t.getIMAPClient(clientID)

	mailboxes, err := clientList(client)
	if err != nil {
		return err
	}

	have := xslices.Map(mailboxes, func(info *imap.MailboxInfo) string {
		return info.Name
	})

	want := xslices.Map(table.Rows[1:], func(row *messages.PickleTableRow) string {
		return row.Cells[0].Value
	})

	if !cmp.Equal(want, have, cmpopts.SortSlices(func(a, b string) bool { return a < b })) {
		return fmt.Errorf("want %v, have %v", want, have)
	}

	return nil
}

func (s *scenario) imapClientSeesMailbox(clientID, mailbox string) error {
	_, client := s.t.getIMAPClient(clientID)

	mailboxes, err := clientList(client)
	if err != nil {
		return err
	}

	if !slices.Contains(xslices.Map(mailboxes, func(info *imap.MailboxInfo) string { return info.Name }), mailbox) {
		return fmt.Errorf("expected %v to contain %v but it doesn't", mailboxes, mailbox)
	}

	return nil
}

func (s *scenario) imapClientDoesNotSeeMailbox(clientID, mailbox string) error {
	_, client := s.t.getIMAPClient(clientID)

	mailboxes, err := clientList(client)
	if err != nil {
		return err
	}

	if slices.Contains(xslices.Map(mailboxes, func(info *imap.MailboxInfo) string { return info.Name }), mailbox) {
		return fmt.Errorf("expected %v to not contain %v but it does", mailboxes, mailbox)
	}

	return nil
}

func (s *scenario) imapClientCountsMailboxesUnder(clientID string, count int, parent string) error {
	_, client := s.t.getIMAPClient(clientID)

	mailboxes, err := clientList(client)
	if err != nil {
		return err
	}

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

	return clientCopy(client, from, to, uid)
}

func (s *scenario) imapClientCopiesAllMessagesFromTo(clientID, from, to string) error {
	_, client := s.t.getIMAPClient(clientID)

	return clientCopy(client, from, to)
}

func (s *scenario) imapClientSeesTheFollowingMessagesInMailbox(clientID, mailbox string, table *godog.Table) error {
	_, client := s.t.getIMAPClient(clientID)

	fetch, err := clientFetch(client, mailbox)
	if err != nil {
		return err
	}

	haveMessages := xslices.Map(fetch, func(msg *imap.Message) Message {
		return newMessageFromIMAP(msg)
	})

	wantMessages, err := unmarshalTable[Message](table)
	if err != nil {
		return err
	}

	return matchMessages(haveMessages, wantMessages)
}

func (s *scenario) imapClientEventuallySeesTheFollowingMessagesInMailbox(clientID, mailbox string, table *godog.Table) error {
	return eventually(
		func() error { return s.imapClientSeesTheFollowingMessagesInMailbox(clientID, mailbox, table) },
		5*time.Second,
		500*time.Millisecond,
	)
}

func (s *scenario) imapClientSeesMessagesInMailbox(clientID string, count int, mailbox string) error {
	_, client := s.t.getIMAPClient(clientID)

	fetch, err := clientFetch(client, mailbox)
	if err != nil {
		return err
	}

	if len(fetch) != count {
		return fmt.Errorf("expected mailbox %v to be empty, got %v", mailbox, fetch)
	}

	return nil
}

func (s *scenario) imapClientEventuallySeesMessagesInMailbox(clientID string, count int, mailbox string) error {
	return eventually(
		func() error { return s.imapClientSeesMessagesInMailbox(clientID, count, mailbox) },
		5*time.Second,
		500*time.Millisecond,
	)
}

func (s *scenario) imapClientMarksMessageAsDeleted(clientID string, seq int) error {
	_, client := s.t.getIMAPClient(clientID)

	_, err := clientStore(client, seq, seq, imap.FormatFlagsOp(imap.AddFlags, true), imap.DeletedFlag)
	if err != nil {
		return err
	}

	return nil
}

func (s *scenario) imapClientMarksMessageAsNotDeleted(clientID string, seq int) error {
	_, client := s.t.getIMAPClient(clientID)

	_, err := clientStore(client, seq, seq, imap.FormatFlagsOp(imap.RemoveFlags, true), imap.DeletedFlag)
	if err != nil {
		return err
	}

	return nil
}

func (s *scenario) imapClientMarksAllMessagesAsDeleted(clientID string) error {
	_, client := s.t.getIMAPClient(clientID)

	_, err := clientStore(client, 1, int(client.Mailbox().Messages), imap.FormatFlagsOp(imap.AddFlags, true), imap.DeletedFlag)
	if err != nil {
		return err
	}

	return nil
}

func (s *scenario) imapClientSeesThatMessageHasTheFlag(clientID string, seq int, flag string) error {
	_, client := s.t.getIMAPClient(clientID)

	fetch, err := clientFetch(client, client.Mailbox().Name)
	if err != nil {
		return err
	}

	idx := xslices.IndexFunc(fetch, func(msg *imap.Message) bool {
		return msg.SeqNum == uint32(seq)
	})

	if !slices.Contains(fetch[idx].Flags, flag) {
		return fmt.Errorf("expected message %v to have flag %v, got %v", seq, flag, fetch[idx].Flags)
	}

	return nil
}

func (s *scenario) imapClientExpunges(clientID string) error {
	_, client := s.t.getIMAPClient(clientID)

	return client.Expunge(nil)
}

func clientList(client *client.Client) ([]*imap.MailboxInfo, error) {
	resCh := make(chan *imap.MailboxInfo)

	go func() {
		if err := client.List("", "*", resCh); err != nil {
			panic(err)
		}
	}()

	return iterator.Collect(iterator.Chan(resCh)), nil
}

func clientStatus(client *client.Client) ([]*imap.MailboxStatus, error) {
	var status []*imap.MailboxStatus

	list, err := clientList(client)
	if err != nil {
		return nil, err
	}

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
			[]imap.FetchItem{imap.FetchFlags, imap.FetchEnvelope, imap.FetchUid},
			resCh,
		); err != nil {
			panic(err)
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

func clientStore(client *client.Client, from, to int, item imap.StoreItem, flags ...string) ([]*imap.Message, error) {
	resCh := make(chan *imap.Message)

	go func() {
		if err := client.Store(
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
