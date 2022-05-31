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
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.

package imap

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"net/mail"
	"strings"
	"time"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/v2/internal/imap/uidplus"
	"github.com/ProtonMail/proton-bridge/v2/pkg/message"
	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-message/textproto"
	"github.com/pkg/errors"
)

// CreateMessage appends a new message to this mailbox. The \Recent flag will
// be added regardless of whether flags is empty or not. If date is nil, the
// current time will be used.
//
// If the Backend implements Updater, it must notify the client immediately
// via a mailbox update.
func (im *imapMailbox) CreateMessage(flags []string, date time.Time, body imap.Literal) error {
	return im.logCommand(func() error {
		return im.createMessage(flags, date, body)
	}, "APPEND", flags, date)
}

func (im *imapMailbox) createMessage(imapFlags []string, date time.Time, r imap.Literal) error { //nolint:funlen
	// Called from go-imap in goroutines - we need to handle panics for each function.
	defer im.panicHandler.HandlePanic()

	// NOTE: Is this lock meant to be here?
	im.user.appendExpungeLock.Lock()
	defer im.user.appendExpungeLock.Unlock()

	body, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	addr := im.storeAddress.APIAddress()
	if addr == nil {
		return errors.New("no available address for encryption")
	}

	kr, err := im.user.client().KeyRingForAddressID(addr.ID)
	if err != nil {
		return err
	}

	if im.storeMailbox.LabelID() == pmapi.DraftLabel {
		return im.createDraftMessage(kr, addr.Email, body)
	}

	if im.storeMailbox.LabelID() == pmapi.SentLabel {
		m, _, _, _, err := message.Parse(bytes.NewReader(body))
		if err != nil {
			return err
		}

		if m.Sender == nil {
			m.Sender = &mail.Address{Address: addr.Email}
		}

		if user, err := im.user.backend.bridge.GetUser(pmapi.SanitizeEmail(m.Sender.Address)); err == nil && user.ID() == im.storeUser.UserID() {
			logEntry := im.log.WithField("sender", m.Sender).WithField("extID", m.Header.Get("Message-Id")).WithField("date", date)

			if foundUID := im.storeMailbox.GetUIDByHeader(&m.Header); foundUID != uint32(0) {
				logEntry.Info("Ignoring APPEND of duplicate to Sent folder")
				return uidplus.AppendResponse(im.storeMailbox.UIDValidity(), &uidplus.OrderedSeq{foundUID})
			}

			logEntry.Info("No matching UID, continuing APPEND to Sent")
		}
	}

	hdr, err := textproto.ReadHeader(bufio.NewReader(bytes.NewReader(body)))
	if err != nil {
		return err
	}

	// Avoid appending a message which is already on the server. Apply the new label instead.
	// This always happens with Outlook because it uses APPEND instead of COPY.
	internalID := hdr.Get("X-Pm-Internal-Id")

	// In case there is a mail client which corrupts headers, try "References" too.
	if internalID == "" {
		if references := strings.Fields(hdr.Get("References")); len(references) > 0 {
			if match := pmapi.RxInternalReferenceFormat.FindStringSubmatch(references[len(references)-1]); len(match) == 2 {
				internalID = match[1]
			}
		}
	}

	if internalID != "" {
		if msg, err := im.storeMailbox.GetMessage(internalID); err == nil {
			if im.user.user.IsCombinedAddressMode() || im.storeAddress.AddressID() == msg.Message().AddressID {
				return im.labelExistingMessage(msg)
			}
		}
	}
	return im.importMessage(kr, hdr, body, imapFlags, date)
}

func (im *imapMailbox) createDraftMessage(kr *crypto.KeyRing, email string, body []byte) error {
	im.log.Info("Creating draft message")

	m, _, _, readers, err := message.Parse(bytes.NewReader(body))
	if err != nil {
		return err
	}

	if m.Sender == nil {
		m.Sender = &mail.Address{}
	}

	m.Sender.Address = pmapi.ConstructAddress(m.Sender.Address, email)

	draft, _, err := im.user.storeUser.CreateDraft(kr, m, readers, "", "", "")
	if err != nil {
		return errors.Wrap(err, "failed to create draft")
	}

	return uidplus.AppendResponse(im.storeMailbox.UIDValidity(), im.storeMailbox.GetUIDList([]string{draft.ID}))
}

