package tests

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ProtonMail/gluon/rfc822"
	"github.com/bradenaw/juniper/xslices"
	"github.com/cucumber/godog"
	"github.com/cucumber/messages-go/v16"
	"github.com/google/uuid"
	"gitlab.protontech.ch/go/liteapi"
	"golang.org/x/exp/slices"
)

func (s *scenario) thereExistsAnAccountWithUsernameAndPassword(username, password string) error {
	userID, addrID, err := s.t.api.AddUser(username, password, username)
	if err != nil {
		return err
	}

	// Set the ID of this user.
	s.t.setUserID(username, userID)

	// Set the address ID of this user.
	s.t.setAddrID(userID, addrID)

	// Set the address of this user (right now just the same as the username, but let's stay flexible).
	s.t.setUserAddr(userID, username)

	return nil
}

func (s *scenario) theAccountHasCustomFolders(username string, count int) error {
	for idx := 0; idx < count; idx++ {
		if _, err := s.t.api.AddLabel(s.t.getUserID(username), uuid.NewString(), liteapi.LabelTypeFolder); err != nil {
			return err
		}
	}

	return nil
}

func (s *scenario) theAccountHasCustomLabels(username string, count int) error {
	for idx := 0; idx < count; idx++ {
		if _, err := s.t.api.AddLabel(s.t.getUserID(username), uuid.NewString(), liteapi.LabelTypeLabel); err != nil {
			return err
		}
	}

	return nil
}

func (s *scenario) theAccountHasTheFollowingCustomMailboxes(username string, table *godog.Table) error {
	type mailbox struct {
		name string
		typ  liteapi.LabelType
	}

	wantMailboxes := xslices.Map(table.Rows[1:], func(row *messages.PickleTableRow) mailbox {
		var mailboxType liteapi.LabelType

		switch row.Cells[1].Value {
		case "folder":
			mailboxType = liteapi.LabelTypeFolder
		case "label":
			mailboxType = liteapi.LabelTypeLabel
		}

		return mailbox{
			name: row.Cells[0].Value,
			typ:  mailboxType,
		}
	})

	for _, wantMailbox := range wantMailboxes {
		if _, err := s.t.api.AddLabel(s.t.getUserID(username), wantMailbox.name, wantMailbox.typ); err != nil {
			return err
		}
	}

	return nil
}

func (s *scenario) theAccountHasTheFollowingMessagesInMailbox(username, mailbox string, table *godog.Table) error {
	userID := s.t.getUserID(username)
	addrID := s.t.getAddrID(userID)
	mboxID := s.t.getMBoxID(userID, mailbox)

	for _, wantMessage := range parseMessages(table) {
		if _, err := s.t.api.AddMessage(
			userID,
			addrID,
			[]string{mboxID},
			wantMessage.Sender,
			wantMessage.Recipient,
			wantMessage.Subject,
			"some body goes here",
			rfc822.TextPlain,
			wantMessage.Unread,
			false,
		); err != nil {
			return err
		}
	}

	return nil
}

func (s *scenario) theAccountHasMessagesInMailbox(username string, count int, mailbox string) error {
	userID := s.t.getUserID(username)
	addrID := s.t.getAddrID(userID)
	mboxID := s.t.getMBoxID(userID, mailbox)

	for idx := 0; idx < count; idx++ {
		if _, err := s.t.api.AddMessage(
			userID,
			addrID,
			[]string{mboxID},
			fmt.Sprintf("sender%v@pm.me", idx),
			fmt.Sprintf("recipient%v@pm.me", idx),
			fmt.Sprintf("subject %v", idx),
			fmt.Sprintf("body %v", idx),
			rfc822.TextPlain,
			false,
			false,
		); err != nil {
			return err
		}
	}

	return nil
}

func (s *scenario) userLogsInWithUsernameAndPassword(username, password string) error {
	userID, err := s.t.bridge.LoginUser(context.Background(), username, password, nil, nil)
	if err != nil {
		s.t.pushError(err)
	} else {
		if userID != s.t.getUserID(username) {
			return errors.New("user ID mismatch")
		}

		info, err := s.t.bridge.GetUserInfo(userID)
		if err != nil {
			return err
		}

		s.t.setUserPass(userID, info.BridgePass)
	}

	return nil
}

func (s *scenario) userLogsOut(username string) error {
	return s.t.bridge.LogoutUser(context.Background(), s.t.getUserID(username))
}

func (s *scenario) userIsDeleted(username string) error {
	return s.t.bridge.DeleteUser(context.Background(), s.t.getUserID(username))
}

func (s *scenario) theAuthOfUserIsRevoked(username string) error {
	return s.t.api.RevokeUser(s.t.getUserID(username))
}

func (s *scenario) userIsListedAndConnected(username string) error {
	user, err := s.t.bridge.GetUserInfo(s.t.getUserID(username))
	if err != nil {
		return err
	}

	if user.Username != username {
		return errors.New("user not listed")
	}

	if !user.Connected {
		return errors.New("user not connected")
	}

	return nil
}

func (s *scenario) userIsEventuallyListedAndConnected(username string) error {
	return eventually(
		func() error { return s.userIsListedAndConnected(username) },
		5*time.Second,
		100*time.Millisecond,
	)
}

func (s *scenario) userIsListedButNotConnected(username string) error {
	user, err := s.t.bridge.GetUserInfo(s.t.getUserID(username))
	if err != nil {
		return err
	}

	if user.Username != username {
		return errors.New("user not listed")
	}

	if user.Connected {
		return errors.New("user connected")
	}

	return nil
}

func (s *scenario) userIsNotListed(username string) error {
	if slices.Contains(s.t.bridge.GetUserIDs(), s.t.getUserID(username)) {
		return errors.New("user listed")
	}

	return nil
}

func (s *scenario) userFinishesSyncing(username string) error {
	return s.bridgeSendsSyncStartedAndFinishedEventsForUser(username)
}
