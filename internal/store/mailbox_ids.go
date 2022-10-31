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
	"bytes"
	"math"
	"net/mail"
	"regexp"
	"strings"

	"github.com/ProtonMail/proton-bridge/v2/internal/imap/uidplus"
	"github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"
)

// GetAPIIDsFromUIDRange returns API IDs by IMAP UID range.
//
// API IDs are the long base64 strings that the API uses to identify messages.
// UIDs are unique increasing integers that must be unique within a mailbox.
func (storeMailbox *Mailbox) GetAPIIDsFromUIDRange(start, stop uint32) (apiIDs []string, err error) {
	err = storeMailbox.db().View(func(tx *bolt.Tx) error {
		b := storeMailbox.txGetIMAPIDsBucket(tx)
		c := b.Cursor()

		// GODT-1153 If the mailbox is empty we should reply BAD to client.
		if uid, _ := c.Last(); uid == nil {
			return nil
		}

		// If the start range is a wildcard, the range can only refer to the last message in the mailbox.
		if start == 0 {
			_, apiID := c.Last()
			apiIDs = append(apiIDs, string(apiID))
			return nil
		}

		// Resolve the stop value to be the final UID in the mailbox.
		if stop == 0 {
			stop = storeMailbox.txGetFinalUID(b)
		}

		// After resolving the stop value, it might be less than start so we sort it.
		if start > stop {
			start, stop = stop, start
		}

		startb := itob(start)
		stopb := itob(stop)

		for k, v := c.Seek(startb); k != nil && bytes.Compare(k, stopb) <= 0; k, v = c.Next() {
			apiIDs = append(apiIDs, string(v))
		}

		return nil
	})

	return apiIDs, err
}

// GetAPIIDsFromSequenceRange returns API IDs by IMAP sequence number range.
func (storeMailbox *Mailbox) GetAPIIDsFromSequenceRange(start, stop uint32) (apiIDs []string, err error) {
	err = storeMailbox.db().View(func(tx *bolt.Tx) error {
		b := storeMailbox.txGetIMAPIDsBucket(tx)
		c := b.Cursor()

		// GODT-1153 If the mailbox is empty we should reply BAD to client.
		if uid, _ := c.Last(); uid == nil {
			return nil
		}

		// If the start range is a wildcard, the range can only refer to the last message in the mailbox.
		if start == 0 {
			_, apiID := c.Last()
			apiIDs = append(apiIDs, string(apiID))
			return nil
		}

		var i uint32

		for k, v := c.First(); k != nil; k, v = c.Next() {
			i++

			if i < start {
				continue
			}

			if stop > 0 && i > stop {
				break
			}

			apiIDs = append(apiIDs, string(v))
		}

		if stop == 0 && len(apiIDs) == 0 {
			if _, apiID := c.Last(); len(apiID) > 0 {
				apiIDs = append(apiIDs, string(apiID))
			}
		}

		return nil
	})

	return apiIDs, err
}

// GetLatestAPIID returns the latest message API ID which still exists.
// Info: not the latest IMAP UID which can be already removed.
func (storeMailbox *Mailbox) GetLatestAPIID() (apiID string, err error) {
	err = storeMailbox.db().View(func(tx *bolt.Tx) error {
		c := storeMailbox.txGetAPIIDsBucket(tx).Cursor()
		lastAPIID, _ := c.Last()
		apiID = string(lastAPIID)
		if apiID == "" {
			return errors.New("cannot get latest API ID: empty mailbox")
		}
		return nil
	})
	return
}

// GetNextUID returns the next IMAP UID.
func (storeMailbox *Mailbox) GetNextUID() (uid uint32, err error) {
	err = storeMailbox.db().View(func(tx *bolt.Tx) error {
		b := storeMailbox.txGetIMAPIDsBucket(tx)
		uid, err = storeMailbox.txGetNextUID(b, false)
		return err
	})
	return
}

func (storeMailbox *Mailbox) txGetNextUID(imapIDBucket *bolt.Bucket, write bool) (uint32, error) {
	var uid uint64
	var err error
	if write {
		uid, err = imapIDBucket.NextSequence()
		if err != nil {
			return 0, err
		}
	} else {
		uid = imapIDBucket.Sequence() + 1
	}
	if math.MaxUint32 <= uid {
		return 0, errors.New("too large sequence number")
	}
	return uint32(uid), nil
}

// getUID returns IMAP UID in this mailbox for message ID.
func (storeMailbox *Mailbox) getUID(apiID string) (uid uint32, err error) {
	err = storeMailbox.db().View(func(tx *bolt.Tx) error {
		uid, err = storeMailbox.txGetUID(tx, apiID)
		return err
	})
	return
}

func (storeMailbox *Mailbox) txGetUID(tx *bolt.Tx, apiID string) (uint32, error) {
	return storeMailbox.txGetUIDFromBucket(storeMailbox.txGetAPIIDsBucket(tx), apiID)
}

// txGetUIDFromBucket expects pointer to API bucket.
func (storeMailbox *Mailbox) txGetUIDFromBucket(b *bolt.Bucket, apiID string) (uint32, error) {
	v := b.Get([]byte(apiID))
	if v == nil {
		return 0, ErrNoSuchAPIID
	}
	return btoi(v), nil
}

