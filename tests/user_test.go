package tests

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/cucumber/godog"
	"github.com/google/uuid"
	"gitlab.protontech.ch/go/liteapi"
	"golang.org/x/exp/slices"
)

func (s *scenario) thereExistsAnAccountWithUsernameAndPassword(username, password string) error {
	// Create the user.
	userID, addrID, err := s.t.api.CreateUser(username, username, []byte(password))
	if err != nil {
		return err
	}

	// Set the ID of this user.
	s.t.setUserID(username, userID)

	// Set the password of this user.
	s.t.setUserPass(userID, password)

	// Set the address of this user (right now just the same as the username, but let's stay flexible).
	s.t.setUserAddr(userID, addrID, username)

	return nil
}

func (s *scenario) theAccountHasAdditionalAddress(username, address string) error {
	userID := s.t.getUserID(username)

	addrID, err := s.t.api.CreateAddress(userID, address, []byte(s.t.getUserPass(userID)))
	if err != nil {
		return err
	}

	s.t.setUserAddr(userID, addrID, address)

	return nil
}

func (s *scenario) theAccountNoLongerHasAdditionalAddress(username, address string) error {
	userID := s.t.getUserID(username)
	addrID := s.t.getUserAddrID(userID, address)

	if err := s.t.api.RemoveAddress(userID, addrID); err != nil {
		return err
	}

	s.t.unsetUserAddr(userID, addrID)

	return nil
}

func (s *scenario) theAccountHasCustomFolders(username string, count int) error {
	for idx := 0; idx < count; idx++ {
		if _, err := s.t.api.CreateLabel(s.t.getUserID(username), uuid.NewString(), liteapi.LabelTypeFolder); err != nil {
			return err
		}
	}

	return nil
}

func (s *scenario) theAccountHasCustomLabels(username string, count int) error {
	for idx := 0; idx < count; idx++ {
		if _, err := s.t.api.CreateLabel(s.t.getUserID(username), uuid.NewString(), liteapi.LabelTypeLabel); err != nil {
			return err
		}
	}

	return nil
}

func (s *scenario) theAccountHasTheFollowingCustomMailboxes(username string, table *godog.Table) error {
	type CustomMailbox struct {
		Name string `bdd:"name"`
		Type string `bdd:"type"`
	}

	wantMailboxes, err := unmarshalTable[CustomMailbox](table)
	if err != nil {
		return err
	}

	for _, wantMailbox := range wantMailboxes {
		var labelType liteapi.LabelType

		switch wantMailbox.Type {
		case "folder":
			labelType = liteapi.LabelTypeFolder

		case "label":
			labelType = liteapi.LabelTypeLabel
		}

		if _, err := s.t.api.CreateLabel(s.t.getUserID(username), wantMailbox.Name, labelType); err != nil {
			return err
		}
	}

	return nil
}

func (s *scenario) theAddressOfAccountHasTheFollowingMessagesInMailbox(address, username, mailbox string, table *godog.Table) error {
	userID := s.t.getUserID(username)
	addrID := s.t.getUserAddrID(userID, address)
	mboxID := s.t.getMBoxID(userID, mailbox)

	wantMessages, err := unmarshalTable[Message](table)
	if err != nil {
		return err
	}

	for _, wantMessage := range wantMessages {
		messageID, err := s.t.api.CreateMessage(userID, addrID, wantMessage.Build(), liteapi.MessageFlagReceived, wantMessage.Unread, false)
		if err != nil {
			return err
		}

		if err := s.t.api.LabelMessage(userID, messageID, mboxID); err != nil {
			return err
		}
	}

	return nil
}

func (s *scenario) theAddressOfAccountHasMessagesInMailbox(address, username string, count int, mailbox string) error {
	userID := s.t.getUserID(username)
	addrID := s.t.getUserAddrID(userID, address)
	mboxID := s.t.getMBoxID(userID, mailbox)

	for idx := 0; idx < count; idx++ {
		messageID, err := s.t.api.CreateMessage(userID, addrID, Message{
			Subject: fmt.Sprintf("%d", idx),
			To:      fmt.Sprintf("%d@pm.me", idx),
			From:    fmt.Sprintf("%d@pm.me", idx),
			Body:    fmt.Sprintf("body %d", idx),
		}.Build(), liteapi.MessageFlagReceived, idx%2 == 0, false)
		if err != nil {
			return err
		}

		if err := s.t.api.LabelMessage(userID, messageID, mboxID); err != nil {
			return err
		}
	}

	return nil
}

func (s *scenario) userLogsInWithUsernameAndPassword(username, password string) error {
	userID, err := s.t.bridge.LoginUser(context.Background(), username, []byte(password), nil, nil)
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

		s.t.setUserBridgePass(userID, info.BridgePass)
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
