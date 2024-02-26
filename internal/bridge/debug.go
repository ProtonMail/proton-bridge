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

package bridge

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/rfc822"
	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/proton-bridge/v3/internal/user"
	"github.com/bradenaw/juniper/iterator"
	"github.com/bradenaw/juniper/xslices"
	goimap "github.com/emersion/go-imap"
	goimapclient "github.com/emersion/go-imap/client"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/maps"
)

type CheckClientStateResult struct {
	MissingMessages map[string]map[string]user.DiagMailboxMessage
}

func (c *CheckClientStateResult) AddMissingMessage(userID string, message user.DiagMailboxMessage) {
	v, ok := c.MissingMessages[userID]
	if !ok {
		c.MissingMessages[userID] = map[string]user.DiagMailboxMessage{message.ID: message}
	} else {
		v[message.ID] = message
	}
}

// CheckClientState checks the current IMAP client reported state against the proton server state and reports
// anything that is out of place.
func (bridge *Bridge) CheckClientState(ctx context.Context, checkFlags bool, progressCB func(string)) (CheckClientStateResult, error) {
	bridge.usersLock.RLock()
	defer bridge.usersLock.RUnlock()

	users := maps.Values(bridge.users)

	result := CheckClientStateResult{
		MissingMessages: make(map[string]map[string]user.DiagMailboxMessage),
	}

	for _, usr := range users {
		if progressCB != nil {
			progressCB(fmt.Sprintf("Checking state for user %v", usr.Name()))
		}
		log := logrus.WithFields(logrus.Fields{
			"pkg":  "bridge/debug",
			"user": usr.Name(),
			"diag": "state-check",
		})
		log.Debug("Retrieving all server metadata")
		meta, err := usr.GetDiagnosticMetadata(ctx)
		if err != nil {
			return result, err
		}

		success := true

		if len(meta.Metadata) != len(meta.MessageIDs) {
			log.Errorf("Metadata (%v) and message(%v) list sizes do not match", len(meta.Metadata), len(meta.MessageIDs))
		}

		log.Debug("Building state")
		state, err := meta.BuildMailboxToMessageMap(ctx, usr)
		if err != nil {
			log.WithError(err).Error("Failed to build state")
			return result, err
		}

		info, err := bridge.GetUserInfo(usr.ID())
		if err != nil {
			log.WithError(err).Error("Failed to get user info")
			return result, err
		}

		addr := fmt.Sprintf("127.0.0.1:%v", bridge.GetIMAPPort())

		for account, mboxMap := range state {
			if progressCB != nil {
				progressCB(fmt.Sprintf("Checking state for user %v's account '%v'", usr.Name(), account))
			}
			if err := func(account string, mboxMap user.AccountMailboxMap) error {
				client, err := goimapclient.Dial(addr)
				if err != nil {
					log.WithError(err).Error("Failed to connect to imap client")
					return err
				}

				defer func() {
					_ = client.Logout()
				}()

				if err := client.Login(account, string(info.BridgePass)); err != nil {
					return fmt.Errorf("failed to login for user %v:%w", usr.Name(), err)
				}

				log := log.WithField("account", account)
				for mboxName, messageList := range mboxMap {
					log := log.WithField("mbox", mboxName)
					status, err := client.Select(mboxName, true)
					if err != nil {
						log.WithError(err).Errorf("Failed to select mailbox %v", messageList)
						return fmt.Errorf("failed to select '%v':%w", mboxName, err)
					}

					log.Debug("Checking message count")

					if int(status.Messages) != len(messageList) {
						success = false
						log.Errorf("Message count doesn't match, got '%v' expected '%v'", status.Messages, len(messageList))
					}

					ids, err := clientGetMessageIDs(client, mboxName)
					if err != nil {
						return fmt.Errorf("failed to get message ids for mbox '%v': %w", mboxName, err)
					}

					for _, msg := range messageList {
						imapFlags, ok := ids[msg.ID]
						if !ok {
							if meta.FailedMessageIDs.Contains(msg.ID) {
								log.Warningf("Missing message '%v', but it is part of failed message set", msg.ID)
							} else {
								log.Errorf("Missing message '%v'", msg.ID)
							}

							result.AddMissingMessage(msg.UserID, msg)
							continue
						}

						if checkFlags {
							if !imapFlags.Equals(msg.Flags) {
								log.Errorf("Message '%v' flags do mot match, got=%v, expected=%v",
									msg.ID,
									imapFlags.ToSlice(),
									msg.Flags.ToSlice(),
								)
							}
						}
					}
				}

				if !success {
					log.Errorf("State does not match")
				} else {
					log.Info("State matches")
				}

				return nil
			}(account, mboxMap); err != nil {
				return result, err
			}
		}

		// Check for orphaned messages (only present in All Mail)
		if progressCB != nil {
			progressCB(fmt.Sprintf("Checking user %v for orphans", usr.Name()))
		}
		log.Debugf("Checking for orphans")

		for _, m := range meta.Metadata {
			filteredLabels := xslices.Filter(m.LabelIDs, func(t string) bool {
				switch t {
				case proton.AllMailLabel:
					return false
				case proton.AllSentLabel:
					return false
				case proton.AllDraftsLabel:
					return false
				case proton.OutboxLabel:
					return false
				default:
					return true
				}
			})

			if len(filteredLabels) == 0 {
				log.Warnf("Message %v is only present in All Mail (Subject=%v)", m.ID, m.Subject)
			}
		}
	}

	return result, nil
}

