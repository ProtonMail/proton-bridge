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

package store

import (
	"bufio"
	"bytes"
	"net/mail"
	"net/textproto"

	pkgMsg "github.com/ProtonMail/proton-bridge/pkg/message"
	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
	"github.com/pkg/errors"
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

// IsMarkedDeleted returns true if message is marked as deleted for specific
// mailbox.
func (message *Message) IsMarkedDeleted() bool {
	isMarkedAsDeleted := false
	err := message.storeMailbox.db().View(func(tx *bolt.Tx) error {
		isMarkedAsDeleted = message.storeMailbox.txGetDeletedIDsBucket(tx).Get([]byte(message.msg.ID)) != nil
		return nil
	})
	if err != nil {
		message.storeMailbox.log.WithError(err).Error("Not able to retrieve deleted mark, assuming false.")
		return false
	}
	return isMarkedAsDeleted
}

// SetSize updates the information about size of decrypted message which can be
// used for IMAP. This should not trigger any IMAP update.
// NOTE: The size from the server corresponds to pure body bytes. Hence it
// should not be used. The correct size has to be calculated from decrypted and
// built message.
func (message *Message) SetSize(size int64) error {
	message.msg.Size = size
	txUpdate := func(tx *bolt.Tx) error {
		stored, err := message.store.txGetMessage(tx, message.msg.ID)
		if err != nil {
			return err
		}
		stored.Size = size
		return message.store.txPutMessage(
			tx.Bucket(metadataBucket),
			stored,
		)
	}
	return message.store.db.Update(txUpdate)
}

// SetContentTypeAndHeader updates the information about content type and
// header of decrypted message. This should not trigger any IMAP update.
// NOTE: Content type depends on details of decrypted message which we want to
// cache.
//
// Deprecated: Use SetHeader instead.
func (message *Message) SetContentTypeAndHeader(mimeType string, header mail.Header) error {
	message.msg.MIMEType = mimeType
	message.msg.Header = header
	txUpdate := func(tx *bolt.Tx) error {
		stored, err := message.store.txGetMessage(tx, message.msg.ID)
		if err != nil {
			return err
		}
		stored.MIMEType = mimeType
		stored.Header = header
		return message.store.txPutMessage(
			tx.Bucket(metadataBucket),
			stored,
		)
	}
	return message.store.db.Update(txUpdate)
}

// SetHeader checks header can be parsed and if yes it stores header bytes in
// database.
func (message *Message) SetHeader(header []byte) error {
	_, err := textproto.NewReader(bufio.NewReader(bytes.NewReader(header))).ReadMIMEHeader()
	if err != nil {
		return err
	}
	return message.store.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(headersBucket).Put([]byte(message.ID()), header)
	})
}

// IsFullHeaderCached will check that valid full header is stored in DB.
func (message *Message) IsFullHeaderCached() bool {
	header, err := message.getRawHeader()
	return err == nil && header != nil
}

func (message *Message) getRawHeader() (raw []byte, err error) {
	err = message.store.db.View(func(tx *bolt.Tx) error {
		raw = tx.Bucket(headersBucket).Get([]byte(message.ID()))
		return nil
	})
	return
}

// GetHeader will return cached header from DB.
func (message *Message) GetHeader() []byte {
	raw, err := message.getRawHeader()
	if err != nil {
		panic(errors.Wrap(err, "failed to get raw message header"))
	}

	return raw
}

// GetMIMEHeader will return cached header from DB, parsed as a textproto.MIMEHeader.
func (message *Message) GetMIMEHeader() textproto.MIMEHeader {
	raw, err := message.getRawHeader()
	if err != nil {
		panic(errors.Wrap(err, "failed to get raw message header"))
	}

	header, err := textproto.NewReader(bufio.NewReader(bytes.NewReader(raw))).ReadMIMEHeader()
	if err != nil {
		return textproto.MIMEHeader(message.msg.Header)
	}

	return header
}

// SetBodyStructure stores serialized body structure in database.
func (message *Message) SetBodyStructure(bs *pkgMsg.BodyStructure) error {
	txUpdate := func(tx *bolt.Tx) error {
		return message.store.txPutBodyStructure(
			tx.Bucket(bodystructureBucket),
			message.ID(), bs,
		)
	}
	return message.store.db.Update(txUpdate)
}

// GetBodyStructure deserializes body structure from database. If body structure
// is not in database it returns nil error and nil body structure. If error
// occurs it returns nil body structure.
func (message *Message) GetBodyStructure() (bs *pkgMsg.BodyStructure, err error) {
	txRead := func(tx *bolt.Tx) error {
		bs, err = message.store.txGetBodyStructure(
			tx.Bucket(bodystructureBucket),
			message.ID(),
		)
		return err
	}
	if err = message.store.db.View(txRead); err != nil {
		return nil, err
	}
	return bs, nil
}

func (message *Message) IncreaseBuildCount() (times uint32, err error) {
	txUpdate := func(tx *bolt.Tx) error {
		times, err = message.store.txIncreaseMsgBuildCount(
			tx.Bucket(msgBuildCountBucket),
			message.ID(),
		)
		return err
	}
	if err = message.store.db.Update(txUpdate); err != nil {
		return 0, err
	}
	return times, nil
}
