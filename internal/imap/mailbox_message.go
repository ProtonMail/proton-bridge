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
	"bytes"
	"context"
	"fmt"
	"io"
	"net/mail"
	"net/textproto"
	"sort"
	"strings"
	"time"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/internal/imap/cache"
	"github.com/ProtonMail/proton-bridge/internal/imap/uidplus"
	"github.com/ProtonMail/proton-bridge/pkg/message"
	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
	"github.com/emersion/go-imap"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

var (
	rfc822Birthday = time.Date(1982, 8, 13, 0, 0, 0, 0, time.UTC) //nolint[gochecknoglobals]
)

type doNotCacheError struct{ e error }

func (dnc *doNotCacheError) Error() string { return dnc.e.Error() }
func (dnc *doNotCacheError) add(err error) { dnc.e = multierror.Append(dnc.e, err) }
func (dnc *doNotCacheError) errorOrNil() error {
	if dnc == nil {
		return nil
	}

	if dnc.e != nil {
		return dnc
	}

	return nil
}

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

func (im *imapMailbox) createMessage(flags []string, date time.Time, body imap.Literal) error { // nolint[funlen]
	// Called from go-imap in goroutines - we need to handle panics for each function.
	defer im.panicHandler.HandlePanic()

	m, _, _, readers, err := message.Parse(body)
	if err != nil {
		return err
	}

	addr := im.storeAddress.APIAddress()
	if addr == nil {
		return errors.New("no available address for encryption")
	}
	m.AddressID = addr.ID

	kr, err := im.user.client().KeyRingForAddressID(addr.ID)
	if err != nil {
		return err
	}

	// Handle imported messages which have no "Sender" address.
	// This sometimes occurs with outlook which reports errors as imported emails or for drafts.
	if m.Sender == nil {
		im.log.Warning("Append: Missing email sender. Will use main address")
		m.Sender = &mail.Address{
			Name:    "",
			Address: addr.Email,
		}
	}

	// "Drafts" needs to call special API routes.
	// Clients always append the whole message again and remove the old one.
	if im.storeMailbox.LabelID() == pmapi.DraftLabel {
		// Sender address needs to be sanitised (drafts need to match cases exactly).
		m.Sender.Address = pmapi.ConstructAddress(m.Sender.Address, addr.Email)

		draft, _, err := im.user.storeUser.CreateDraft(kr, m, readers, "", "", "")
		if err != nil {
			return errors.Wrap(err, "failed to create draft")
		}

		targetSeq := im.storeMailbox.GetUIDList([]string{draft.ID})
		return uidplus.AppendResponse(im.storeMailbox.UIDValidity(), targetSeq)
	}

	// We need to make sure this is an import, and not a sent message from this account
	// (sent messages from the account will be added by the event loop).
	if im.storeMailbox.LabelID() == pmapi.SentLabel {
		sanitizedSender := pmapi.SanitizeEmail(m.Sender.Address)

		// Check whether this message was sent by a bridge user.
		user, err := im.user.backend.bridge.GetUser(sanitizedSender)
		if err == nil && user.ID() == im.storeUser.UserID() {
			logEntry := im.log.WithField("addr", sanitizedSender).WithField("extID", m.Header.Get("Message-Id"))

			// If we find the message in the store already, we can skip importing it.
			if foundUID := im.storeMailbox.GetUIDByHeader(&m.Header); foundUID != uint32(0) {
				logEntry.Info("Ignoring APPEND of duplicate to Sent folder")
				return uidplus.AppendResponse(im.storeMailbox.UIDValidity(), &uidplus.OrderedSeq{foundUID})
			}

			// We didn't find the message in the store, so we are currently sending it.
			logEntry.WithField("time", date).Info("No matching UID, continuing APPEND to Sent")
		}
	}

	message.ParseFlags(m, flags)
	if !date.IsZero() {
		m.Time = date.Unix()
	}

	internalID := m.Header.Get("X-Pm-Internal-Id")
	references := m.Header.Get("References")
	referenceList := strings.Fields(references)

	// In case there is a mail client which corrupts headers, try
	// "References" too.
	if internalID == "" && len(referenceList) > 0 {
		lastReference := referenceList[len(referenceList)-1]
		match := pmapi.RxInternalReferenceFormat.FindStringSubmatch(lastReference)
		if len(match) == 2 {
			internalID = match[1]
		}
	}

	im.user.appendExpungeLock.Lock()
	defer im.user.appendExpungeLock.Unlock()

	// Avoid appending a message which is already on the server. Apply the
	// new label instead. This always happens with Outlook (it uses APPEND
	// instead of COPY).
	if internalID != "" {
		// Check to see if this belongs to a different address in split mode or another ProtonMail account.
		msg, err := im.storeMailbox.GetMessage(internalID)
		if err == nil && (im.user.user.IsCombinedAddressMode() || (im.storeAddress.AddressID() == msg.Message().AddressID)) {
			IDs := []string{internalID}

			// See the comment bellow.
			if msg.IsMarkedDeleted() {
				if err := im.storeMailbox.MarkMessagesUndeleted(IDs); err != nil {
					log.WithError(err).Error("Failed to undelete re-imported internal message")
				}
			}

			err = im.storeMailbox.LabelMessages(IDs)
			if err != nil {
				return err
			}

			targetSeq := im.storeMailbox.GetUIDList(IDs)
			return uidplus.AppendResponse(im.storeMailbox.UIDValidity(), targetSeq)
		}
	}

	im.log.Info("Importing external message")
	if err := im.importMessage(m, readers, kr); err != nil {
		im.log.Error("Import failed: ", err)
		return err
	}

	// IMAP clients can move message to local folder (setting \Deleted flag)
	// and then move it back (IMAP client does not remember the message,
	// so instead removing the flag it imports duplicate message).
	// Regular IMAP server would keep the message twice and later EXPUNGE would
	// not delete the message (EXPUNGE would delete the original message and
	// the new duplicate one would stay). API detects duplicates; therefore
	// we need to remove \Deleted flag if IMAP client re-imports.
	msg, err := im.storeMailbox.GetMessage(m.ID)
	if err == nil && msg.IsMarkedDeleted() {
		if err := im.storeMailbox.MarkMessagesUndeleted([]string{m.ID}); err != nil {
			log.WithError(err).Error("Failed to undelete re-imported message")
		}
	}

	targetSeq := im.storeMailbox.GetUIDList([]string{m.ID})
	return uidplus.AppendResponse(im.storeMailbox.UIDValidity(), targetSeq)
}

