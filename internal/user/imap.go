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

package user

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/queue"
	"github.com/ProtonMail/gluon/rfc822"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/v2/internal/safe"
	"github.com/ProtonMail/proton-bridge/v2/internal/vault"
	"github.com/ProtonMail/proton-bridge/v2/pkg/message"
	"github.com/bradenaw/juniper/stream"
	"github.com/google/go-cmp/cmp"
	"gitlab.protontech.ch/go/liteapi"
	"golang.org/x/exp/slices"
)

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

// GetMailbox returns information about the mailbox with the given ID.
func (conn *imapConnector) GetMailbox(ctx context.Context, mailboxID imap.MailboxID) (imap.Mailbox, error) {
	mailbox, ok := safe.MapGetRet(conn.apiLabels, string(mailboxID), func(label liteapi.Label) imap.Mailbox {
		return toIMAPMailbox(label, conn.flags, conn.permFlags, conn.attrs)
	})
	if !ok {
		return imap.Mailbox{}, fmt.Errorf("no such mailbox: %s", mailboxID)
	}

	return mailbox, nil
}

// CreateMailbox creates a label with the given name.
func (conn *imapConnector) CreateMailbox(ctx context.Context, name []string) (imap.Mailbox, error) {
	if len(name) < 2 {
		return imap.Mailbox{}, fmt.Errorf("invalid mailbox name %q", name)
	}

	switch name[0] {
	case folderPrefix:
		return conn.createFolder(ctx, name[1:])

	case labelPrefix:
		return conn.createLabel(ctx, name[1:])

	default:
		return imap.Mailbox{}, fmt.Errorf("invalid mailbox name %q", name)
	}
}

func (conn *imapConnector) createLabel(ctx context.Context, name []string) (imap.Mailbox, error) {
	if len(name) != 1 {
		return imap.Mailbox{}, fmt.Errorf("a label cannot have children")
	}

	label, err := conn.client.CreateLabel(ctx, liteapi.CreateLabelReq{
		Name:  name[0],
		Color: "#f66",
		Type:  liteapi.LabelTypeLabel,
	})
	if err != nil {
		return imap.Mailbox{}, err
	}

	return toIMAPMailbox(label, conn.flags, conn.permFlags, conn.attrs), nil
}

func (conn *imapConnector) createFolder(ctx context.Context, name []string) (imap.Mailbox, error) {
	var parentID string

	if len(name) > 1 {
		if ok := conn.apiLabels.GetFunc(func(label liteapi.Label) bool {
			return cmp.Equal(label.Path, name[:len(name)-1])
		}, func(label liteapi.Label) {
			parentID = label.ID
		}); !ok {
			return imap.Mailbox{}, fmt.Errorf("parent folder %q does not exist", name[:len(name)-1])
		}
	}

	label, err := conn.client.CreateLabel(ctx, liteapi.CreateLabelReq{
		Name:     name[len(name)-1],
		Color:    "#f66",
		Type:     liteapi.LabelTypeFolder,
		ParentID: parentID,
	})
	if err != nil {
		return imap.Mailbox{}, err
	}

	return toIMAPMailbox(label, conn.flags, conn.permFlags, conn.attrs), nil
}

// UpdateMailboxName sets the name of the label with the given ID.
func (conn *imapConnector) UpdateMailboxName(ctx context.Context, labelID imap.MailboxID, name []string) error {
	if len(name) < 2 {
		return fmt.Errorf("invalid mailbox name %q", name)
	}

	switch name[0] {
	case folderPrefix:
		return conn.updateFolder(ctx, labelID, name[1:])

	case labelPrefix:
		return conn.updateLabel(ctx, labelID, name[1:])

	default:
		return fmt.Errorf("invalid mailbox name %q", name)
	}
}

func (conn *imapConnector) updateLabel(ctx context.Context, labelID imap.MailboxID, name []string) error {
	if len(name) != 1 {
		return fmt.Errorf("a label cannot have children")
	}

	label, err := conn.client.GetLabel(ctx, string(labelID), liteapi.LabelTypeLabel)
	if err != nil {
		return err
	}

	if _, err := conn.client.UpdateLabel(ctx, label.ID, liteapi.UpdateLabelReq{
		Name:  name[0],
		Color: label.Color,
	}); err != nil {
		return err
	}

	return nil
}

