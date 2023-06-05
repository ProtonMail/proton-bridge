// Copyright (c) 2023 Proton AG
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

package user

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/mail"
	"sync/atomic"
	"time"

	"github.com/ProtonMail/gluon/connector"
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/rfc822"
	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/v3/internal/safe"
	"github.com/ProtonMail/proton-bridge/v3/internal/vault"
	"github.com/ProtonMail/proton-bridge/v3/pkg/message"
	"github.com/ProtonMail/proton-bridge/v3/pkg/message/parser"
	"github.com/bradenaw/juniper/stream"
	"github.com/bradenaw/juniper/xslices"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
)

// Verify that *imapConnector implements connector.Connector.
var _ connector.Connector = (*imapConnector)(nil)

var (
	defaultFlags          = imap.NewFlagSet(imap.FlagSeen, imap.FlagFlagged, imap.FlagDeleted) // nolint:gochecknoglobals
	defaultPermanentFlags = imap.NewFlagSet(imap.FlagSeen, imap.FlagFlagged, imap.FlagDeleted) // nolint:gochecknoglobals
	defaultAttributes     = imap.NewFlagSet()                                                  // nolint:gochecknoglobals
)

const (
	folderPrefix = "Folders"
	labelPrefix  = "Labels"
)

type imapConnector struct {
	*User

	addrID string

	flags, permFlags, attrs imap.FlagSet
}

func newIMAPConnector(user *User, addrID string) *imapConnector {
	return &imapConnector{
		User: user,

		addrID: addrID,

		flags:     defaultFlags,
		permFlags: defaultPermanentFlags,
		attrs:     defaultAttributes,
	}
}

// Authorize returns whether the given username/password combination are valid for this connector.
func (conn *imapConnector) Authorize(username string, password []byte) bool {
	addrID, err := conn.CheckAuth(username, password)
	if err != nil {
		return false
	}

	if conn.vault.AddressMode() == vault.SplitMode && addrID != conn.addrID {
		return false
	}

	return true
}

// CreateMailbox creates a label with the given name.
func (conn *imapConnector) CreateMailbox(ctx context.Context, name []string) (imap.Mailbox, error) {
	defer conn.goPollAPIEvents(false)

	if len(name) < 2 {
		return imap.Mailbox{}, fmt.Errorf("invalid mailbox name %q: %w", name, connector.ErrOperationNotAllowed)
	}

	switch name[0] {
	case folderPrefix:
		return conn.createFolder(ctx, name[1:])

	case labelPrefix:
		return conn.createLabel(ctx, name[1:])

	default:
		return imap.Mailbox{}, fmt.Errorf("invalid mailbox name %q: %w", name, connector.ErrOperationNotAllowed)
	}
}

func (conn *imapConnector) createLabel(ctx context.Context, name []string) (imap.Mailbox, error) {
	if len(name) != 1 {
		return imap.Mailbox{}, fmt.Errorf("a label cannot have children: %w", connector.ErrOperationNotAllowed)
	}

	return safe.LockRetErr(func() (imap.Mailbox, error) {
		label, err := conn.client.CreateLabel(ctx, proton.CreateLabelReq{
			Name:  name[0],
			Color: "#f66",
			Type:  proton.LabelTypeLabel,
		})
		if err != nil {
			return imap.Mailbox{}, err
		}

		conn.apiLabels[label.ID] = label

		return toIMAPMailbox(label, conn.flags, conn.permFlags, conn.attrs), nil
	}, conn.apiLabelsLock)
}

