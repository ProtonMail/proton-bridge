// Copyright (c) 2020 Proton Technologies AG
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
	"fmt"
	"io"
	"mime/multipart"
	"net/mail"
	"net/textproto"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/internal/imap/cache"
	"github.com/ProtonMail/proton-bridge/internal/imap/uidplus"
	"github.com/ProtonMail/proton-bridge/pkg/message"
	"github.com/ProtonMail/proton-bridge/pkg/parallel"
	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
	"github.com/emersion/go-imap"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	openpgperrors "golang.org/x/crypto/openpgp/errors"
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
func (im *imapMailbox) CreateMessage(flags []string, date time.Time, body imap.Literal) error { // nolint[funlen]
	// Called from go-imap in goroutines - we need to handle panics for each function.
	defer im.panicHandler.HandlePanic()

	m, _, _, readers, err := message.Parse(body, "", "")
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

		// This is an APPEND to the Sent folder, so we will set the sent flag
		m.Flags |= pmapi.FlagSent
	}

	message.ParseFlags(m, flags)
	if !date.IsZero() {
		m.Time = date.Unix()
	}

	internalID := m.Header.Get("X-Pm-Internal-Id")
	references := m.Header.Get("References")
	referenceList := strings.Fields(references)

	if len(referenceList) > 0 {
		lastReference := referenceList[len(referenceList)-1]
		// In case we are using a mail client which corrupts headers, try "References" too.
		re := regexp.MustCompile(pmapi.InternalReferenceFormat)
		match := re.FindStringSubmatch(lastReference)
		if len(match) > 0 {
			internalID = match[0]
		}
	}

	// Avoid appending a message which is already on the server. Apply the new
	// label instead. This sometimes happens which Outlook (it uses APPEND instead of COPY).
	if internalID != "" {
		// Check to see if this belongs to a different address in split mode or another ProtonMail account.
		msg, err := im.storeMailbox.GetMessage(internalID)
		if err == nil && (im.user.user.IsCombinedAddressMode() || (im.storeAddress.AddressID() == msg.Message().AddressID)) {
			IDs := []string{internalID}

			err = im.storeMailbox.LabelMessages(IDs)
			if err != nil {
				return err
			}

			targetSeq := im.storeMailbox.GetUIDList([]string{m.ID})
			return uidplus.AppendResponse(im.storeMailbox.UIDValidity(), targetSeq)
		}
	}

	im.log.Info("Importing external message")
	if err := im.importMessage(m, readers, kr); err != nil {
		im.log.Error("Import failed: ", err)
		return err
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

func (im *imapMailbox) getMessage(storeMessage storeMessageProvider, items []imap.FetchItem) (msg *imap.Message, err error) {
	im.log.WithField("msgID", storeMessage.ID()).Trace("Getting message")

	seqNum, err := storeMessage.SequenceNumber()
	if err != nil {
		return
	}

	m := storeMessage.Message()

	msg = imap.NewMessage(seqNum, items)
	for _, item := range items {
		switch item {
		case imap.FetchEnvelope:
			msg.Envelope = message.GetEnvelope(m)
		case imap.FetchBody, imap.FetchBodyStructure:
			var structure *message.BodyStructure
			if structure, _, err = im.getBodyStructure(storeMessage); err != nil {
				return
			}
			if msg.BodyStructure, err = structure.IMAPBodyStructure([]int{}); err != nil {
				return
			}
		case imap.FetchFlags:
			msg.Flags = message.GetFlags(m)
		case imap.FetchInternalDate:
			msg.InternalDate = time.Unix(m.Time, 0)
		case imap.FetchRFC822Size:
			// Size attribute on the server counts encrypted data. The value is cleared
			// on our part and we need to compute "real" size of decrypted data.
			if m.Size <= 0 {
				if _, _, err = im.getBodyStructure(storeMessage); err != nil {
					return
				}
			}
			msg.Size = uint32(m.Size)
		case imap.FetchUid:
			msg.Uid, err = storeMessage.UID()
			if err != nil {
				return nil, err
			}
		default:
			s := item

			var section *imap.BodySectionName
			if section, err = imap.ParseBodySectionName(s); err != nil {
				err = nil // Ignore error
				break
			}

			var literal imap.Literal
			if literal, err = im.getMessageBodySection(storeMessage, section); err != nil {
				return
			}

			msg.Body[section] = literal
		}
	}

	return msg, err
}

func (im *imapMailbox) getBodyStructure(storeMessage storeMessageProvider) (
	structure *message.BodyStructure,
	bodyReader *bytes.Reader, err error,
) {
	m := storeMessage.Message()
	id := im.storeUser.UserID() + m.ID
	cache.BuildLock(id)
	if bodyReader, structure = cache.LoadMail(id); bodyReader.Len() == 0 || structure == nil {
		var body []byte
		structure, body, err = im.buildMessage(m)
		if err == nil && structure != nil && len(body) > 0 {
			m.Size = int64(len(body))
			if err := storeMessage.SetSize(m.Size); err != nil {
				im.log.WithError(err).
					WithField("newSize", m.Size).
					WithField("msgID", m.ID).
					Warn("Cannot update size while building")
			}
			if err := storeMessage.SetContentTypeAndHeader(m.MIMEType, m.Header); err != nil {
				im.log.WithError(err).
					WithField("msgID", m.ID).
					Warn("Cannot update header while building")
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
func (im *imapMailbox) getMessageBodySection(storeMessage storeMessageProvider, section *imap.BodySectionName) (literal imap.Literal, err error) { // nolint[funlen]
	var (
		structure  *message.BodyStructure
		bodyReader *bytes.Reader
		header     textproto.MIMEHeader
		response   []byte
	)

	im.log.WithField("msgID", storeMessage.ID()).Trace("Getting message body")

	m := storeMessage.Message()

	if len(section.Path) == 0 && section.Specifier == imap.HeaderSpecifier {
		// We can extract message header without decrypting.
		header = message.GetHeader(m)
		// We need to ensure we use the correct content-type,
		// otherwise AppleMail expects `text/plain` in HTML mails.
		if header.Get("Content-Type") == "" {
			if err = im.fetchMessage(m); err != nil {
				return
			}
			if _, err = im.setMessageContentType(m); err != nil {
				return
			}
			if err = storeMessage.SetContentTypeAndHeader(m.MIMEType, m.Header); err != nil {
				return
			}
			header = message.GetHeader(m)
		}
	} else {
		// The rest of cases need download and decrypt.
		structure, bodyReader, err = im.getBodyStructure(storeMessage)
		if err != nil {
			return
		}

		switch {
		case section.Specifier == imap.EntireSpecifier && len(section.Path) == 0:
			//  An empty section specification refers to the entire message, including the header.
			response, err = structure.GetSection(bodyReader, section.Path)
		case section.Specifier == imap.TextSpecifier || (section.Specifier == imap.EntireSpecifier && len(section.Path) != 0):
			// The TEXT specifier refers to the content of the message (or section), omitting the [RFC-2822] header.
			// Non-empty section with no specifier (imap.EntireSpecifier) refers to section content without header.
			response, err = structure.GetSectionContent(bodyReader, section.Path)
		case section.Specifier == imap.MIMESpecifier:
			// The MIME part specifier refers to the [MIME-IMB] header for this part.
			fallthrough
		case section.Specifier == imap.HeaderSpecifier:
			header, err = structure.GetSectionHeader(section.Path)
		default:
			err = errors.New("Unknown specifier " + string(section.Specifier))
		}
	}

	if err != nil {
		return
	}

	// Filter header. Options are: all fields, only selected fields, all fields except selected.
	if header != nil {
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
		response = headerBuf.Bytes()
	}

	// Trim any output if requested.
	literal = bytes.NewBuffer(section.ExtractPartial(response))
	return literal, nil
}

func (im *imapMailbox) fetchMessage(m *pmapi.Message) (err error) {
	im.log.Trace("Fetching message")

	complete, err := im.storeMailbox.FetchMessage(m.ID)
	if err != nil {
		im.log.WithError(err).Error("Could not get message from store")
		return
	}

	*m = *complete.Message()

	return
}

func (im *imapMailbox) writeMessageBody(w io.Writer, m *pmapi.Message) (err error) {
	im.log.Trace("Writing message body")

	if m.Body == "" {
		im.log.Trace("While writing message body, noticed message body is null, need to fetch")
		if err = im.fetchMessage(m); err != nil {
			return
		}
	}

	kr, err := im.user.client().KeyRingForAddressID(m.AddressID)
	if err != nil {
		return errors.Wrap(err, "failed to get keyring for address ID")
	}

	err = message.WriteBody(w, kr, m)
	if err != nil {
		if customMessageErr := message.CustomMessage(m, err, true); customMessageErr != nil {
			im.log.WithError(customMessageErr).Warn("Failed to make custom message")
		}
		_, _ = io.WriteString(w, m.Body)
		err = nil
	}

	return
}

func (im *imapMailbox) writeAttachmentBody(w io.Writer, m *pmapi.Message, att *pmapi.Attachment) (err error) {
	// Retrieve encrypted attachment.
	r, err := im.user.client().GetAttachment(att.ID)
	if err != nil {
		return
	}
	defer r.Close() //nolint[errcheck]

	kr, err := im.user.client().KeyRingForAddressID(m.AddressID)
	if err != nil {
		return errors.Wrap(err, "failed to get keyring for address ID")
	}

	if err = message.WriteAttachmentBody(w, kr, m, att, r); err != nil {
		// Returning an error here makes certain mail clients behave badly,
		// trying to retrieve the message again and again.
		im.log.Warn("Cannot write attachment body: ", err)
		err = nil
	}
	return
}

func (im *imapMailbox) writeRelatedPart(p io.Writer, m *pmapi.Message, inlines []*pmapi.Attachment) (err error) {
	related := multipart.NewWriter(p)

	_ = related.SetBoundary(message.GetRelatedBoundary(m))

	buf := &bytes.Buffer{}
	if err = im.writeMessageBody(buf, m); err != nil {
		return
	}

	// Write the body part.
	h := message.GetBodyHeader(m)

	if p, err = related.CreatePart(h); err != nil {
		return
	}

	_, _ = buf.WriteTo(p)

	for _, inline := range inlines {
		buf = &bytes.Buffer{}
		if err = im.writeAttachmentBody(buf, m, inline); err != nil {
			return
		}

		h := message.GetAttachmentHeader(inline)
		if p, err = related.CreatePart(h); err != nil {
			return
		}
		_, _ = buf.WriteTo(p)
	}

	_ = related.Close()
	return nil
}

const (
	noMultipart      = iota // only body
	simpleMultipart         // body + attachment or inline
	complexMultipart        // mixed, rfc822, alternatives, ...
)

func (im *imapMailbox) setMessageContentType(m *pmapi.Message) (multipartType int, err error) {
	if m.MIMEType == "" {
		err = fmt.Errorf("trying to set Content-Type without MIME TYPE")
		return
	}
	// message.MIMEType can have just three values from our server:
	// * `text/html` (refers to body type, but might contain attachments and inlines)
	// * `text/plain` (refers to body type, but might contain attachments and inlines)
	// * `multipart/mixed` (refers to external message with multipart structure)
	// The proper header content fields must be set and saved to DB based MIMEType and content.
	multipartType = noMultipart
	if m.MIMEType == pmapi.ContentTypeMultipartMixed {
		multipartType = complexMultipart
	} else if m.NumAttachments != 0 {
		multipartType = simpleMultipart
	}

	h := textproto.MIMEHeader(m.Header)
	if multipartType == noMultipart {
		message.SetBodyContentFields(&h, m)
	} else {
		h.Set("Content-Type",
			fmt.Sprintf("%s; boundary=%s", "multipart/mixed", message.GetBoundary(m)),
		)
	}
	m.Header = mail.Header(h)

	return
}

// buildMessage from PM to IMAP.
func (im *imapMailbox) buildMessage(m *pmapi.Message) (structure *message.BodyStructure, msgBody []byte, err error) {
	im.log.Trace("Building message")

	var errNoCache doNotCacheError

	// If fetch or decryption fails we need to change the MIMEType (in customMessage).
	err = im.fetchMessage(m)
	if err != nil {
		return
	}

	kr, err := im.user.client().KeyRingForAddressID(m.AddressID)
	if err != nil {
		err = errors.Wrap(err, "failed to get keyring for address ID")
		return
	}

	errDecrypt := m.Decrypt(kr)

	if errDecrypt != nil && errDecrypt != openpgperrors.ErrSignatureExpired {
		errNoCache.add(errDecrypt)
		if customMessageErr := message.CustomMessage(m, errDecrypt, true); customMessageErr != nil {
			im.log.WithError(customMessageErr).Warn("Failed to make custom message")
		}
	}

	// Inner function can fail even when message is decrypted.
	// #1048 For example we have problem with double-encrypted messages
	// which seems as still encrypted and we try them to decrypt again
	// and that fails. For any building error is better to return custom
	// message than error because it will not be fixed and users would
	// get error message all the time and could not see some messages.
	structure, msgBody, err = im.buildMessageInner(m, kr)
	if err == pmapi.ErrAPINotReachable || err == pmapi.ErrInvalidToken || err == pmapi.ErrUpgradeApplication {
		return nil, nil, err
	} else if err != nil {
		errNoCache.add(err)
		if customMessageErr := message.CustomMessage(m, err, true); customMessageErr != nil {
			im.log.WithError(customMessageErr).Warn("Failed to make custom message")
		}
		structure, msgBody, err = im.buildMessageInner(m, kr)
		if err != nil {
			return nil, nil, err
		}
	}

	err = errNoCache.errorOrNil()

	return structure, msgBody, err
}

func (im *imapMailbox) buildMessageInner(m *pmapi.Message, kr *crypto.KeyRing) (structure *message.BodyStructure, msgBody []byte, err error) { // nolint[funlen]
	multipartType, err := im.setMessageContentType(m)
	if err != nil {
		return
	}

	tmpBuf := &bytes.Buffer{}
	mainHeader := message.GetHeader(m)
	if err = writeHeader(tmpBuf, mainHeader); err != nil {
		return
	}
	_, _ = io.WriteString(tmpBuf, "\r\n")

	switch multipartType {
	case noMultipart:
		err = message.WriteBody(tmpBuf, kr, m)
		if err != nil {
			return
		}
	case complexMultipart:
		_, _ = io.WriteString(tmpBuf, "\r\n--"+message.GetBoundary(m)+"\r\n")
		err = message.WriteBody(tmpBuf, kr, m)
		if err != nil {
			return
		}
		_, _ = io.WriteString(tmpBuf, "\r\n--"+message.GetBoundary(m)+"--\r\n")
	case simpleMultipart:
		atts, inlines := message.SeparateInlineAttachments(m)
		mw := multipart.NewWriter(tmpBuf)
		_ = mw.SetBoundary(message.GetBoundary(m))

		var partWriter io.Writer

		if len(inlines) > 0 {
			relatedHeader := message.GetRelatedHeader(m)
			if partWriter, err = mw.CreatePart(relatedHeader); err != nil {
				return
			}
			_ = im.writeRelatedPart(partWriter, m, inlines)
		} else {
			buf := &bytes.Buffer{}
			if err = im.writeMessageBody(buf, m); err != nil {
				return
			}

			// Write the body part.
			bodyHeader := message.GetBodyHeader(m)
			if partWriter, err = mw.CreatePart(bodyHeader); err != nil {
				return
			}

			_, _ = buf.WriteTo(partWriter)
		}

		// Write the attachments parts.
		input := make([]interface{}, len(atts))
		for i, att := range atts {
			input[i] = att
		}

		processCallback := func(value interface{}) (interface{}, error) {
			att := value.(*pmapi.Attachment)

			buf := &bytes.Buffer{}
			if err = im.writeAttachmentBody(buf, m, att); err != nil {
				return nil, err
			}
			return buf, nil
		}

		collectCallback := func(idx int, value interface{}) error {
			buf := value.(*bytes.Buffer)
			defer buf.Reset()
			att := atts[idx]

			attachmentHeader := message.GetAttachmentHeader(att)
			if partWriter, err = mw.CreatePart(attachmentHeader); err != nil {
				return err
			}

			_, _ = buf.WriteTo(partWriter)
			return nil
		}

		err = parallel.RunParallel(fetchAttachmentsWorkers, input, processCallback, collectCallback)
		if err != nil {
			return
		}

		_ = mw.Close()
	default:
		fmt.Fprintf(tmpBuf, "\r\n\r\nUknown multipart type: %d\r\n\r\n", multipartType)
	}

	// We need to copy buffer before building body structure.
	msgBody = tmpBuf.Bytes()
	structure, err = message.NewBodyStructure(tmpBuf)
	if err != nil {
		// NOTE: We need to set structure if it fails and is empty.
		if structure == nil {
			structure = &message.BodyStructure{}
		}
	}
	return structure, msgBody, err
}
