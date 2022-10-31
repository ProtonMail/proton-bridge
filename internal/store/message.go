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

package store

import (
	"bufio"
	"bytes"
	"net/textproto"

	pkgMsg "github.com/ProtonMail/proton-bridge/v2/pkg/message"
	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
	bolt "go.etcd.io/bbolt"
)

// Message is wrapper around `pmapi.Message` with connection to
// a specific mailbox with helper functions to get IMAP UID, sequence
// numbers and similar.
type Message struct {
	msg *pmapi.Message

	store        *Store
	storeMailbox *Mailbox
}

func newStoreMessage(storeMailbox *Mailbox, msg *pmapi.Message) *Message {
	return &Message{
		msg:          msg,
		store:        storeMailbox.store,
		storeMailbox: storeMailbox,
	}
}

// ID returns message ID on our API (always the same ID for all mailboxes).
func (message *Message) ID() string {
	return message.msg.ID
}

// UID returns message UID for IMAP, specific for mailbox used to get the message.
func (message *Message) UID() (uint32, error) {
	return message.storeMailbox.getUID(message.ID())
}

// SequenceNumber returns index of message in used mailbox.
func (message *Message) SequenceNumber() (uint32, error) {
	return message.storeMailbox.getSequenceNumber(message.ID())
}

// Message returns message struct from pmapi.
func (message *Message) Message() *pmapi.Message {
	return message.msg
}

// IsMarkedDeleted returns true if message is marked as deleted for specific mailbox.
func (message *Message) IsMarkedDeleted() bool {
	var isMarkedAsDeleted bool

	if err := message.storeMailbox.db().View(func(tx *bolt.Tx) error {
		isMarkedAsDeleted = message.storeMailbox.txGetDeletedIDsBucket(tx).Get([]byte(message.msg.ID)) != nil
		return nil
	}); err != nil {
		message.storeMailbox.log.WithError(err).Error("Not able to retrieve deleted mark, assuming false.")
		return false
	}

	return isMarkedAsDeleted
}

// IsFullHeaderCached will check that valid full header is stored in DB.
func (message *Message) IsFullHeaderCached() bool {
	var raw []byte
	err := message.store.db.View(func(tx *bolt.Tx) error {
		raw = tx.Bucket(bodystructureBucket).Get([]byte(message.ID()))
		return nil
	})
	return err == nil && raw != nil
}

func (message *Message) getRawHeader() ([]byte, error) {
	bs, err := message.GetBodyStructure()
	if err != nil {
		return nil, err
	}

	return bs.GetMailHeaderBytes()
}

// GetHeader will return cached header from DB.
func (message *Message) GetHeader() ([]byte, error) {
	raw, err := message.getRawHeader()
	if err != nil {
		message.store.log.
			WithField("msgID", message.ID()).
			WithError(err).
			Warn("Cannot get raw header")
		return nil, err
	}

	return raw, nil
}

// GetMIMEHeaderFast returns full header if message was cached. If full header
// is not available it will return header from metadata.
// NOTE: Returned header may not contain all fields.
func (message *Message) GetMIMEHeaderFast() (header textproto.MIMEHeader) {
	var err error
	if message.IsFullHeaderCached() {
		header, err = message.GetMIMEHeader()
	}
	if header == nil || err != nil {
		header = textproto.MIMEHeader(message.Message().Header)
	}
	return
}

// GetMIMEHeader will return cached header from DB, parsed as a textproto.MIMEHeader.
func (message *Message) GetMIMEHeader() (textproto.MIMEHeader, error) {
	raw, err := message.getRawHeader()
	if err != nil {
		message.store.log.
			WithField("msgID", message.ID()).
			WithError(err).
			Warn("Cannot get raw header for MIME header")
		return nil, err
	}

	header, err := textproto.NewReader(bufio.NewReader(bytes.NewReader(raw))).ReadMIMEHeader()
	if err != nil {
		message.store.log.
			WithField("msgID", message.ID()).
			WithError(err).
			Warn("Cannot build header from bodystructure")
		return nil, err
	}

	return header, nil
}

// GetBodyStructure returns the message's body structure.
// It checks first if it's in the store. If it is, it returns it from store,
// otherwise it computes it from the message cache (and saves the result to the store).
func (message *Message) GetBodyStructure() (*pkgMsg.BodyStructure, error) {
	var raw []byte

	if err := message.store.db.View(func(tx *bolt.Tx) error {
		raw = tx.Bucket(bodystructureBucket).Get([]byte(message.ID()))
		return nil
	}); err != nil {
		return nil, err
	}

	if len(raw) > 0 {
		// If not possible to deserialize just continue with build.
		if bs, err := pkgMsg.DeserializeBodyStructure(raw); err == nil {
			return bs, nil
		}
	}

	literal, err := message.store.getCachedMessage(message.ID())
	if err != nil {
		return nil, err
	}

	bs, err := pkgMsg.NewBodyStructure(bytes.NewReader(literal))
	if err != nil {
		return nil, err
	}

	// Do not cache draft bodystructure
	if message.msg.IsDraft() {
		return bs, nil
	}

	if raw, err = bs.Serialize(); err != nil {
		return nil, err
	}

	if err := message.store.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bodystructureBucket).Put([]byte(message.ID()), raw)
	}); err != nil {
		return nil, err
	}

	return bs, nil
}

// GetRFC822 returns the raw message literal.
func (message *Message) GetRFC822() ([]byte, error) {
	return message.store.getCachedMessage(message.ID())
}

// GetRFC822Size returns the size of the raw message literal.
func (message *Message) GetRFC822Size() (uint32, error) {
	var raw []byte

	if err := message.store.db.View(func(tx *bolt.Tx) error {
		raw = tx.Bucket(sizeBucket).Get([]byte(message.ID()))
		return nil
	}); err != nil {
		return 0, err
	}

	if len(raw) > 0 {
		return btoi(raw), nil
	}

	literal, err := message.store.getCachedMessage(message.ID())
	if err != nil {
		return 0, err
	}

	// Do not cache draft size
	if !message.msg.IsDraft() {
		if err := message.store.db.Update(func(tx *bolt.Tx) error {
			return tx.Bucket(sizeBucket).Put([]byte(message.ID()), itob(uint32(len(literal))))
		}); err != nil {
			return 0, err
		}
	}

	return uint32(len(literal)), nil
}
