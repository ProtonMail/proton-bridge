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
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/ProtonMail/gluon/rfc822"
	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/v3/internal/bridge"
	"github.com/ProtonMail/proton-bridge/v3/internal/services/notifications"
	"github.com/ProtonMail/proton-bridge/v3/internal/vault"
	"github.com/ProtonMail/proton-bridge/v3/pkg/algo"
	"github.com/bradenaw/juniper/iterator"
	"github.com/bradenaw/juniper/xslices"
	"github.com/cucumber/godog"
	"github.com/google/uuid"
)

func (s *scenario) thereExistsAnAccountWithUsernameAndPassword(username, password string) error {
	return s.createUserAccount(username, password, false)
}

func (s *scenario) thereExistsAnAccountWithUsernameAndPasswordWithDisablePrimary(username, password string) error {
	return s.createUserAccount(username, password, true)
}

func (s *scenario) theAccountHasAdditionalAddress(username, address string) error {
	return s.addAdditionalAddressToAccount(username, address, false)
}

func (s *scenario) theAccountHasAdditionalDisabledAddress(username, address string) error {
	return s.addAdditionalAddressToAccount(username, address, true)
}

func (s *scenario) theAccountHasAdditionalAddressWithoutKeys(username, address string) error {
	userID := s.t.getUserByName(username).getUserID()

	// Decrypt the user's encrypted ID for use with quark.
	userDecID, err := s.t.decryptID(userID)
	if err != nil {
		return err
	}

	// Create the user's additional address.
	if _, err := s.t.runQuarkCmd(
		context.Background(),
		"user:create:address",
		"--",
		string(userDecID),
		s.t.getUserByID(userID).getUserPass(),

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
		s.t.getUserByID(userID).addAddress(addr[len(addr)-1].ID, address)

		return nil
	})
}

func (s *scenario) theAccountNoLongerHasAdditionalAddress(username, address string) error {
	addrID := s.t.getUserByName(username).getAddrID(address)

	if err := s.t.withClient(context.Background(), username, func(ctx context.Context, c *proton.Client) error {
		if err := c.DisableAddress(ctx, addrID); err != nil {
			return err
		}

		return c.DeleteAddress(ctx, addrID)
	}); err != nil {
		return err
	}

	return nil
}

func (s *scenario) theAccountHasCustomFolders(username string, count int) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	return s.t.withClient(ctx, username, func(ctx context.Context, client *proton.Client) error {
		for idx := 0; idx < count; idx++ {
			if _, err := client.CreateLabel(ctx, proton.CreateLabelReq{
				Name:  uuid.NewString(),
				Type:  proton.LabelTypeFolder,
				Color: "#f66",
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
				Name:  uuid.NewString(),
				Type:  proton.LabelTypeLabel,
				Color: "#f66",
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
				Name:  wantMailbox.Name,
				Type:  labelType,
				Color: "#f66",
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

	userID := s.t.getUserByName(username).getUserID()
	addrID := s.t.getUserByName(username).getAddrID(address)
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
				Flags:     flagsForMailbox(mailbox),
			},
			Message: message.Build(),
		}
	}))
}