func (conn *imapConnector) createFolder(ctx context.Context, name []string) (imap.Mailbox, error) {
	return safe.LockRetErr(func() (imap.Mailbox, error) {
		var parentID string

		if len(name) > 1 {
			for _, label := range conn.apiLabels {
				if !slices.Equal(label.Path, name[:len(name)-1]) {
					continue
				}

				parentID = label.ID

				break
			}

			if parentID == "" {
				return imap.Mailbox{}, fmt.Errorf("parent folder %q does not exist: %w", name[:len(name)-1], connector.ErrOperationNotAllowed)
			}
		}

		label, err := conn.client.CreateLabel(ctx, proton.CreateLabelReq{
			Name:     name[len(name)-1],
			Color:    "#f66",
			Type:     proton.LabelTypeFolder,
			ParentID: parentID,
		})
		if err != nil {
			return imap.Mailbox{}, err
		}

		// Add label to list so subsequent sub folder create requests work correct.
		conn.apiLabels[label.ID] = label

		return toIMAPMailbox(label, conn.flags, conn.permFlags, conn.attrs), nil
	}, conn.apiLabelsLock)
}

// UpdateMailboxName sets the name of the label with the given ID.
func (conn *imapConnector) UpdateMailboxName(ctx context.Context, labelID imap.MailboxID, name []string) error {
	return safe.LockRet(func() error {
		defer conn.goPollAPIEvents(false)

		if len(name) < 2 {
			return fmt.Errorf("invalid mailbox name %q: %w", name, connector.ErrOperationNotAllowed)
		}

		switch name[0] {
		case folderPrefix:
			return conn.updateFolder(ctx, labelID, name[1:])

		case labelPrefix:
			return conn.updateLabel(ctx, labelID, name[1:])

		default:
			return fmt.Errorf("invalid mailbox name %q: %w", name, connector.ErrOperationNotAllowed)
		}
	}, conn.apiLabelsLock)
}

func (conn *imapConnector) updateLabel(ctx context.Context, labelID imap.MailboxID, name []string) error {
	if len(name) != 1 {
		return fmt.Errorf("a label cannot have children: %w", connector.ErrOperationNotAllowed)
	}

	label, err := conn.client.GetLabel(ctx, string(labelID), proton.LabelTypeLabel)
	if err != nil {
		return err
	}

	update, err := conn.client.UpdateLabel(ctx, label.ID, proton.UpdateLabelReq{
		Name:  name[0],
		Color: label.Color,
	})
	if err != nil {
		return err
	}

	conn.apiLabels[label.ID] = update

	return nil
}

func (conn *imapConnector) updateFolder(ctx context.Context, labelID imap.MailboxID, name []string) error {
	var parentID string

	if len(name) > 1 {
		for _, label := range conn.apiLabels {
			if !slices.Equal(label.Path, name[:len(name)-1]) {
				continue
			}

			parentID = label.ID

			break
		}

		if parentID == "" {
			return fmt.Errorf("parent folder %q does not exist: %w", name[:len(name)-1], connector.ErrOperationNotAllowed)
		}
	}

	label, err := conn.client.GetLabel(ctx, string(labelID), proton.LabelTypeFolder)
	if err != nil {
		return err
	}

	update, err := conn.client.UpdateLabel(ctx, string(labelID), proton.UpdateLabelReq{
		Name:     name[len(name)-1],
		Color:    label.Color,
		ParentID: parentID,
	})
	if err != nil {
		return err
	}

	conn.apiLabels[label.ID] = update

	return nil
}

// DeleteMailbox deletes the label with the given ID.
func (conn *imapConnector) DeleteMailbox(ctx context.Context, labelID imap.MailboxID) error {
	return safe.LockRet(func() error {
		defer conn.goPollAPIEvents(false)

		if err := conn.client.DeleteLabel(ctx, string(labelID)); err != nil {
			return err
		}

		delete(conn.apiLabels, string(labelID))

		return nil
	}, conn.apiLabelsLock)
}