// GetDeletedAPIIDs returns API IDs in this mailbox for message ID.
func (storeMailbox *Mailbox) GetDeletedAPIIDs() (apiIDs []string, err error) {
	err = storeMailbox.db().Update(func(tx *bolt.Tx) error {
		b := storeMailbox.txGetDeletedIDsBucket(tx)
		c := b.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			apiIDs = append(apiIDs, string(k))
		}
		return nil
	})
	return
}

// getSequenceNumber returns IMAP sequence number in the mailbox for the message with the given API ID `apiID`.
func (storeMailbox *Mailbox) getSequenceNumber(apiID string) (seqNum uint32, err error) {
	err = storeMailbox.db().View(func(tx *bolt.Tx) error {
		b := storeMailbox.txGetIMAPIDsBucket(tx)
		uid, err := storeMailbox.txGetUID(tx, apiID)
		if err != nil {
			return err
		}
		seqNum, err = storeMailbox.txGetSequenceNumberOfUID(b, itob(uid))
		return err
	})
	return
}

// txGetSequenceNumberOfUID returns the IMAP sequence number of the message
// with the given IMAP UID bytes `uidb`.
//
// NOTE: The `bolt.Cursor.Next()` loops in order of ascending key bytes. The
// IMAP UID bucket is ordered by increasing UID because it's using BigEndian to
// encode uint into byte. Hence the sequence number (IMAP ID) corresponds to
// position of uid key in this order.
func (storeMailbox *Mailbox) txGetSequenceNumberOfUID(bucket *bolt.Bucket, uidb []byte) (uint32, error) {
	seqNum := uint32(0)
	c := bucket.Cursor()

	// Speed up for the case of last message. This is always true for
	// adding new message. It will return number of keys in bucket because
	// sequence number starts with 1.
	// We cannot use bucket.Stats() for that--it doesn't work in the same
	// transaction because stats are updated when transaction is committed.
	// But we can at least optimise to not do equal for all keys.
	lastKey, _ := c.Last()
	isLast := bytes.Equal(lastKey, uidb)

	for k, _ := c.First(); k != nil; k, _ = c.Next() {
		seqNum++ // Sequence number starts at 1.
		if isLast {
			continue
		}
		if bytes.Equal(k, uidb) {
			return seqNum, nil
		}
	}

	if isLast {
		return seqNum, nil
	}

	return 0, ErrNoSuchUID
}

// GetUIDList returns UID list corresponding to messageIDs in a requested order.
func (storeMailbox *Mailbox) GetUIDList(apiIDs []string) *uidplus.OrderedSeq {
	seqSet := &uidplus.OrderedSeq{}
	_ = storeMailbox.db().View(func(tx *bolt.Tx) error {
		b := storeMailbox.txGetAPIIDsBucket(tx)
		for _, apiID := range apiIDs {
			v := b.Get([]byte(apiID))
			if v == nil {
				storeMailbox.log.
					WithField("msgID", apiID).
					Warn("Cannot find UID")
				continue
			}

			seqSet.Add(btoi(v))
		}
		return nil
	})
	return seqSet
}

// GetUIDByHeader returns UID of message existing in mailbox or zero if no match found.
func (storeMailbox *Mailbox) GetUIDByHeader(header *mail.Header) (foundUID uint32) {
	if header == nil {
		return uint32(0)
	}

	// Message-Id in appended-after-send mail is processed as ExternalID
	// in PM message. Message-Id in normal copy/move will be the PM internal ID.
	messageID := header.Get("Message-Id")

	// There is nothing to find, when no Message-Id given.
	if messageID == "" {
		return uint32(0)
	}

	// The most often situation is that message is APPENDed after it was sent so the
	// Message-ID will be reflected by ExternalID in API message meta-data.
	externalID := strings.Trim(messageID, "<> ") // remove '<>' to improve match
	matchExternalID := regexp.MustCompile(`"ExternalID":"` +
		` *(\\u003c)? *` + // \u003c is equivalent to `<`
		regexp.QuoteMeta(externalID) +
		` *(\\u003e)? *` + // \u0033 is equivalent to `>`
		`"`,
	)

	// It is possible that client will try to COPY existing message to Sent
	// using APPEND command. In that case the Message-Id from header will
	// be internal message ID and we need to check whether it's already there.
	matchInternalID := bytes.Split([]byte(externalID), []byte("@"))[0]

	_ = storeMailbox.db().View(func(tx *bolt.Tx) error {
		metaBucket := tx.Bucket(metadataBucket)
		b := storeMailbox.txGetIMAPIDsBucket(tx)
		c := b.Cursor()
		imapID, apiID := c.Last()
		for ; imapID != nil; imapID, apiID = c.Prev() {
			rawMeta := metaBucket.Get(apiID)
			if rawMeta == nil {
				storeMailbox.log.
					WithField("IMAP-UID", imapID).
					WithField("API-ID", apiID).
					Warn("Cannot find meta-data while searching for externalID")
				continue
			}

			if !matchExternalID.Match(rawMeta) && !bytes.Equal(apiID, matchInternalID) {
				continue
			}

			foundUID = btoi(imapID)
			return nil
		}
		return nil
	})

	return foundUID
}

func (storeMailbox *Mailbox) txGetFinalUID(b *bolt.Bucket) uint32 {
	uid, _ := b.Cursor().Last()

	if uid == nil {
		// This happened most probably due to empty mailbox and whole
		// store needs to be re-initialize in order to fix it.
		panic(errors.New("cannot get final UID"))
	}

	return btoi(uid)
}
