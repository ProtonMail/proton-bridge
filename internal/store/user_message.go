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
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/mail"
	"net/textproto"
	"strings"

	pmcrypto "github.com/ProtonMail/gopenpgp/crypto"
	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	bolt "go.etcd.io/bbolt"
)

// CreateDraft creates draft with attachments.
// If `attachedPublicKey` is passed, it's added to attachments.
// Both draft and attachments are encrypted with passed `kr` key.
func (store *Store) CreateDraft(
	kr *pmcrypto.KeyRing,
	message *pmapi.Message,
	attachmentReaders []io.Reader,
	attachedPublicKey,
	attachedPublicKeyName string,
	parentID string) (*pmapi.Message, []*pmapi.Attachment, error) {
	defer store.eventLoop.pollNow()

	// Since this is a draft, we don't need to sign it.
	if err := message.Encrypt(kr, nil); err != nil {
		return nil, nil, errors.Wrap(err, "failed to encrypt draft")
	}

	attachments := message.Attachments
	message.Attachments = nil

	draftAction := store.getDraftAction(message)
	draft, err := store.api.CreateDraft(message, parentID, draftAction)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to create draft")
	}

	if attachedPublicKey != "" {
		attachmentReaders = append(attachmentReaders, strings.NewReader(attachedPublicKey))
		publicKeyAttachment := &pmapi.Attachment{
			Name:     attachedPublicKeyName + ".asc",
			MIMEType: "application/pgp-key",
			Header:   textproto.MIMEHeader{},
		}
		attachments = append(attachments, publicKeyAttachment)
	}

	for idx, attachment := range attachments {
		attachment.MessageID = draft.ID
		attachmentBody, _ := ioutil.ReadAll(attachmentReaders[idx])

		createdAttachment, err := store.createAttachment(kr, attachment, attachmentBody)
		if err != nil {
			return nil, nil, errors.Wrap(err, "failed to create attachment for draft")
		}

		attachments[idx] = createdAttachment
	}

	return draft, attachments, nil
}

func (store *Store) getDraftAction(message *pmapi.Message) int {
	// If not a reply, must be a forward.
	if len(message.Header["In-Reply-To"]) == 0 {
		return pmapi.DraftActionForward
	}
	return pmapi.DraftActionReply
}

func (store *Store) createAttachment(kr *pmcrypto.KeyRing, attachment *pmapi.Attachment, attachmentBody []byte) (*pmapi.Attachment, error) {
	r := bytes.NewReader(attachmentBody)
	sigReader, err := attachment.DetachedSign(kr, r)
	if err != nil {
		return nil, errors.Wrap(err, "failed to sign attachment")
	}

	r = bytes.NewReader(attachmentBody)
	encReader, err := attachment.Encrypt(kr, r)
	if err != nil {
		return nil, errors.Wrap(err, "failed to encrypt attachment")
	}

	createdAttachment, err := store.api.CreateAttachment(attachment, encReader, sigReader)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create attachment")
	}

	return createdAttachment, nil
}

// SendMessage sends the message.
func (store *Store) SendMessage(messageID string, req *pmapi.SendMessageReq) error {
	defer store.eventLoop.pollNow()
	_, _, err := store.api.SendMessage(messageID, req)
	return err
}

// getAllMessageIDs returns all API IDs of messages in the local database.
func (store *Store) getAllMessageIDs() (apiIDs []string, err error) {
	err = store.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(metadataBucket)
		return b.ForEach(func(k, v []byte) error {
			apiIDs = append(apiIDs, string(k))
			return nil
		})
	})
	return
}

// getMessageFromDB returns pmapi struct of message by API ID.
func (store *Store) getMessageFromDB(apiID string) (msg *pmapi.Message, err error) {
	err = store.db.View(func(tx *bolt.Tx) error {
		msg, err = store.txGetMessage(tx, apiID)
		return err
	})

	return
}

func (store *Store) txGetMessage(tx *bolt.Tx, apiID string) (*pmapi.Message, error) {
	b := tx.Bucket(metadataBucket)

	msgb := b.Get([]byte(apiID))
	if msgb == nil {
		return nil, ErrNoSuchAPIID
	}
	msg := &pmapi.Message{}
	if err := json.Unmarshal(msgb, msg); err != nil {
		return nil, err
	}
	return msg, nil
}

func (store *Store) txPutMessage(metaBucket *bolt.Bucket, onlyMeta *pmapi.Message) error {
	b, err := json.Marshal(onlyMeta)
	if err != nil {
		return errors.Wrap(err, "cannot marshall metadata")
	}
	err = metaBucket.Put([]byte(onlyMeta.ID), b)
	if err != nil {
		return errors.Wrap(err, "cannot add to metadata bucket")
	}
	return nil
}