func (s *scenario) theAddressOfAccountHasMessagesInMailbox(address, username string, count int, mailbox string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	userID := s.t.getUserByName(username).getUserID()
	addrID := s.t.getUserByName(username).getAddrID(address)
	mboxID := s.t.getMBoxID(userID, mailbox)

	return s.t.createMessages(ctx, username, addrID, iterator.Collect(iterator.Map(iterator.Counter(count), func(idx int) proton.ImportReq {
		return proton.ImportReq{
			Metadata: proton.ImportMetadata{
				AddressID: addrID,
				LabelIDs:  []string{mboxID},
				Flags:     flagsForMailbox(mailbox),
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

func flagsForMailbox(mailboxName string) proton.MessageFlag {
	if strings.EqualFold(mailboxName, "Sent") {
		return proton.MessageFlagSent
	}

	if strings.EqualFold(mailboxName, "Scheduled") {
		return proton.MessageFlagScheduledSend
	}

	return proton.MessageFlagReceived
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
		return s.t.withAddrKR(ctx, c, username, s.t.getUserByName(username).getAddrID(address), func(_ context.Context, addrKR *crypto.KeyRing) error {
			var changes proton.DraftTemplate

			if wantMessages[0].From != "" {
				return fmt.Errorf("changing From address is not supported")
			}

			changes.Sender = &mail.Address{Address: address}
			changes.MIMEType = rfc822.TextPlain

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

func (s *scenario) drafAtIndexWasMovedToTrashForAddressOfAccount(draftIndex int, address, username string) error {
	draftID, err := s.t.getDraftID(username, draftIndex)
	if err != nil {
		return fmt.Errorf("failed to get draft ID: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	return s.t.withClient(ctx, username, func(ctx context.Context, c *proton.Client) error {
		return s.t.withAddrKR(ctx, c, username, s.t.getUserByName(username).getAddrID(address), func(_ context.Context, _ *crypto.KeyRing) error {
			if err := c.UnlabelMessages(ctx, []string{draftID}, proton.DraftsLabel); err != nil {
				return fmt.Errorf("failed to unlabel draft")
			}
			if err := c.LabelMessages(ctx, []string{draftID}, proton.TrashLabel); err != nil {
				return fmt.Errorf("failed to label draft to trah")
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
		// We need to wait for server to be up or we won't be able to connect. It should only happen once to avoid
		// blocking on multiple Logins.
		s.t.imapServerStarted = true
		s.t.smtpServerStarted = true

		if userID != s.t.getUserByName(username).getUserID() {
			return errors.New("user ID mismatch")
		}

		info, err := s.t.bridge.GetUserInfo(userID)
		if err != nil {
			return err
		}

		s.t.getUserByID(userID).setBridgePass(string(info.BridgePass))
	}

	return nil
}

func (s *scenario) userLogsInWithAliasAddressAndPassword(alias, password string) error {
	userID, err := s.t.bridge.LoginFull(context.Background(), s.t.getUserByAddress(alias).getName(), []byte(password), nil, nil)
	if err != nil {
		s.t.pushError(err)
	} else {
		// We need to wait for server to be up or we won't be able to connect. It should only happen once to avoid
		// blocking on multiple Logins.
		s.t.imapServerStarted = true
		s.t.smtpServerStarted = true

		if userID != s.t.getUserByAddress(alias).getUserID() {
			return errors.New("user ID mismatch")
		}

		info, err := s.t.bridge.GetUserInfo(userID)
		if err != nil {
			return err
		}

		s.t.getUserByID(userID).setBridgePass(string(info.BridgePass))
	}

	return nil
}

func (s *scenario) userLogsOut(username string) error {
	return s.t.bridge.LogoutUser(context.Background(), s.t.getUserByName(username).getUserID())
}

func (s *scenario) userIsDeleted(username string) error {
	return s.t.bridge.DeleteUser(context.Background(), s.t.getUserByName(username).getUserID())
}

func (s *scenario) theAuthOfUserIsRevoked(username string) error {
	return s.t.withClient(context.Background(), username, func(ctx context.Context, client *proton.Client) error {
		return client.AuthRevokeAll(ctx)
	})
}

func (s *scenario) userIsListedAndConnected(username string) error {
	user, err := s.t.bridge.GetUserInfo(s.t.getUserByName(username).getUserID())
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
	user, err := s.t.bridge.GetUserInfo(s.t.getUserByName(username).getUserID())
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
	if _, err := s.t.bridge.QueryUserInfo(username); !errors.Is(err, bridge.ErrNoSuchUser) {
		return errors.New("user listed")
	}

	return nil
}

func (s *scenario) userFinishesSyncing(username string) error {
	return s.bridgeSendsSyncStartedAndFinishedEventsForUser(username)
}

func (s *scenario) userHasTelemetrySetTo(username string, telemetry int) error {
	return s.t.withClientPass(context.Background(), username, s.t.getUserByName(username).userPass, func(ctx context.Context, c *proton.Client) error {
		var req proton.SetTelemetryReq
		req.Telemetry = proton.SettingsBool(telemetry)
		_, err := c.SetUserSettingsTelemetry(ctx, req)
		if err != nil {
			return err
		}
		return nil
	})
}

func (s *scenario) bridgePasswordOfUserIsChangedTo(username, bridgePassword string) error {
	b, err := algo.B64RawDecode([]byte(bridgePassword))
	if err != nil {
		return errors.New("the password is not base64 encoded")
	}

	var setErr error
	if err := s.t.vault.GetUser(
		s.t.getUserByName(username).getUserID(),
		func(user *vault.User) { setErr = user.SetBridgePass(b) },
	); err != nil {
		return err
	}

	return setErr
}

func (s *scenario) bridgePasswordOfUserIsEqualTo(username, bridgePassword string) error {
	userInfo, err := s.t.bridge.QueryUserInfo(username)
	if err != nil {
		return err
	}

	readPassword := string(userInfo.BridgePass)
	if readPassword != bridgePassword {
		return fmt.Errorf("bridge password mismatch, expected '%v', got '%v'", bridgePassword, readPassword)
	}

	return nil
}

func (s *scenario) addAdditionalAddressToAccount(username, address string, disabled bool) error {
	userID := s.t.getUserByName(username).getUserID()

	// Decrypt the user's encrypted ID for use with quark.
	userDecID, err := s.t.decryptID(userID)
	if err != nil {
		return err
	}

	args := []string{
		"--gen-keys", "RSA2048",
	}

	if disabled {
		args = append(args, "--status", "1")
	}

	args = append(args,
		"--",
		string(userDecID),
		s.t.getUserByID(userID).getUserPass(),
		address,
	)

	// Create the user's additional address.
	if _, err := s.t.runQuarkCmd(
		context.Background(),
		"user:create:address",
		args...,
	); err != nil {
		return err
	}

	return s.t.withClient(context.Background(), username, func(ctx context.Context, c *proton.Client) error {
		addr, err := c.GetAddresses(ctx)
		if err != nil {
			return err
		}

		// Set the new address of the user.
		s.t.getUserByID(userID).addAddress(addr[len(addr)-1].ID, address)

		return nil
	})
}

func (s *scenario) createUserAccount(username, password string, disabled bool) error {
	// Create the user and generate its default address (with keys).

	if len(username) == 0 || username[0] == '-' {
		panic("username must be non-empty and not start with minus")
	}

	if len(password) == 0 || password[0] == '-' {
		panic("password must be non-empty and not start with minus")
	}

	args := []string{
		"--name", username,
		"--password", password,
		"--gen-keys", "RSA2048",
	}

	if disabled {
		args = append(args, "--status", "1")
	}

	if _, err := s.t.runQuarkCmd(
		context.Background(),
		"user:create",
		args...,
	); err != nil {
		return err
	}

	return s.t.withClientPass(context.Background(), username, password, func(ctx context.Context, c *proton.Client) error {
		user, err := c.GetUser(ctx)
		if err != nil {
			return err
		}

		// Decrypt the user's encrypted ID for use with quark.
		userDecID, err := s.t.decryptID(user.ID)
		if err != nil {
			return err
		}

		// Upgrade the user to a paid account.
		if _, err := s.t.runQuarkCmd(
			context.Background(),
			"user:create:subscription",
			"--planID", "visionary2022",
			"--",
			string(userDecID),
		); err != nil {
			return err
		}

		addr, err := c.GetAddresses(ctx)
		if err != nil {
			return err
		}

		// Add the test user.
		s.t.addUser(user.ID, username, password)

		// Set the address of the user.
		s.t.getUserByID(user.ID).addAddress(addr[0].ID, addr[0].Email)

		return nil
	})
}

func (s *scenario) accountHasPublicKeyAttachment(account, enabled string) error {
	value := true
	switch {
	case enabled == "enabled":
		value = true
	case enabled == "disabled":
		value = false
	default:
		return errors.New("parameter should either be 'enabled' or 'disabled'")
	}

	return s.t.withClient(context.Background(), account, func(ctx context.Context, c *proton.Client) error {
		_, err := c.SetAttachPublicKey(ctx, proton.SetAttachPublicKeyReq{AttachPublicKey: proton.Bool(value)})
		return err
	})
}

func (s *scenario) accountHasSignExternalMessages(account, enabled string) error {
	value := proton.SignExternalMessagesDisabled
	switch {
	case enabled == "enabled":
		value = proton.SignExternalMessagesEnabled
	case enabled == "disabled":
		value = proton.SignExternalMessagesDisabled
	default:
		return errors.New("parameter should either be 'enabled' or 'disabled'")
	}
	return s.t.withClient(context.Background(), account, func(ctx context.Context, c *proton.Client) error {
		_, err := c.SetSignExternalMessages(ctx, proton.SetSignExternalMessagesReq{Sign: value})
		return err
	})
}

func (s *scenario) accountHasDefaultDraftFormat(account, format string) error {
	value := rfc822.TextPlain
	switch {
	case format == "plain":
		value = rfc822.TextPlain
	case format == "HTML":
		value = rfc822.TextHTML
	default:
		return errors.New("parameter should either be 'plain' or 'HTML'")
	}
	return s.t.withClient(context.Background(), account, func(ctx context.Context, c *proton.Client) error {
		_, err := c.SetDraftMIMEType(ctx, proton.SetDraftMIMETypeReq{MIMEType: value})
		return err
	})
}

func (s *scenario) accountHasDefaultPGPSchema(account, schema string) error {
	value := proton.PGPInlineScheme
	switch {
	case schema == "inline":
		value = proton.PGPInlineScheme
	case schema == "MIME":
		value = proton.PGPMIMEScheme
	default:
		return errors.New("parameter should either be 'inline' or 'MIME'")
	}
	return s.t.withClient(context.Background(), account, func(ctx context.Context, c *proton.Client) error {
		_, err := c.SetDefaultPGPScheme(ctx, proton.SetDefaultPGPSchemeReq{PGPScheme: value})
		return err
	})
}

func (s *scenario) accountMatchesSettings(account string, table *godog.Table) error {
	return s.t.withClient(context.Background(), account, func(ctx context.Context, c *proton.Client) error {
		wantSettings, err := unmarshalTable[MailSettings](table)
		if err != nil {
			return err
		}
		settings, err := c.GetMailSettings(ctx)
		if err != nil {
			return err
		}
		if len(wantSettings) != 1 {
			return errors.New("this step only supports one settings definition at a time")
		}

		return matchSettings(settings, wantSettings[0])
	})
}

func matchSettings(have proton.MailSettings, want MailSettings) error {
	if !IsSub(ToAny(have), ToAny(want)) {
		return fmt.Errorf("missing mailsettings: have %#v, want %#v", have, want)
	}

	return nil
}

func (s *scenario) userRemoteNotificationMetricTest(username string, metricName string) error {
	var metricToTest proton.ObservabilityMetric
	switch strings.ToLower(metricName) {
	case "processed":
		metricToTest = notifications.GenerateProcessedMetric(1)
	case "received":
		metricToTest = notifications.GenerateReceivedMetric(1)
	default:
		return fmt.Errorf("invalid metric name specified")
	}

	// Account for endpoint throttle
	time.Sleep(time.Second * 5)

	return s.t.withClientPass(context.Background(), username, s.t.getUserByName(username).userPass, func(ctx context.Context, c *proton.Client) error {
		batch := proton.ObservabilityBatch{Metrics: []proton.ObservabilityMetric{metricToTest}}
		err := c.SendObservabilityBatch(ctx, batch)
		return err
	})
}
