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
	"bytes"

	"github.com/ProtonMail/proton-bridge/v2/pkg/message"
	"github.com/emersion/go-imap"
	"github.com/pkg/errors"
)

func (im *imapMailbox) getMessage(storeMessage storeMessageProvider, items []imap.FetchItem) (msg *imap.Message, err error) {
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
			// No need to retrieve full header here. API header
			// contains enough information to build the envelope.
			msg.Envelope = message.GetEnvelope(m, storeMessage.GetMIMEHeaderFast())
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
			size, err := storeMessage.GetRFC822Size()
			if err != nil {
				return nil, err
			}

			msg.Size = size
		case imap.FetchUid:
			if msg.Uid, err = storeMessage.UID(); err != nil {
				return nil, err
			}
		case imap.FetchAll, imap.FetchFast, imap.FetchFull, imap.FetchRFC822, imap.FetchRFC822Header, imap.FetchRFC822Text:
			fallthrough // this is list of defined items by go-imap, but items can be also sections generated from requests
		default:
			if err = im.getLiteralForSection(item, msg, storeMessage); err != nil {
				return
			}
		}
	}

	return msg, err
}

func (im *imapMailbox) getLiteralForSection(itemSection imap.FetchItem, msg *imap.Message, storeMessage storeMessageProvider) error {
	section, err := imap.ParseBodySectionName(itemSection)
	if err != nil {
		log.WithError(err).Warn("Failed to parse body section name; part will be skipped")
		return nil //nolint:nilerr ignore error
	}

	var literal imap.Literal
	if literal, err = im.getMessageBodySection(storeMessage, section); err != nil {
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
		if bs, _, err = im.getBodyAndStructure(storeMessage); err != nil {
			return
		}
	}
	return
}

func (im *imapMailbox) getBodyAndStructure(storeMessage storeMessageProvider) (*message.BodyStructure, *bytes.Reader, error) {
	rfc822, err := storeMessage.GetRFC822()
	if err != nil {
		return nil, nil, err
	}

	structure, err := storeMessage.GetBodyStructure()
	if err != nil {
		return nil, nil, err
	}

	return structure, bytes.NewReader(rfc822), nil
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
func (im *imapMailbox) getMessageBodySection(storeMessage storeMessageProvider, section *imap.BodySectionName) (imap.Literal, error) {
	var header []byte
	var response []byte

	im.log.WithField("msgID", storeMessage.ID()).Trace("Getting message body")

	isMainHeaderRequested := len(section.Path) == 0 && section.Specifier == imap.HeaderSpecifier
	if isMainHeaderRequested && storeMessage.IsFullHeaderCached() {
		var err error
		if header, err = storeMessage.GetHeader(); err != nil {
			return nil, err
		}
	} else {
		structure, bodyReader, err := im.getBodyAndStructure(storeMessage)
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
			header, err = structure.GetSectionHeaderBytes(section.Path)
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