// createOrUpdateMessageEvent is helper to create only one message with
// createOrUpdateMessagesEvent.
func (store *Store) createOrUpdateMessageEvent(msg *pmapi.Message) error {
	return store.createOrUpdateMessagesEvent([]*pmapi.Message{msg})
}

// createOrUpdateMessagesEvent tries to create or update messages in database.
// This function is optimised for insertion of many messages at once.
// It calls createLabelsIfMissing if needed.
func (store *Store) createOrUpdateMessagesEvent(msgs []*pmapi.Message) error { //nolint[funlen]
	store.log.WithField("msgs", msgs).Trace("Creating or updating messages in the store")

	// Strip non meta first to reduce memory (no need to keep all old msg ID data during update).
	err := store.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(metadataBucket)
		for _, msg := range msgs {
			clearNonMetadata(msg)
			txUpdateMetadaFromDB(b, msg, store.log)
		}
		return nil
	})
	if err != nil {
		return err
	}

	affectedLabels := map[string]bool{}
	for _, m := range msgs {
		for _, l := range m.LabelIDs {
			affectedLabels[l] = true
		}
	}
	if err = store.createLabelsIfMissing(affectedLabels); err != nil {
		return err
	}

	// Updating metadata and mailboxes is not atomic, but this is OK.
	// The worst case scenario is we have metadata but not updated mailboxes
	// which is OK as without information in mailboxes IMAP we will never ask
	// for metadata. Also, when doing the operation again, it will simply
	// update the metadata.
	// The reason to split is efficiency--it's more memory efficient.

	// Update metadata.
	err = store.db.Update(func(tx *bolt.Tx) error {
		metaBucket := tx.Bucket(metadataBucket)
		for _, msg := range msgs {
			err := store.txPutMessage(metaBucket, msg)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	// Update mailboxes.
	err = store.db.Update(func(tx *bolt.Tx) error {
		for _, a := range store.addresses {
			if err := a.txCreateOrUpdateMessages(tx, msgs); err != nil {
				store.log.WithError(err).Error("cannot update maiboxes")
				return errors.Wrap(err, "cannot add to mailboxes bucket")
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

// clearNonMetadata to not allow to store decrypted or encrypted data i.e. body
// and attachments.
func clearNonMetadata(onlyMeta *pmapi.Message) {
	onlyMeta.Body = ""
	onlyMeta.Attachments = nil
}

// txUpdateMetadaFromDB changes the the onlyMeta data.
// If there is stored message in metaBucket the size, header and MIMEType are
// not changed if already set. To change these:
// * size must be updated by Message.SetSize
// * contentType and header must be updated by Message.SetContentTypeAndHeader
func txUpdateMetadaFromDB(metaBucket *bolt.Bucket, onlyMeta *pmapi.Message, log *logrus.Entry) {
	// Size attribute on the server is counting encrypted data. We need to compute
	// "real" size of decrypted data. Negative values will be processed during fetch.
	onlyMeta.Size = -1

	msgb := metaBucket.Get([]byte(onlyMeta.ID))
	if msgb == nil {
		return
	}

	// It is faster to unmarshal only the needed items.
	stored := &struct {
		Size     int64
		Header   string
		MIMEType string
	}{}
	if err := json.Unmarshal(msgb, stored); err != nil {
		log.WithError(err).
			Error("Fail to unmarshal from DB, metadata will be overwritten")
		return
	}

	// Keep already calculated size and content type.
	onlyMeta.Size = stored.Size
	onlyMeta.MIMEType = stored.MIMEType
	if stored.Header != "" && stored.Header != "(No Header)" {
		tmpMsg, err := mail.ReadMessage(
			strings.NewReader(stored.Header + "\r\n\r\n"),
		)
		if err == nil {
			onlyMeta.Header = tmpMsg.Header
		} else {
			log.WithError(err).
				Error("Fail to parse, the header will be overwritten")
		}
	}
}

// deleteMessageEvent is helper to delete only one message with deleteMessagesEvent.
func (store *Store) deleteMessageEvent(apiID string) error {
	return store.deleteMessagesEvent([]string{apiID})
}

// deleteMessagesEvent deletes the message from metadata and all mailbox buckets.
func (store *Store) deleteMessagesEvent(apiIDs []string) error {
	return store.db.Update(func(tx *bolt.Tx) error {
		for _, apiID := range apiIDs {
			if err := tx.Bucket(metadataBucket).Delete([]byte(apiID)); err != nil {
				return err
			}

			for _, a := range store.addresses {
				if err := a.txDeleteMessage(tx, apiID); err != nil {
					return err
				}
			}
		}
		return nil
	})
}
