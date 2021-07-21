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

	"github.com/ProtonMail/proton-bridge/internal/imap/cache"
	"github.com/ProtonMail/proton-bridge/pkg/message"
	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
	"github.com/emersion/go-imap"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func (im *imapMailbox) getMessage(
	storeMessage storeMessageProvider,
	items []imap.FetchItem,
	msgBuildCountHistogram *msgBuildCountHistogram,
) (msg *imap.Message, err error) {
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
			msg.Envelope = message.GetEnvelope(m, storeMessage.GetMIMEHeader())
		case imap.FetchBody, imap.FetchBodyStructure:
			structure, err := im.getBodyStructure(storeMessage)
			if err != nil {
				return nil, err
			}
			if msg.BodyStructure, err = structure.IMAPBodyStructure([]int{}); err != nil {
				return nil, err
			}
		case imap.FetchFlags:
			msg.Flags = message.GetFlags(m)
			if storeMessage.IsMarkedDeleted() {
				msg.Flags = append(msg.Flags, imap.DeletedFlag)
			}
		case imap.FetchInternalDate:
			// Apple Mail crashes fetching messages with date older than 1970.
			// There is no point having message older than RFC itself, it's not possible.
			msg.InternalDate = message.SanitizeMessageDate(m.Time)
		case imap.FetchRFC822Size:
			if msg.Size, err = im.getSize(storeMessage); err != nil {
				return nil, err
			}
		case imap.FetchUid:
			if msg.Uid, err = storeMessage.UID(); err != nil {
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

// getSize returns cached size or it will build the message, save the size in
// DB and then returns the size after build.
//
// We are storing size in DB as part of pmapi messages metada. The size
// attribute on the server represents size of encrypted body. The value is
// cleared in Bridge and the final decrypted size (including header, attachment
// and MIME structure) is computed after building the message.
func (im *imapMailbox) getSize(storeMessage storeMessageProvider) (uint32, error) {
	m := storeMessage.Message()
	if m.Size <= 0 {
		im.log.WithField("msgID", m.ID).Debug("Size unknown - downloading body")
		// We are sure the size is not a problem right now. Clients
		// might not first check sizes of all messages so we couldn't
		// be sure if seeing 1st or 2nd sync is all right or not.
		// Therefore, it's better to exclude getting size from the
		// counting and see build count as real message build.
		if _, _, err := im.getBodyAndStructure(storeMessage, nil); err != nil {
			return 0, err
		}
	}
	return uint32(m.Size), nil
}

func (im *imapMailbox) getLiteralForSection(
	itemSection imap.FetchItem,
	msg *imap.Message,
	storeMessage storeMessageProvider,
	msgBuildCountHistogram *msgBuildCountHistogram,
) error {
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

// getBodyStructure returns the cached body structure or it will build the message,
// save the structure in DB and then returns the structure after build.
//
// Apple Mail requests body structure for all messages irregularly. We cache
// bodystructure in local database in order to not re-download all messages
// from server.
func (im *imapMailbox) getBodyStructure(storeMessage storeMessageProvider) (bs *message.BodyStructure, err error) {
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

func (im *imapMailbox) getBodyAndStructure(
	storeMessage storeMessageProvider, msgBuildCountHistogram *msgBuildCountHistogram,
) (
	structure *message.BodyStructure, bodyReader *bytes.Reader, err error,
) {
	m := storeMessage.Message()
	id := im.storeUser.UserID() + m.ID
	cache.BuildLock(id)
	defer cache.BuildUnlock(id)
	bodyReader, structure = cache.LoadMail(id)

	// return the message which was found in cache
	if bodyReader.Len() != 0 && structure != nil {
		return structure, bodyReader, nil
	}

	structure, body, err := im.buildMessage(m)
	bodyReader = bytes.NewReader(body)
	size := int64(len(body))
	l := im.log.WithField("newSize", size).WithField("msgID", m.ID)

	if err != nil || structure == nil || size == 0 {
		l.WithField("hasStructure", structure != nil).Warn("Failed to build message")
		return structure, bodyReader, err
	}

	// Save the size, body structure and header even for messages which
	// were unable to decrypt. Hence they doesn't have to be computed every
	// time.
	m.Size = size
	cacheMessageInStore(storeMessage, structure, body, l)

	if msgBuildCountHistogram != nil {
		times, errCount := storeMessage.IncreaseBuildCount()
		if errCount != nil {
			l.WithError(errCount).Warn("Cannot increase build count")
		}
		msgBuildCountHistogram.add(times)
	}

	// Drafts can change therefore we don't want to cache them.
	if !isMessageInDraftFolder(m) {
		cache.SaveMail(id, body, structure)
	}

	return structure, bodyReader, err
}

func cacheMessageInStore(storeMessage storeMessageProvider, structure *message.BodyStructure, body []byte, l *logrus.Entry) {
	m := storeMessage.Message()
	if errSize := storeMessage.SetSize(m.Size); errSize != nil {
		l.WithError(errSize).Warn("Cannot update size while building")
	}
	if structure != nil && !isMessageInDraftFolder(m) {
		if errStruct := storeMessage.SetBodyStructure(structure); errStruct != nil {
			l.WithError(errStruct).Warn("Cannot update bodystructure while building")
		}
	}
	header, errHead := structure.GetMailHeaderBytes(bytes.NewReader(body))
	if errHead == nil && len(header) != 0 {
		if errStore := storeMessage.SetHeader(header); errStore != nil {
			l.WithError(errStore).Warn("Cannot update header in store")
		}
	} else {
		l.WithError(errHead).Warn("Cannot get header bytes from structure")
	}
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
//
// In order to speed up (avoid download and decryptions) we
// cache the header. If a mail header was requested and DB
// contains full header (it means it was already built once)
// the DB header can be used without downloading and decrypting.
// Otherwise header is incomplete and clients would have issues
// e.g. AppleMail expects `text/plain` in HTML mails.
//
// For all other cases it is necessary to download and decrypt the message
// and drop the header which was obtained from cache. The header will
// will be stored in DB once successfully built. Check `getBodyAndStructure`.
func (im *imapMailbox) getMessageBodySection(
	storeMessage storeMessageProvider,
	section *imap.BodySectionName,
	msgBuildCountHistogram *msgBuildCountHistogram,
) (imap.Literal, error) {
	var header []byte
	var response []byte

	im.log.WithField("msgID", storeMessage.ID()).Trace("Getting message body")

	isMainHeaderRequested := len(section.Path) == 0 && section.Specifier == imap.HeaderSpecifier
	if isMainHeaderRequested && storeMessage.IsFullHeaderCached() {
		header = storeMessage.GetHeader()
	} else {
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
			header, err = structure.GetSectionHeaderBytes(bodyReader, section.Path)
		default:
			err = errors.New("Unknown specifier " + string(section.Specifier))
		}

		if err != nil {
			return nil, err
		}
	}

	if header != nil {
		response = filterHeader(header, section)
	}

	// Trim any output if requested.
	return bytes.NewBuffer(section.ExtractPartial(response)), nil
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