// CreateMessage creates a new message on the remote.
func (conn *imapConnector) CreateMessage(
	ctx context.Context,
	mailboxID imap.MailboxID,
	literal []byte,
	flags imap.FlagSet,
	_ time.Time,
) (imap.Message, []byte, error) {
	defer conn.goPollAPIEvents(false)

	if mailboxID == proton.AllMailLabel {
		return imap.Message{}, nil, connector.ErrOperationNotAllowed
	}

	// Compute the hash of the message (to match it against SMTP messages).
	hash, err := getMessageHash(literal)
	if err != nil {
		return imap.Message{}, nil, err
	}

	// Check if we already tried to send this message recently.
	if messageID, ok, err := conn.sendHash.hasEntryWait(ctx, hash, time.Now().Add(90*time.Second)); err != nil {
		return imap.Message{}, nil, fmt.Errorf("failed to check send hash: %w", err)
	} else if ok {
		conn.log.WithField("messageID", messageID).Warn("Message already sent")

		// Query the server-side message.
		full, err := conn.client.GetFullMessage(ctx, messageID, newProtonAPIScheduler(conn.panicHandler), proton.NewDefaultAttachmentAllocator())
		if err != nil {
			return imap.Message{}, nil, fmt.Errorf("failed to fetch message: %w", err)
		}

		// Build the message as it is on the server.
		if err := safe.RLockRet(func() error {
			return withAddrKR(conn.apiUser, conn.apiAddrs[full.AddressID], conn.vault.KeyPass(), func(_, addrKR *crypto.KeyRing) error {
				var err error

				if literal, err = message.BuildRFC822(addrKR, full.Message, full.AttData, defaultJobOpts()); err != nil {
					return err
				}

				return nil
			})
		}, conn.apiUserLock, conn.apiAddrsLock); err != nil {
			return imap.Message{}, nil, fmt.Errorf("failed to build message: %w", err)
		}

		return toIMAPMessage(full.MessageMetadata), literal, nil
	}

	wantLabelIDs := []string{string(mailboxID)}

	if flags.Contains(imap.FlagFlagged) {
		wantLabelIDs = append(wantLabelIDs, proton.StarredLabel)
	}

	var wantFlags proton.MessageFlag

	unread := !flags.Contains(imap.FlagSeen)

	if mailboxID != proton.DraftsLabel {
		header, err := rfc822.Parse(literal).ParseHeader()
		if err != nil {
			return imap.Message{}, nil, err
		}

		switch {
		case mailboxID == proton.InboxLabel:
			wantFlags = wantFlags.Add(proton.MessageFlagReceived)

		case mailboxID == proton.SentLabel:
			wantFlags = wantFlags.Add(proton.MessageFlagSent)

		case header.Has("Received"):
			wantFlags = wantFlags.Add(proton.MessageFlagReceived)

		default:
			wantFlags = wantFlags.Add(proton.MessageFlagSent)
		}
	} else {
		unread = false
	}

	if flags.Contains(imap.FlagAnswered) {
		wantFlags = wantFlags.Add(proton.MessageFlagReplied)
	}

	msg, literal, err := conn.importMessage(ctx, literal, wantLabelIDs, wantFlags, unread)
	if err != nil && errors.Is(err, proton.ErrImportSizeExceeded) {
		// Remap error so that Gluon does not put this message in the recovery mailbox.
		err = fmt.Errorf("%v: %w", err, connector.ErrMessageSizeExceedsLimits)
	}

	if apiErr := new(proton.APIError); errors.As(err, &apiErr) {
		logrus.WithError(apiErr).WithField("Details", apiErr.DetailsToString()).Error("Failed to import message")
	} else {
		logrus.WithError(err).Error("Failed to import message")
	}

	return msg, literal, err
}

func (conn *imapConnector) GetMessageLiteral(ctx context.Context, id imap.MessageID) ([]byte, error) {
	msg, err := conn.client.GetFullMessage(ctx, string(id), newProtonAPIScheduler(conn.panicHandler), proton.NewDefaultAttachmentAllocator())
	if err != nil {
		return nil, err
	}

	return safe.RLockRetErr(func() ([]byte, error) {
		var literal []byte
		err := withAddrKR(conn.apiUser, conn.apiAddrs[msg.AddressID], conn.vault.KeyPass(), func(_, addrKR *crypto.KeyRing) error {
			l, buildErr := message.BuildRFC822(addrKR, msg.Message, msg.AttData, defaultJobOpts())
			if buildErr != nil {
				return buildErr
			}

			literal = l

			return nil
		})

		return literal, err
	}, conn.apiUserLock, conn.apiAddrsLock)
}

