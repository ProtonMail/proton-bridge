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
	"time"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/queue"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/v2/internal/safe"
	"github.com/ProtonMail/proton-bridge/v2/internal/vault"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
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
	addrID, err := conn.checkAuth(username, password)
	if err != nil {
		return false
	}

	if conn.vault.AddressMode() == vault.SplitMode && addrID != conn.addrID {
		return false
	}

	return true
}

// GetLabel returns information about the label with the given ID.
func (conn *imapConnector) GetLabel(ctx context.Context, labelID imap.LabelID) (imap.Mailbox, error) {
	label, err := conn.client.GetLabel(ctx, string(labelID), liteapi.LabelTypeLabel, liteapi.LabelTypeFolder)
	if err != nil {
		return imap.Mailbox{}, err
	}

	var name []string

	switch label.Type {
	case liteapi.LabelTypeLabel:
		name = []string{labelPrefix, label.Name}

	case liteapi.LabelTypeFolder:
		name = []string{folderPrefix, label.Name}

	case liteapi.LabelTypeContactGroup:
		fallthrough
	case liteapi.LabelTypeSystem:
		fallthrough
	default:
		name = []string{label.Name}
	}

	return imap.Mailbox{
		ID:             imap.LabelID(label.ID),
		Name:           name,
		Flags:          conn.flags,
		PermanentFlags: conn.permFlags,
		Attributes:     conn.attrs,
	}, nil
}

// CreateLabel creates a label with the given name.
func (conn *imapConnector) CreateLabel(ctx context.Context, name []string) (imap.Mailbox, error) {
	if len(name) != 2 {
		panic("subfolders are unsupported")
	}

	var labelType liteapi.LabelType

	if name[0] == folderPrefix {
		labelType = liteapi.LabelTypeFolder
	} else {
		labelType = liteapi.LabelTypeLabel
	}

	label, err := conn.client.CreateLabel(ctx, liteapi.CreateLabelReq{
		Name:  name[1:][0],
		Color: "#f66",
		Type:  labelType,
	})
	if err != nil {
		return imap.Mailbox{}, err
	}

	return imap.Mailbox{
		ID:             imap.LabelID(label.ID),
		Name:           name,
		Flags:          conn.flags,
		PermanentFlags: conn.permFlags,
		Attributes:     conn.attrs,
	}, nil
}

// UpdateLabel sets the name of the label with the given ID.
func (conn *imapConnector) UpdateLabel(ctx context.Context, labelID imap.LabelID, newName []string) error {
	if len(newName) != 2 {
		panic("subfolders are unsupported")
	}

	label, err := conn.client.GetLabel(ctx, string(labelID), liteapi.LabelTypeLabel, liteapi.LabelTypeFolder)
	if err != nil {
		return err
	}

	switch label.Type {
	case liteapi.LabelTypeFolder:
		if newName[0] != folderPrefix {
			return fmt.Errorf("cannot rename folder to label")
		}

	case liteapi.LabelTypeLabel:
		if newName[0] != labelPrefix {
			return fmt.Errorf("cannot rename label to folder")
		}

	case liteapi.LabelTypeSystem:
		return fmt.Errorf("cannot rename system label %q", label.Name)

	case liteapi.LabelTypeContactGroup:
		return fmt.Errorf("cannot rename contact group label %q", label.Name)
	}

	if _, err := conn.client.UpdateLabel(ctx, label.ID, liteapi.UpdateLabelReq{
		Name:  newName[1:][0],
		Color: label.Color,
	}); err != nil {
		return err
	}

	return nil
}

// DeleteLabel deletes the label with the given ID.
func (conn *imapConnector) DeleteLabel(ctx context.Context, labelID imap.LabelID) error {
	return conn.client.DeleteLabel(ctx, string(labelID))
}

// GetMessage returns the message with the given ID.
func (conn *imapConnector) GetMessage(ctx context.Context, messageID imap.MessageID) (imap.Message, []imap.LabelID, error) {
	message, err := conn.client.GetMessage(ctx, string(messageID))
	if err != nil {
		return imap.Message{}, nil, err
	}

	flags := imap.NewFlagSet()

	if !message.Unread {
		flags = flags.Add(imap.FlagSeen)
	}

	if slices.Contains(message.LabelIDs, liteapi.StarredLabel) {
		flags = flags.Add(imap.FlagFlagged)
	}

	return imap.Message{
		ID:    imap.MessageID(message.ID),
		Flags: flags,
		Date:  time.Unix(message.Time, 0),
	}, mapTo[string, imap.LabelID](message.LabelIDs), nil
}

// CreateMessage creates a new message on the remote.
func (conn *imapConnector) CreateMessage(
	ctx context.Context,
	labelID imap.LabelID,
	literal []byte,
	flags imap.FlagSet,
	date time.Time,
) (imap.Message, []byte, error) {
	var msgFlags liteapi.MessageFlag

	switch labelID {
	case liteapi.SentLabel:
		msgFlags |= liteapi.MessageFlagSent

	default:
		msgFlags |= liteapi.MessageFlagReceived
	}

	var importResult liteapi.ImportRes
	if err := conn.withAddrKR(conn.addrID, func(ring *crypto.KeyRing) error {
		requestName := uuid.NewString()

		importReq := []liteapi.ImportReq{{
			Name: requestName,
			Metadata: liteapi.ImportMetadata{
				AddressID: conn.addrID,
				LabelIDs:  []string{string(labelID)},
				Flags:     msgFlags,
			},
			Message: literal,
		}}

		r, err := conn.client.ImportMessages(ctx, ring, importReq)
		if err != nil {
			return err
		}

		importResult = r[requestName]

		return nil
	}); err != nil {
		return imap.Message{}, nil, err
	}

	if importResult.Code != liteapi.SuccessCode {
		logrus.Errorf("Failed to import message: %v", importResult.Message)
		return imap.Message{}, nil, fmt.Errorf("failed to create message: %08x", importResult.Code)
	}

	return imap.Message{
		ID:    imap.MessageID(importResult.MessageID),
		Flags: flags,
		Date:  date,
	}, literal, nil
}

// LabelMessages labels the given messages with the given label ID.
func (conn *imapConnector) LabelMessages(ctx context.Context, messageIDs []imap.MessageID, labelID imap.LabelID) error {
	return conn.client.LabelMessages(ctx, mapTo[imap.MessageID, string](messageIDs), string(labelID))
}

// UnlabelMessages unlabels the given messages with the given label ID.
func (conn *imapConnector) UnlabelMessages(ctx context.Context, messageIDs []imap.MessageID, labelID imap.LabelID) error {
	return conn.client.UnlabelMessages(ctx, mapTo[imap.MessageID, string](messageIDs), string(labelID))
}

// MoveMessages removes the given messages from one label and adds them to the other label.
func (conn *imapConnector) MoveMessages(ctx context.Context, messageIDs []imap.MessageID, labelFromID imap.LabelID, labelToID imap.LabelID) error {
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
