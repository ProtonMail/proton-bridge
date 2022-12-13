// Copyright (c) 2022 Proton AG
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
	"context"
	"errors"
	"fmt"
	"net/mail"

	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/v3/internal/bridge"
	"github.com/bradenaw/juniper/iterator"
	"github.com/bradenaw/juniper/xslices"
	"github.com/cucumber/godog"
	"github.com/google/uuid"
	"golang.org/x/exp/slices"
)

func (s *scenario) thereExistsAnAccountWithUsernameAndPassword(username, password string) error {
	// Create the user and generate its default address (with keys).
	if _, err := s.t.runQuarkCmd(
		context.Background(),
		"user:create",
		"--name", username,
		"--password", password,
		"--gen-keys", "RSA2048",
	); err != nil {
		return err
	}

	return s.t.withClientPass(context.Background(), username, password, func(ctx context.Context, c *proton.Client) error {
		user, err := c.GetUser(ctx)
		if err != nil {
			return err
		}

		// Decrypt the user's encrypted ID for use with quark.
		userDecID, err := s.t.runQuarkCmd(context.Background(), "encryption:id", "--decrypt", user.ID)
		if err != nil {
			return err
		}

		// Upgrade the user to a paid account.
		if _, err := s.t.runQuarkCmd(
			context.Background(),
			"user:create:subscription",
			"--planID", "plus",
			string(userDecID),
		); err != nil {
			return err
		}

		addr, err := c.GetAddresses(ctx)
		if err != nil {
			return err
		}

		// Set the ID of the user.
		s.t.setUserID(username, user.ID)

		// Set the password of the user.
		s.t.setUserPass(user.ID, password)

		// Set the address of the user.
		s.t.setUserAddr(user.ID, addr[0].ID, addr[0].Email)

		return nil
	})
}

func (s *scenario) theAccountHasAdditionalAddress(username, address string) error {
	userID := s.t.getUserID(username)

	// Decrypt the user's encrypted ID for use with quark.
	userDecID, err := s.t.runQuarkCmd(context.Background(), "encryption:id", "--decrypt", userID)
	if err != nil {
		return err
	}

	// Create the user's additional address.
	if _, err := s.t.runQuarkCmd(
		context.Background(),
		"user:create:address",
		"--gen-keys", "RSA2048",
		string(userDecID),
		s.t.getUserPass(userID),
		address,
	); err != nil {
		return err
	}

	return s.t.withClient(context.Background(), username, func(ctx context.Context, c *proton.Client) error {
		addr, err := c.GetAddresses(ctx)
		if err != nil {
			return err
		}

		// Set the new address of the user.
		s.t.setUserAddr(userID, addr[len(addr)-1].ID, address)

		return nil
	})
}

func (s *scenario) theAccountHasAdditionalAddressWithoutKeys(username, address string) error {
	userID := s.t.getUserID(username)

	// Decrypt the user's encrypted ID for use with quark.
	userDecID, err := s.t.runQuarkCmd(context.Background(), "--decrypt", "encryption:id", userID)
	if err != nil {
		return err
	}

	// Create the user's additional address.
	if _, err := s.t.runQuarkCmd(
		context.Background(),
		"user:create:address",
		string(userDecID),
		s.t.getUserPass(userID),
		address,
	); err != nil {
		return err
	}

	return s.t.withClient(context.Background(), username, func(ctx context.Context, c *proton.Client) error {
		addr, err := c.GetAddresses(ctx)
		if err != nil {
			return err
		}

		// Set the new address of the user.
		s.t.setUserAddr(userID, addr[len(addr)-1].ID, address)

		return nil
	})
}

func (s *scenario) theAccountNoLongerHasAdditionalAddress(username, address string) error {
	userID := s.t.getUserID(username)
	addrID := s.t.getUserAddrID(userID, address)

	if err := s.t.withClient(context.Background(), username, func(ctx context.Context, c *proton.Client) error {
		if err := c.DisableAddress(ctx, addrID); err != nil {
			return err
		}

		return c.DeleteAddress(ctx, addrID)
	}); err != nil {
		return err
	}

	s.t.unsetUserAddr(userID, addrID)

	return nil
}