// AddMessagesToMailbox labels the given messages with the given label ID.
func (conn *imapConnector) AddMessagesToMailbox(ctx context.Context, messageIDs []imap.MessageID, mailboxID imap.MailboxID) error {
	defer conn.goPollAPIEvents(false)

	if isAllMailOrScheduled(mailboxID) {
		return connector.ErrOperationNotAllowed
	}

	return conn.client.LabelMessages(ctx, mapTo[imap.MessageID, string](messageIDs), string(mailboxID))
}

// RemoveMessagesFromMailbox unlabels the given messages with the given label ID.
func (conn *imapConnector) RemoveMessagesFromMailbox(ctx context.Context, messageIDs []imap.MessageID, mailboxID imap.MailboxID) error {
	defer conn.goPollAPIEvents(false)

	if isAllMailOrScheduled(mailboxID) {
		return connector.ErrOperationNotAllowed
	}

	if err := conn.client.UnlabelMessages(ctx, mapTo[imap.MessageID, string](messageIDs), string(mailboxID)); err != nil {
		return err
	}

	if mailboxID == proton.TrashLabel || mailboxID == proton.DraftsLabel {
		var metadata []proton.MessageMetadata

		// There's currently no limit on how many IDs we can filter on,
		// but to be nice to API, let's chunk it by 150.
		for _, messageIDs := range xslices.Chunk(messageIDs, 150) {
			m, err := conn.client.GetMessageMetadata(ctx, proton.MessageFilter{
				ID: mapTo[imap.MessageID, string](messageIDs),
			})
			if err != nil {
				return err
			}

			// If a message is not preset in any other label other than AllMail, AllDrafts and AllSent, it can be
			// permanently deleted.
			m = xslices.Filter(m, func(m proton.MessageMetadata) bool {
				labelsThatMatter := xslices.Filter(m.LabelIDs, func(id string) bool {
					return id != proton.AllDraftsLabel && id != proton.AllMailLabel && id != proton.AllSentLabel
				})
				return len(labelsThatMatter) == 0
			})

			metadata = append(metadata, m...)
		}

		if err := conn.client.DeleteMessage(ctx, xslices.Map(metadata, func(m proton.MessageMetadata) string {
			return m.ID
		})...); err != nil {
			return err
		}
	}

	return nil
}

// MoveMessages removes the given messages from one label and adds them to the other label.
func (conn *imapConnector) MoveMessages(ctx context.Context, messageIDs []imap.MessageID, labelFromID imap.MailboxID, labelToID imap.MailboxID) (bool, error) {
	defer conn.goPollAPIEvents(false)

	if (labelFromID == proton.InboxLabel && labelToID == proton.SentLabel) ||
		(labelFromID == proton.SentLabel && labelToID == proton.InboxLabel) ||
		isAllMailOrScheduled(labelFromID) ||
		isAllMailOrScheduled(labelToID) {
		return false, connector.ErrOperationNotAllowed
	}

	shouldExpungeOldLocation := func() bool {
		conn.apiLabelsLock.RLock()
		defer conn.apiLabelsLock.RUnlock()

		var result bool

		if v, ok := conn.apiLabels[string(labelFromID)]; ok && v.Type == proton.LabelTypeLabel {
			result = true
		}

		if v, ok := conn.apiLabels[string(labelToID)]; ok && (v.Type == proton.LabelTypeFolder || v.Type == proton.LabelTypeSystem) {
			result = true
		}

		return result
	}()

	if err := conn.client.LabelMessages(ctx, mapTo[imap.MessageID, string](messageIDs), string(labelToID)); err != nil {
		return false, fmt.Errorf("labeling messages: %w", err)
	}

	if shouldExpungeOldLocation {
		if err := conn.client.UnlabelMessages(ctx, mapTo[imap.MessageID, string](messageIDs), string(labelFromID)); err != nil {
			return false, fmt.Errorf("unlabeling messages: %w", err)
		}
	}

	return shouldExpungeOldLocation, nil
}