func (conn *imapConnector) updateFolder(ctx context.Context, labelID imap.MailboxID, name []string) error {
	var parentID string

	if len(name) > 1 {
		if ok := conn.apiLabels.GetFunc(func(label liteapi.Label) bool {
			return cmp.Equal(label.Path, name[:len(name)-1])
		}, func(label liteapi.Label) {
			parentID = label.ID
		}); !ok {
			return fmt.Errorf("parent folder %q does not exist", name[:len(name)-1])
		}
	}

	label, err := conn.client.GetLabel(ctx, string(labelID), liteapi.LabelTypeFolder)
	if err != nil {
		return err
	}

	if _, err := conn.client.UpdateLabel(ctx, string(labelID), liteapi.UpdateLabelReq{
		Name:     name[len(name)-1],
		Color:    label.Color,
		ParentID: parentID,
	}); err != nil {
		return err
	}

	return nil
}

// DeleteMailbox deletes the label with the given ID.
func (conn *imapConnector) DeleteMailbox(ctx context.Context, labelID imap.MailboxID) error {
	return conn.client.DeleteLabel(ctx, string(labelID))
}

// GetMessage returns the message with the given ID.
func (conn *imapConnector) GetMessage(ctx context.Context, messageID imap.MessageID) (imap.Message, []imap.MailboxID, error) {
	message, err := conn.client.GetMessage(ctx, string(messageID))
	if err != nil {
		return imap.Message{}, nil, err
	}

	return toIMAPMessage(message.MessageMetadata), mapTo[string, imap.MailboxID](message.LabelIDs), nil
}

// CreateMessage creates a new message on the remote.
func (conn *imapConnector) CreateMessage(
	ctx context.Context,
	mailboxID imap.MailboxID,
	literal []byte,
	flags imap.FlagSet,
	date time.Time,
) (imap.Message, []byte, error) {
	// Compute the hash of the message (to match it against SMTP messages).
	hash, err := getMessageHash(literal)
	if err != nil {
		return imap.Message{}, nil, err
	}

	// Check if we already tried to send this message recently.
	if messageID, ok, err := conn.sendHash.hasEntryWait(ctx, hash, time.Now().Add(90*time.Second)); err != nil {
		return imap.Message{}, nil, fmt.Errorf("failed to check send hash: %w", err)
	} else if ok {
		message, err := conn.client.GetMessage(ctx, messageID)
		if err != nil {
			return imap.Message{}, nil, fmt.Errorf("failed to fetch message: %w", err)
		}

		return toIMAPMessage(message.MessageMetadata), nil, nil
	}

	wantLabelIDs := []string{string(mailboxID)}

	if flags.Contains(imap.FlagFlagged) {
		wantLabelIDs = append(wantLabelIDs, liteapi.StarredLabel)
	}

	var wantFlags liteapi.MessageFlag

	if mailboxID != liteapi.DraftsLabel {
		header, err := rfc822.Parse(literal).ParseHeader()
		if err != nil {
			return imap.Message{}, nil, err
		}

		if header.Has("Received") {
			wantFlags = wantFlags.Add(liteapi.MessageFlagReceived)
		} else {
			wantFlags = wantFlags.Add(liteapi.MessageFlagSent)
		}
	}

	if flags.Contains(imap.FlagAnswered) {
		wantFlags = wantFlags.Add(liteapi.MessageFlagReplied)
	}

	return conn.importMessage(ctx, literal, wantLabelIDs, wantFlags, !flags.Contains(imap.FlagSeen))
}

// AddMessagesToMailbox labels the given messages with the given label ID.
func (conn *imapConnector) AddMessagesToMailbox(ctx context.Context, messageIDs []imap.MessageID, mailboxID imap.MailboxID) error {
	return conn.client.LabelMessages(ctx, mapTo[imap.MessageID, string](messageIDs), string(mailboxID))
}

// RemoveMessagesFromMailbox unlabels the given messages with the given label ID.
func (conn *imapConnector) RemoveMessagesFromMailbox(ctx context.Context, messageIDs []imap.MessageID, mailboxID imap.MailboxID) error {
	return conn.client.UnlabelMessages(ctx, mapTo[imap.MessageID, string](messageIDs), string(mailboxID))
}

// MoveMessages removes the given messages from one label and adds them to the other label.
func (conn *imapConnector) MoveMessages(ctx context.Context, messageIDs []imap.MessageID, labelFromID imap.MailboxID, labelToID imap.MailboxID) error {
	if err := conn.client.LabelMessages(ctx, mapTo[imap.MessageID, string](messageIDs), string(labelToID)); err != nil {
		return fmt.Errorf("labeling messages: %w", err)
	}

	if err := conn.client.UnlabelMessages(ctx, mapTo[imap.MessageID, string](messageIDs), string(labelFromID)); err != nil {
		return fmt.Errorf("unlabeling messages: %w", err)
	}

	return nil
}