func findMailboxForAddress(address storeAddressProvider, labelID string) (storeMailboxProvider, error) {
	for _, mailBox := range address.ListMailboxes() {
		if mailBox.LabelID() == labelID {
			return mailBox, nil
		}
	}
	return nil, fmt.Errorf("could not find %v label in mailbox for user %v", labelID,
		address.AddressString())
}

func (im *imapMailbox) labelExistingMessage(msg storeMessageProvider) error { //nolint:funlen
	im.log.Info("Labelling existing message")

	// IMAP clients can move message to local folder (setting \Deleted flag)
	// and then move it back (IMAP client does not remember the message,
	// so instead removing the flag it imports duplicate message).
	// Regular IMAP server would keep the message twice and later EXPUNGE would
	// not delete the message (EXPUNGE would delete the original message and
	// the new duplicate one would stay). API detects duplicates; therefore
	// we need to remove \Deleted flag if IMAP client re-imports.
	if msg.IsMarkedDeleted() {
		if err := im.storeMailbox.MarkMessagesUndeleted([]string{msg.ID()}); err != nil {
			log.WithError(err).Error("Failed to undelete re-imported message")
		}
	}

	// Outlook Uses APPEND instead of COPY. There is no need to copy to All Mail because messages are already there.
	// If the message is copied from Spam or Trash, it must be moved otherwise we will have data loss.
	// If the message is moved from any folder, the moment when expunge happens on source we will move message trash unless we move it to archive.
	// If the message is already in Archive we should not call API at all.
	// Otherwise the message is already in All mail, Return OK.
	storeMBox := im.storeMailbox
	if pmapi.AllMailLabel == storeMBox.LabelID() {
		if msg.Message().HasLabelID(pmapi.ArchiveLabel) {
			return uidplus.AppendResponse(storeMBox.UIDValidity(), storeMBox.GetUIDList([]string{msg.ID()}))
		}
		var err error
		storeMBox, err = findMailboxForAddress(im.storeAddress, pmapi.ArchiveLabel)
		if err != nil {
			return err
		}
	}

	if err := storeMBox.LabelMessages([]string{msg.ID()}); err != nil {
		return err
	}

	return uidplus.AppendResponse(im.storeMailbox.UIDValidity(), im.storeMailbox.GetUIDList([]string{msg.ID()}))
}

func (im *imapMailbox) importMessage(kr *crypto.KeyRing, hdr textproto.Header, body []byte, imapFlags []string, date time.Time) error { //nolint:funlen
	im.log.Info("Importing external message")

	var (
		seen     bool
		flags    int64
		labelIDs []string
		time     int64
	)

	if hdr.Get("received") == "" {
		flags = pmapi.FlagSent
	} else {
		flags = pmapi.FlagReceived
	}

	for _, flag := range imapFlags {
		switch flag {
		case imap.DraftFlag:
			flags &= ^pmapi.FlagSent
			flags &= ^pmapi.FlagReceived

		case imap.SeenFlag:
			seen = true

		case imap.FlaggedFlag:
			labelIDs = append(labelIDs, pmapi.StarredLabel)

		case imap.AnsweredFlag:
			flags |= pmapi.FlagReplied
		}
	}

	if !date.IsZero() {
		time = date.Unix()
	}

	enc, err := message.EncryptRFC822(kr, bytes.NewReader(body))
	if err != nil {
		return err
	}

	targetMailbox := im.storeMailbox
	if targetMailbox.LabelID() == pmapi.AllMailLabel {
		// Importing mail in directly into All Mail is not allowed. Instead we redirect the import to Archive
		// The mail will automatically appear in All mail. The appends response still reports that the mail was
		// successfully APPEND to All Mail.
		targetMailbox, err = findMailboxForAddress(im.storeAddress, pmapi.ArchiveLabel)
		if err != nil {
			return err
		}
	}

	messageID, err := targetMailbox.ImportMessage(enc, seen, labelIDs, flags, time)
	if err != nil {
		return err
	}

	msg, err := targetMailbox.GetMessage(messageID)
	if err != nil {
		return err
	}

	if msg.IsMarkedDeleted() {
		if err := targetMailbox.MarkMessagesUndeleted([]string{messageID}); err != nil {
			log.WithError(err).Error("Failed to undelete re-imported message")
		}
	}

	return uidplus.AppendResponse(im.storeMailbox.UIDValidity(), im.storeMailbox.GetUIDList([]string{messageID}))
}