func (im *imapMailbox) importMessage(m *pmapi.Message, readers []io.Reader, kr *crypto.KeyRing) (err error) { // nolint[funlen]
	body, err := message.BuildEncrypted(m, readers, kr)
	if err != nil {
		return err
	}

	labels := []string{}
	for _, l := range m.LabelIDs {
		if l == pmapi.StarredLabel {
			labels = append(labels, pmapi.StarredLabel)
		}
	}

	return im.storeMailbox.ImportMessage(m, body, labels)
}

func (im *imapMailbox) getMessage(storeMessage storeMessageProvider, items []imap.FetchItem, msgBuildCountHistogram *msgBuildCountHistogram) (msg *imap.Message, err error) { //nolint[funlen]
	msglog := im.log.WithField("msgID", storeMessage.ID())
	msglog.Trace("Getting message")

	seqNum, err := storeMessage.SequenceNumber()
	if err != nil {
		return
	}

	m := storeMessage.Message()

	msg = imap.NewMessage(seqNum, items)
	for _, item := range items {
		switch item {
		case imap.FetchEnvelope:
			// No need to check IsFullHeaderCached here. API header
			// contain enough information to build the envelope.
			msg.Envelope = message.GetEnvelope(m, storeMessage.GetHeader())
		case imap.FetchBody, imap.FetchBodyStructure:
			var structure *message.BodyStructure
			structure, err = im.getBodyStructure(storeMessage)
			if err != nil {
				return
			}
			if msg.BodyStructure, err = structure.IMAPBodyStructure([]int{}); err != nil {
				return
			}
		case imap.FetchFlags:
			msg.Flags = message.GetFlags(m)
			if storeMessage.IsMarkedDeleted() {
				msg.Flags = append(msg.Flags, imap.DeletedFlag)
			}
		case imap.FetchInternalDate:
			msg.InternalDate = time.Unix(m.Time, 0)

			// Apple Mail crashes fetching messages with date older than 1970.
			// There is no point having message older than RFC itself, it's not possible.
			if msg.InternalDate.Before(rfc822Birthday) {
				msg.InternalDate = rfc822Birthday
			}
		case imap.FetchRFC822Size:
			// Size attribute on the server counts encrypted data. The value is cleared
			// on our part and we need to compute "real" size of decrypted data.
			if m.Size <= 0 {
				msglog.Debug("Size unknown - downloading body")
				// We are sure the size is not a problem right now. Clients
				// might not first check sizes of all messages so we couldn't
				// be sure if seeing 1st or 2nd sync is all right or not.
				// Therefore, it's better to exclude getting size from the
				// counting and see build count as real message build.
				if _, _, err = im.getBodyAndStructure(storeMessage, nil); err != nil {
					return
				}
			}
			msg.Size = uint32(m.Size)
		case imap.FetchUid:
			msg.Uid, err = storeMessage.UID()
			if err != nil {
				return nil, err
			}
		case imap.FetchAll, imap.FetchFast, imap.FetchFull, imap.FetchRFC822, imap.FetchRFC822Header, imap.FetchRFC822Text:
			fallthrough // this is list of defined items by go-imap, but items can be also sections generated from requests
		default:
			if err = im.getLiteralForSection(item, msg, storeMessage, msgBuildCountHistogram); err != nil {
				return
			}
		}
	}

	return msg, err
}

