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

package store

import (
	"net/mail"

	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
	bolt "go.etcd.io/bbolt"
)

// Message is wrapper around `pmapi.Message` with connection to
// a specific mailbox with helper functions to get IMAP UID, sequence
// numbers and similar.
type Message struct {
	api PMAPIProvider
	msg *pmapi.Message

	store        *Store
	storeMailbox *Mailbox
}

func newStoreMessage(storeMailbox *Mailbox, msg *pmapi.Message) *Message {
	return &Message{
		api:          storeMailbox.store.api,
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