// MarkMessagesSeen sets the seen value of the given messages.
func (conn *imapConnector) MarkMessagesSeen(ctx context.Context, messageIDs []imap.MessageID, seen bool) error {
	if seen {
		return conn.client.MarkMessagesRead(ctx, mapTo[imap.MessageID, string](messageIDs)...)
	}

	return conn.client.MarkMessagesUnread(ctx, mapTo[imap.MessageID, string](messageIDs)...)
}

// MarkMessagesFlagged sets the flagged value of the given messages.
func (conn *imapConnector) MarkMessagesFlagged(ctx context.Context, messageIDs []imap.MessageID, flagged bool) error {
	if flagged {
		return conn.client.LabelMessages(ctx, mapTo[imap.MessageID, string](messageIDs), liteapi.StarredLabel)
	}

	return conn.client.UnlabelMessages(ctx, mapTo[imap.MessageID, string](messageIDs), liteapi.StarredLabel)
}

// GetUpdates returns a stream of updates that the gluon server should apply.
// It is recommended that the returned channel is buffered with at least constants.ChannelBufferCount.
func (conn *imapConnector) GetUpdates() <-chan imap.Update {
	updateCh, ok := safe.MapGetRet(conn.updateCh, conn.addrID, func(updateCh *queue.QueuedChannel[imap.Update]) <-chan imap.Update {
		return updateCh.GetChannel()
	})
	if !ok {
		panic(fmt.Sprintf("update channel for %q not found", conn.addrID))
	}

	return updateCh
}

// GetUIDValidity returns the default UID validity for this user.
func (conn *imapConnector) GetUIDValidity() imap.UID {
	return imap.UID(1)
}

// SetUIDValidity sets the default UID validity for this user.
func (conn *imapConnector) SetUIDValidity(uidValidity imap.UID) error {
	return nil
}

// Close the connector will no longer be used and all resources should be closed/released.
func (conn *imapConnector) Close(ctx context.Context) error {
	return nil
}

// IsMailboxVisible returns whether this mailbox should be visible over IMAP.
func (conn *imapConnector) IsMailboxVisible(_ context.Context, mailboxID imap.MailboxID) bool {
	return atomic.LoadUint32(&conn.showAllMail) != 0 || mailboxID != liteapi.AllMailLabel
}

func (conn *imapConnector) importMessage(
	ctx context.Context,
	literal []byte,
	labelIDs []string,
	flags liteapi.MessageFlag,
	unread bool,
) (imap.Message, []byte, error) {
	var full liteapi.FullMessage

	if err := safe.RLockRet(func() error {
		return withAddrKR(conn.apiUser, conn.apiAddrs[conn.addrID], conn.vault.KeyPass(), func(_, addrKR *crypto.KeyRing) error {
			res, err := stream.Collect(ctx, conn.client.ImportMessages(ctx, addrKR, 1, 1, []liteapi.ImportReq{{
				Metadata: liteapi.ImportMetadata{
					AddressID: conn.addrID,
					LabelIDs:  labelIDs,
					Unread:    liteapi.Bool(unread),
					Flags:     flags,
				},
				Message: literal,
			}}...))
			if err != nil {
				return fmt.Errorf("failed to import message: %w", err)
			}

			if full, err = conn.client.GetFullMessage(ctx, res[0].MessageID); err != nil {
				return fmt.Errorf("failed to fetch message: %w", err)
			}

			if literal, err = message.BuildRFC822(addrKR, full.Message, full.AttData, defaultJobOpts()); err != nil {
				return fmt.Errorf("failed to build message: %w", err)
			}

			return nil
		})
	}); err != nil {
		return imap.Message{}, nil, err
	}

	return toIMAPMessage(full.MessageMetadata), literal, nil
}

func toIMAPMessage(message liteapi.MessageMetadata) imap.Message {
	flags := imap.NewFlagSet()

	if !message.Unread {
		flags = flags.Add(imap.FlagSeen)
	}

	if slices.Contains(message.LabelIDs, liteapi.StarredLabel) {
		flags = flags.Add(imap.FlagFlagged)
	}

	if slices.Contains(message.LabelIDs, liteapi.DraftsLabel) {
		flags = flags.Add(imap.FlagDraft)
	}

	return imap.Message{
		ID:    imap.MessageID(message.ID),
		Flags: flags,
		Date:  time.Unix(message.Time, 0),
	}
}

func toIMAPMailbox(label liteapi.Label, flags, permFlags, attrs imap.FlagSet) imap.Mailbox {
	if label.Type == liteapi.LabelTypeLabel {
		label.Path = append([]string{labelPrefix}, label.Path...)
	} else if label.Type == liteapi.LabelTypeFolder {
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