func (im *imapMailbox) getLiteralForSection(itemSection imap.FetchItem, msg *imap.Message, storeMessage storeMessageProvider, msgBuildCountHistogram *msgBuildCountHistogram) error {
	section, err := imap.ParseBodySectionName(itemSection)
	if err != nil {
		log.WithError(err).Warn("Failed to parse body section name; part will be skipped")
		return nil //nolint[nilerr] ignore error
	}

	var literal imap.Literal
	if literal, err = im.getMessageBodySection(storeMessage, section, msgBuildCountHistogram); err != nil {
		return err
	}

	msg.Body[section] = literal
	return nil
}

func (im *imapMailbox) getBodyStructure(storeMessage storeMessageProvider) (bs *message.BodyStructure, err error) {
	// Apple Mail requests body structure for all
	// messages irregularly. We cache bodystructure in
	// local database in order to not re-download all
	// messages from server.
	bs, err = storeMessage.GetBodyStructure()
	if err != nil {
		im.log.WithError(err).Debug("Fail to retrieve bodystructure from database")
	}
	if bs == nil {
		// We are sure the body structure is not a problem right now.
		// Clients might do first fetch body structure so we couldn't
		// be sure if seeing 1st or 2nd sync is all right or not.
		// Therefore, it's better to exclude first body structure fetch
		// from the counting and see build count as real message build.
		if bs, _, err = im.getBodyAndStructure(storeMessage, nil); err != nil {
			return
		}
	}
	return
}

//nolint[funlen] Jakub will fix in refactor
func (im *imapMailbox) getBodyAndStructure(storeMessage storeMessageProvider, msgBuildCountHistogram *msgBuildCountHistogram) (
	structure *message.BodyStructure,
	bodyReader *bytes.Reader, err error,
) {
	m := storeMessage.Message()
	id := im.storeUser.UserID() + m.ID
	cache.BuildLock(id)
	if bodyReader, structure = cache.LoadMail(id); bodyReader.Len() == 0 || structure == nil {
		var body []byte
		structure, body, err = im.buildMessage(m)
		m.Size = int64(len(body))
		// Save size and body structure even for messages unable to decrypt
		// so the size or body structure doesn't have to be computed every time.
		if err := storeMessage.SetSize(m.Size); err != nil {
			im.log.WithError(err).
				WithField("newSize", m.Size).
				WithField("msgID", m.ID).
				Warn("Cannot update size while building")
		}
		if structure != nil && !isMessageInDraftFolder(m) {
			if err := storeMessage.SetBodyStructure(structure); err != nil {
				im.log.WithError(err).
					WithField("msgID", m.ID).
					Warn("Cannot update bodystructure while building")
			}
		}
		if err == nil && structure != nil && len(body) > 0 {
			header, errHead := structure.GetMailHeaderBytes(bytes.NewReader(body))
			if errHead == nil {
				if errHead := storeMessage.SetHeader(header); errHead != nil {
					im.log.WithError(errHead).
						WithField("msgID", m.ID).
						Warn("Cannot update header after building")
				}
			} else {
				im.log.WithError(errHead).
					WithField("msgID", m.ID).
					Warn("Cannot get header bytes after building")
			}
			if msgBuildCountHistogram != nil {
				times, err := storeMessage.IncreaseBuildCount()
				if err != nil {
					im.log.WithError(err).
						WithField("msgID", m.ID).
						Warn("Cannot increase build count")
				}
				msgBuildCountHistogram.add(times)
			}
			// Drafts can change and we don't want to cache them.
			if !isMessageInDraftFolder(m) {
				cache.SaveMail(id, body, structure)
			}
			bodyReader = bytes.NewReader(body)
		}
		if _, ok := err.(*doNotCacheError); ok {
			im.log.WithField("msgID", m.ID).Errorf("do not cache message: %v", err)
			err = nil
			bodyReader = bytes.NewReader(body)
		}
	}
	cache.BuildUnlock(id)
	return structure, bodyReader, err
}

func isMessageInDraftFolder(m *pmapi.Message) bool {
	for _, labelID := range m.LabelIDs {
		if labelID == pmapi.DraftLabel {
			return true
		}
	}
	return false
}

