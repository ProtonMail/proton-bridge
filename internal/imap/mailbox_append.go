// Copyright (c) 2021 Proton Technologies AG
//
// This file is part of ProtonMail Bridge.
//
// ProtonMail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ProtonMail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ProtonMail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package imap

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"net/mail"
	"strings"
	"time"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/internal/imap/uidplus"
	"github.com/ProtonMail/proton-bridge/pkg/message"
	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
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

func (im *imapMailbox) createMessage(imapFlags []string, date time.Time, r imap.Literal) error { //nolint[funlen]
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
				return im.labelExistingMessage(msg.ID(), msg.IsMarkedDeleted())
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

func (im *imapMailbox) labelExistingMessage(messageID string, isDeleted bool) error {
	im.log.Info("Labelling existing message")

	// IMAP clients can move message to local folder (setting \Deleted flag)
	// and then move it back (IMAP client does not remember the message,
	// so instead removing the flag it imports duplicate message).
	// Regular IMAP server would keep the message twice and later EXPUNGE would
	// not delete the message (EXPUNGE would delete the original message and
	// the new duplicate one would stay). API detects duplicates; therefore
	// we need to remove \Deleted flag if IMAP client re-imports.
	if isDeleted {
		if err := im.storeMailbox.MarkMessagesUndeleted([]string{messageID}); err != nil {
			log.WithError(err).Error("Failed to undelete re-imported message")
		}
	}

	if err := im.storeMailbox.LabelMessages([]string{messageID}); err != nil {
		return err
	}

	return uidplus.AppendResponse(im.storeMailbox.UIDValidity(), im.storeMailbox.GetUIDList([]string{messageID}))
}

func (im *imapMailbox) importMessage(kr *crypto.KeyRing, hdr textproto.Header, body []byte, imapFlags []string, date time.Time) error {
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

	messageID, err := im.storeMailbox.ImportMessage(enc, seen, labelIDs, flags, time)
	if err != nil {
		return err
	}

	msg, err := im.storeMailbox.GetMessage(messageID)
	if err != nil {
		return err
	}

	if msg.IsMarkedDeleted() {
		if err := im.storeMailbox.MarkMessagesUndeleted([]string{messageID}); err != nil {
			log.WithError(err).Error("Failed to undelete re-imported message")
		}
	}

	return uidplus.AppendResponse(im.storeMailbox.UIDValidity(), im.storeMailbox.GetUIDList([]string{messageID}))
}