// MarkMessagesSeen sets the seen value of the given messages.
func (conn *imapConnector) MarkMessagesSeen(ctx context.Context, messageIDs []imap.MessageID, seen bool) error {
	defer conn.goPollAPIEvents(false)

	if seen {
		return conn.client.MarkMessagesRead(ctx, mapTo[imap.MessageID, string](messageIDs)...)
	}

	return conn.client.MarkMessagesUnread(ctx, mapTo[imap.MessageID, string](messageIDs)...)
}

// MarkMessagesFlagged sets the flagged value of the given messages.
func (conn *imapConnector) MarkMessagesFlagged(ctx context.Context, messageIDs []imap.MessageID, flagged bool) error {
	defer conn.goPollAPIEvents(false)

	if flagged {
		return conn.client.LabelMessages(ctx, mapTo[imap.MessageID, string](messageIDs), proton.StarredLabel)
	}

	return conn.client.UnlabelMessages(ctx, mapTo[imap.MessageID, string](messageIDs), proton.StarredLabel)
}

// GetUpdates returns a stream of updates that the gluon server should apply.
// It is recommended that the returned channel is buffered with at least constants.ChannelBufferCount.
func (conn *imapConnector) GetUpdates() <-chan imap.Update {
	return safe.RLockRet(func() <-chan imap.Update {
		return conn.updateCh[conn.addrID].GetChannel()
	}, conn.updateChLock)
}

// GetMailboxVisibility returns the visibility of a mailbox over IMAP.
func (conn *imapConnector) GetMailboxVisibility(_ context.Context, mailboxID imap.MailboxID) imap.MailboxVisibility {
	switch mailboxID {
	case proton.AllMailLabel:
		if atomic.LoadUint32(&conn.showAllMail) != 0 {
			return imap.Visible
		}
		return imap.Hidden

	case proton.AllScheduledLabel:
		return imap.HiddenIfEmpty
	default:
		return imap.Visible
	}
}

// Close the connector will no longer be used and all resources should be closed/released.
func (conn *imapConnector) Close(_ context.Context) error {
	return nil
}

func (conn *imapConnector) importMessage(
	ctx context.Context,
	literal []byte,
	labelIDs []string,
	flags proton.MessageFlag,
	unread bool,
) (imap.Message, []byte, error) {
	var full proton.FullMessage

	if err := safe.RLockRet(func() error {
		return withAddrKR(conn.apiUser, conn.apiAddrs[conn.addrID], conn.vault.KeyPass(), func(_, addrKR *crypto.KeyRing) error {
			var messageID string

			if slices.Contains(labelIDs, proton.DraftsLabel) {
				msg, err := conn.createDraft(ctx, literal, addrKR, conn.apiAddrs[conn.addrID])
				if err != nil {
					return fmt.Errorf("failed to create draft: %w", err)
				}

				// apply labels

				messageID = msg.ID
			} else {
				str, err := conn.client.ImportMessages(ctx, addrKR, 1, 1, []proton.ImportReq{{
					Metadata: proton.ImportMetadata{
						AddressID: conn.addrID,
						LabelIDs:  labelIDs,
						Unread:    proton.Bool(unread),
						Flags:     flags,
					},
					Message: literal,
				}}...)
				if err != nil {
					return fmt.Errorf("failed to prepare message for import: %w", err)
				}

				res, err := stream.Collect(ctx, str)
				if err != nil {
					return fmt.Errorf("failed to import message: %w", err)
				}

				messageID = res[0].MessageID
			}

			var err error

			if full, err = conn.client.GetFullMessage(ctx, messageID, newProtonAPIScheduler(conn.panicHandler), proton.NewDefaultAttachmentAllocator()); err != nil {
				return fmt.Errorf("failed to fetch message: %w", err)
			}

			if literal, err = message.BuildRFC822(addrKR, full.Message, full.AttData, defaultJobOpts()); err != nil {
				return fmt.Errorf("failed to build message: %w", err)
			}

			return nil
		})
	}, conn.apiUserLock, conn.apiAddrsLock); err != nil {
		return imap.Message{}, nil, err
	}

	return toIMAPMessage(full.MessageMetadata), literal, nil
}