func (s *scenario) theAccountHasCustomFolders(username string, count int) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	return s.t.withClient(ctx, username, func(ctx context.Context, client *proton.Client) error {
		for idx := 0; idx < count; idx++ {
			if _, err := client.CreateLabel(ctx, proton.CreateLabelReq{
				Name: uuid.NewString(),
				Type: proton.LabelTypeFolder,
			}); err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *scenario) theAccountHasCustomLabels(username string, count int) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	return s.t.withClient(ctx, username, func(ctx context.Context, client *proton.Client) error {
		for idx := 0; idx < count; idx++ {
			if _, err := client.CreateLabel(ctx, proton.CreateLabelReq{
				Name: uuid.NewString(),
				Type: proton.LabelTypeLabel,
			}); err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *scenario) theAccountHasTheFollowingCustomMailboxes(username string, table *godog.Table) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	type CustomMailbox struct {
		Name string `bdd:"name"`
		Type string `bdd:"type"`
	}

	wantMailboxes, err := unmarshalTable[CustomMailbox](table)
	if err != nil {
		return err
	}

	return s.t.withClient(ctx, username, func(ctx context.Context, client *proton.Client) error {
		for _, wantMailbox := range wantMailboxes {
			var labelType proton.LabelType

			switch wantMailbox.Type {
			case "folder":
				labelType = proton.LabelTypeFolder

			case "label":
				labelType = proton.LabelTypeLabel
			}

			if _, err := client.CreateLabel(ctx, proton.CreateLabelReq{
				Name: wantMailbox.Name,
				Type: labelType,
			}); err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *scenario) theAddressOfAccountHasTheFollowingMessagesInMailbox(address, username, mailbox string, table *godog.Table) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	userID := s.t.getUserID(username)
	addrID := s.t.getUserAddrID(userID, address)
	mboxID := s.t.getMBoxID(userID, mailbox)

	wantMessages, err := unmarshalTable[Message](table)
	if err != nil {
		return err
	}

	return s.t.createMessages(ctx, username, addrID, xslices.Map(wantMessages, func(message Message) proton.ImportReq {
		return proton.ImportReq{
			Metadata: proton.ImportMetadata{
				AddressID: addrID,
				LabelIDs:  []string{mboxID},
				Unread:    proton.Bool(message.Unread),
				Flags:     proton.MessageFlagReceived,
			},
			Message: message.Build(),
		}
	}))
}

func (s *scenario) theAddressOfAccountHasMessagesInMailbox(address, username string, count int, mailbox string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	userID := s.t.getUserID(username)
	addrID := s.t.getUserAddrID(userID, address)
	mboxID := s.t.getMBoxID(userID, mailbox)

	return s.t.createMessages(ctx, username, addrID, iterator.Collect(iterator.Map(iterator.Counter(count), func(idx int) proton.ImportReq {
		return proton.ImportReq{
			Metadata: proton.ImportMetadata{
				AddressID: addrID,
				LabelIDs:  []string{mboxID},
				Flags:     proton.MessageFlagReceived,
			},
			Message: Message{
				Subject: fmt.Sprintf("%d", idx),
				To:      fmt.Sprintf("%d@pm.me", idx),
				From:    fmt.Sprintf("%d@pm.me", idx),
				Body:    fmt.Sprintf("body %d", idx),
			}.Build(),
		}
	})))
}

// accountDraftChanged changes the draft attributes, where draftIndex is
// similar to sequential ID i.e. 1 represents the first message of draft folder
// sorted by API creation time.
func (s *scenario) theFollowingFieldsWereChangedInDraftForAddressOfAccount(draftIndex int, address, username string, table *godog.Table) error {
	wantMessages, err := unmarshalTable[Message](table)
	if err != nil {
		return err
	}

	if len(wantMessages) != 1 {
		return fmt.Errorf("expected to have one row in table but got %d instead", len(wantMessages))
	}

	draftID, err := s.t.getDraftID(username, draftIndex)
	if err != nil {
		return fmt.Errorf("failed to get draft ID: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	return s.t.withClient(ctx, username, func(ctx context.Context, c *proton.Client) error {
		return s.t.withAddrKR(ctx, c, username, s.t.getUserAddrID(s.t.getUserID(username), address), func(_ context.Context, addrKR *crypto.KeyRing) error {
			var changes proton.DraftTemplate

			if wantMessages[0].From != "" {
				return fmt.Errorf("changing from address is not supported")
			}

			if wantMessages[0].To != "" {
				changes.ToList = []*mail.Address{{Address: wantMessages[0].To}}
			}

			if wantMessages[0].CC != "" {
				changes.CCList = []*mail.Address{{Address: wantMessages[0].CC}}
			}

			if wantMessages[0].BCC != "" {
				changes.BCCList = []*mail.Address{{Address: wantMessages[0].BCC}}
			}

			if wantMessages[0].Subject != "" {
				changes.Subject = wantMessages[0].Subject
			}

			if wantMessages[0].Body != "" {
				changes.Body = wantMessages[0].Body
			}

			if _, err := c.UpdateDraft(ctx, draftID, addrKR, proton.UpdateDraftReq{Message: changes}); err != nil {
				return fmt.Errorf("failed to update draft: %w", err)
			}

			return nil
		})
	})
}

func (s *scenario) userLogsInWithUsernameAndPassword(username, password string) error {
	userID, err := s.t.bridge.LoginFull(context.Background(), username, []byte(password), nil, nil)
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
	return s.t.withClient(context.Background(), username, func(ctx context.Context, client *proton.Client) error {
		return client.AuthRevokeAll(ctx)
	})
}

func (s *scenario) userIsListedAndConnected(username string) error {
	user, err := s.t.bridge.GetUserInfo(s.t.getUserID(username))
	if err != nil {
		return err
	}

	if user.Username != username {
		return errors.New("user not listed")
	}

	if user.State != bridge.Connected {
		return errors.New("user not connected")
	}

	return nil
}

func (s *scenario) userIsEventuallyListedAndConnected(username string) error {
	return eventually(func() error {
		return s.userIsListedAndConnected(username)
	})
}

func (s *scenario) userIsListedButNotConnected(username string) error {
	user, err := s.t.bridge.GetUserInfo(s.t.getUserID(username))
	if err != nil {
		return err
	}

	if user.Username != username {
		return errors.New("user not listed")
	}

	if user.State == bridge.Connected {
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