func (bridge *Bridge) DebugDownloadFailedMessages(
	ctx context.Context,
	result CheckClientStateResult,
	exportPath string,
	progressCB func(string, int, int),
) error {
	bridge.usersLock.RLock()
	defer bridge.usersLock.RUnlock()

	for userID, messages := range result.MissingMessages {
		usr, ok := bridge.users[userID]
		if !ok {
			return fmt.Errorf("failed to find user with id %v", userID)
		}

		userDir := filepath.Join(exportPath, userID)
		if err := os.MkdirAll(userDir, 0o700); err != nil {
			return fmt.Errorf("failed to create directory '%v': %w", userDir, err)
		}

		if err := usr.DebugDownloadMessages(ctx, userDir, messages, progressCB); err != nil {
			return err
		}
	}

	return nil
}

func clientGetMessageIDs(client *goimapclient.Client, mailbox string) (map[string]imap.FlagSet, error) {
	status, err := client.Select(mailbox, true)
	if err != nil {
		return nil, err
	}

	if status.Messages == 0 {
		return nil, nil
	}

	resCh := make(chan *goimap.Message)

	section, err := goimap.ParseBodySectionName("BODY[HEADER]")
	if err != nil {
		return nil, err
	}

	fetchItems := []goimap.FetchItem{"BODY[HEADER]", goimap.FetchFlags}

	seq, err := goimap.ParseSeqSet("1:*")
	if err != nil {
		return nil, err
	}

	go func() {
		if err := client.Fetch(
			seq,
			fetchItems,
			resCh,
		); err != nil {
			panic(err)
		}
	}()

	messages := iterator.Collect(iterator.Chan(resCh))

	ids := make(map[string]imap.FlagSet, len(messages))

	for i, m := range messages {
		literal, err := io.ReadAll(m.GetBody(section))
		if err != nil {
			return nil, err
		}

		header, err := rfc822.NewHeader(literal)
		if err != nil {
			return nil, fmt.Errorf("failed to parse header for msg %v: %w", i, err)
		}

		internalID, ok := header.GetChecked("X-Pm-Internal-Id")
		if !ok {
			logrus.WithField("pkg", "bridge/debug").Errorf("Message %v does not have internal id", internalID)
			continue
		}

		messageFlags := imap.NewFlagSet(m.Flags...)

		// Recent and Deleted are not part of the proton flag set.
		messageFlags.RemoveFromSelf("\\Recent")
		messageFlags.RemoveFromSelf("\\Deleted")

		ids[internalID] = messageFlags
	}

	return ids, nil
}