// This will download message (or read from cache) and pick up the section,
// extract data (header,body, both) and trim the output if needed.
func (im *imapMailbox) getMessageBodySection(
	storeMessage storeMessageProvider,
	section *imap.BodySectionName,
	msgBuildCountHistogram *msgBuildCountHistogram,
) (imap.Literal, error) {
	var header textproto.MIMEHeader
	var response []byte

	im.log.WithField("msgID", storeMessage.ID()).Trace("Getting message body")

	isMainHeaderRequested := len(section.Path) == 0 && section.Specifier == imap.HeaderSpecifier
	if isMainHeaderRequested && storeMessage.IsFullHeaderCached() {
		// In order to speed up (avoid download and decryptions) we
		// cache the header. If a mail header was requested and DB
		// contains full header (it means it was already built once)
		// the DB header can be used without downloading and decrypting.
		// Otherwise header is incomplete and clients would have issues
		// e.g. AppleMail expects `text/plain` in HTML mails.
		header = storeMessage.GetHeader()
	} else {
		// For all other cases it is necessary to download and decrypt the message
		// and drop the header which was obtained from cache. The header will
		// will be stored in DB once successfully built. Check `getBodyAndStructure`.
		structure, bodyReader, err := im.getBodyAndStructure(storeMessage, msgBuildCountHistogram)
		if err != nil {
			return nil, err
		}

		switch {
		case section.Specifier == imap.EntireSpecifier && len(section.Path) == 0:
			//  An empty section specification refers to the entire message, including the header.
			response, err = structure.GetSection(bodyReader, section.Path)
		case section.Specifier == imap.TextSpecifier || (section.Specifier == imap.EntireSpecifier && len(section.Path) != 0):
			// The TEXT specifier refers to the content of the message (or section), omitting the [RFC-2822] header.
			// Non-empty section with no specifier (imap.EntireSpecifier) refers to section content without header.
			response, err = structure.GetSectionContent(bodyReader, section.Path)
		case section.Specifier == imap.MIMESpecifier: // The MIME part specifier refers to the [MIME-IMB] header for this part.
			fallthrough
		case section.Specifier == imap.HeaderSpecifier:
			header, err = structure.GetSectionHeader(section.Path)
		default:
			err = errors.New("Unknown specifier " + string(section.Specifier))
		}

		if err != nil {
			return nil, err
		}
	}

	if header != nil {
		response = filteredHeaderAsBytes(header, section)
	}

	// Trim any output if requested.
	return bytes.NewBuffer(section.ExtractPartial(response)), nil
}

// filteredHeaderAsBytes filters the header fields by section fields and it
// returns the filtered fields as bytes.
// Options are: all fields, only selected fields, all fields except selected.
func filteredHeaderAsBytes(header textproto.MIMEHeader, section *imap.BodySectionName) []byte {
	// remove fields
	if len(section.Fields) != 0 && section.NotFields {
		for _, field := range section.Fields {
			header.Del(field)
		}
	}

	fields := make([]string, 0, len(header))
	if len(section.Fields) == 0 || section.NotFields { // add all and sort
		for f := range header {
			fields = append(fields, f)
		}
		sort.Strings(fields)
	} else { // add only requested (in requested order)
		for _, f := range section.Fields {
			fields = append(fields, textproto.CanonicalMIMEHeaderKey(f))
		}
	}

	headerBuf := &bytes.Buffer{}
	for _, canonical := range fields {
		if values, ok := header[canonical]; !ok {
			continue
		} else {
			for _, val := range values {
				fmt.Fprintf(headerBuf, "%s: %s\r\n", canonical, val)
			}
		}
	}
	return headerBuf.Bytes()
}

// buildMessage from PM to IMAP.
func (im *imapMailbox) buildMessage(m *pmapi.Message) (*message.BodyStructure, []byte, error) {
	body, err := im.builder.NewJobWithOptions(
		context.Background(),
		im.user.client(),
		m.ID,
		message.JobOptions{
			IgnoreDecryptionErrors: true, // Whether to ignore decryption errors and create a "custom message" instead.
			SanitizeDate:           true, // Whether to replace all dates before 1970 with RFC822's birthdate.
			AddInternalID:          true, // Whether to include MessageID as X-Pm-Internal-Id.
			AddExternalID:          true, // Whether to include ExternalID as X-Pm-External-Id.
			AddMessageDate:         true, // Whether to include message time as X-Pm-Date.
			AddMessageIDReference:  true, // Whether to include the MessageID in References.
		},
	).GetResult()
	if err != nil {
		return nil, nil, err
	}

	structure, err := message.NewBodyStructure(bytes.NewReader(body))
	if err != nil {
		return nil, nil, err
	}

	return structure, body, nil
}