func toIMAPMessage(message proton.MessageMetadata) imap.Message {
	flags := imap.NewFlagSet()

	if !message.Unread {
		flags = flags.Add(imap.FlagSeen)
	}

	if slices.Contains(message.LabelIDs, proton.StarredLabel) {
		flags = flags.Add(imap.FlagFlagged)
	}

	if slices.Contains(message.LabelIDs, proton.DraftsLabel) {
		flags = flags.Add(imap.FlagDraft)
	}

	var date time.Time

	if message.Time > 0 {
		date = time.Unix(message.Time, 0)
	} else {
		date = time.Now()
	}

	return imap.Message{
		ID:    imap.MessageID(message.ID),
		Flags: flags,
		Date:  date,
	}
}

func (conn *imapConnector) createDraft(ctx context.Context, literal []byte, addrKR *crypto.KeyRing, sender proton.Address) (proton.Message, error) {
	// Create a new message parser from the reader.
	parser, err := parser.New(bytes.NewReader(literal))
	if err != nil {
		return proton.Message{}, fmt.Errorf("failed to create parser: %w", err)
	}

	message, err := message.ParseWithParser(parser, true)
	if err != nil {
		return proton.Message{}, fmt.Errorf("failed to parse message: %w", err)
	}

	decBody := string(message.PlainBody)
	if message.RichBody != "" {
		decBody = string(message.RichBody)
	}

	draft, err := conn.client.CreateDraft(ctx, addrKR, proton.CreateDraftReq{
		Message: proton.DraftTemplate{
			Subject:  message.Subject,
			Body:     decBody,
			MIMEType: message.MIMEType,

			Sender:  &mail.Address{Name: sender.DisplayName, Address: sender.Email},
			ToList:  message.ToList,
			CCList:  message.CCList,
			BCCList: message.BCCList,

			ExternalID: message.ExternalID,
		},
	})

	if err != nil {
		return proton.Message{}, fmt.Errorf("failed to create draft: %w", err)
	}

	for _, att := range message.Attachments {
		disposition := proton.AttachmentDisposition
		if att.Disposition == "inline" && att.ContentID != "" {
			disposition = proton.InlineDisposition
		}

		if _, err := conn.client.UploadAttachment(ctx, addrKR, proton.CreateAttachmentReq{
			MessageID:   draft.ID,
			Filename:    att.Name,
			MIMEType:    rfc822.MIMEType(att.MIMEType),
			Disposition: disposition,
			ContentID:   att.ContentID,
			Body:        att.Data,
		}); err != nil {
			return proton.Message{}, fmt.Errorf("failed to add attachment to draft: %w", err)
		}
	}

	return draft, nil
}

func toIMAPMailbox(label proton.Label, flags, permFlags, attrs imap.FlagSet) imap.Mailbox {
	if label.Type == proton.LabelTypeLabel {
		label.Path = append([]string{labelPrefix}, label.Path...)
	} else if label.Type == proton.LabelTypeFolder {
		label.Path = append([]string{folderPrefix}, label.Path...)
	}

	return imap.Mailbox{
		ID:             imap.MailboxID(label.ID),
		Name:           label.Path,
		Flags:          flags,
		PermanentFlags: permFlags,
		Attributes:     attrs,
	}
}

func isAllMailOrScheduled(mailboxID imap.MailboxID) bool {
	return (mailboxID == proton.AllMailLabel) || (mailboxID == proton.AllScheduledLabel)
}
