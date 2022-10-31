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
	"encoding/json"
	"fmt"
	"strings"
	"sync/atomic"

	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
	"github.com/sirupsen/logrus"
	bolt "go.etcd.io/bbolt"
)

// Mailbox is mailbox for specific address and mailbox.
type Mailbox struct {
	store        *Store
	storeAddress *Address

	labelID     string
	labelPrefix string
	labelName   string
	color       string

	log *logrus.Entry

	isDeleting atomic.Value
}

func newMailbox(storeAddress *Address, labelID, labelPrefix, labelName, color string) (mb *Mailbox, err error) {
	err = storeAddress.store.db.Update(func(tx *bolt.Tx) error {
		mb, err = txNewMailbox(tx, storeAddress, labelID, labelPrefix, labelName, color)
		return err
	})
	return
}

func txNewMailbox(tx *bolt.Tx, storeAddress *Address, labelID, labelPrefix, labelName, color string) (*Mailbox, error) {
	l := log.WithField("addrID", storeAddress.addressID).WithField("labelID", labelID)
	mb := &Mailbox{
		store:        storeAddress.store,
		storeAddress: storeAddress,
		labelID:      labelID,
		labelPrefix:  labelPrefix,
		labelName:    labelPrefix + labelName,
		color:        color,
		log:          l,
	}
	mb.isDeleting.Store(false)

	err := initMailboxBucket(tx, mb.getBucketName())
	if err != nil {
		l.WithError(err).Error("Could not initialise mailbox buckets")
	}

	syncDraftsIfNecssary(tx, mb)

	return mb, err
}

func syncDraftsIfNecssary(tx *bolt.Tx, mb *Mailbox) { //nolint:funlen
	// We didn't support drafts before v1.2.6 and therefore if we now created
	// Drafts mailbox we need to check whether counts match (drafts are synced).
	// If not, sync them from local metadata without need to do full resync,
	// Can be removed with 1.2.7 or later.
	if mb.labelID != pmapi.DraftLabel {
		return
	}

	// If the drafts mailbox total is non-zero, it means it has already been used
	// and there is no need to continue. Otherwise, we may need to do an initial sync.
	total, _, _, err := mb.txGetCounts(tx)
	if err != nil || total != 0 {
		return
	}

	counts, err := mb.store.txGetOnAPICounts(tx)
	if err != nil {
		return
	}

	foundCounts := false
	doSync := false
	for _, count := range counts {
		if count.LabelID != pmapi.DraftLabel {
			continue
		}
		foundCounts = true
		log.WithField("total", total).WithField("total-api", count.TotalOnAPI).Debug("Drafts mailbox created: checking need for sync")
		if count.TotalOnAPI == total {
			continue
		}
		doSync = true
		break
	}

	if !foundCounts {
		log.Debug("Drafts mailbox created: missing counts, refreshing")
		_ = mb.store.updateCountsFromServer()
	}

	if !foundCounts || doSync {
		err := tx.Bucket(metadataBucket).ForEach(func(k, v []byte) error {
			msg := &pmapi.Message{}
			if err := json.Unmarshal(v, msg); err != nil {
				return err
			}
			for _, msgLabelID := range msg.LabelIDs {
				if msgLabelID == pmapi.DraftLabel {
					log.WithField("id", msg.ID).Trace("Drafts mailbox created: syncing draft locally")
					_ = mb.txCreateOrUpdateMessages(tx, []*pmapi.Message{msg})
					break
				}
			}
			return nil
		})
		log.WithError(err).Info("Drafts mailbox created: synced localy")
	}
}

func initMailboxBucket(tx *bolt.Tx, bucketName []byte) error {
	bucket, err := tx.Bucket(mailboxesBucket).CreateBucketIfNotExists(bucketName)
	if err != nil {
		return err
	}

	if _, err := bucket.CreateBucketIfNotExists(imapIDsBucket); err != nil {
		return err
	}
	if _, err := bucket.CreateBucketIfNotExists(apiIDsBucket); err != nil {
		return err
	}
	if _, err := bucket.CreateBucketIfNotExists(deletedIDsBucket); err != nil {
		return err
	}

	return nil
}

// LabelID returns ID of mailbox.
func (storeMailbox *Mailbox) LabelID() string {
	return storeMailbox.labelID
}

// Name returns the name of mailbox.
func (storeMailbox *Mailbox) Name() string {
	return storeMailbox.labelName
}

// Color returns the color of mailbox.
func (storeMailbox *Mailbox) Color() string {
	return storeMailbox.color
}

// UIDValidity returns the current value of structure version.
func (storeMailbox *Mailbox) UIDValidity() uint32 {
	return storeMailbox.store.getMailboxesVersion()
}

// IsFolder returns whether the mailbox is a folder (has "Folders/" prefix).
func (storeMailbox *Mailbox) IsFolder() bool {
	return storeMailbox.labelPrefix == UserFoldersPrefix
}

// IsLabel returns whether the mailbox is a label (has "Labels/" prefix).
func (storeMailbox *Mailbox) IsLabel() bool {
	return storeMailbox.labelPrefix == UserLabelsPrefix
}

// IsSystem returns whether the mailbox is one of the specific system mailboxes (has no prefix).
func (storeMailbox *Mailbox) IsSystem() bool {
	return storeMailbox.labelPrefix == ""
}

// Rename updates the mailbox by calling an API.
// Change has to be propagated to all the same mailboxes in all addresses.
// The propagation is processed by the event loop.
func (storeMailbox *Mailbox) Rename(newName string) error {
	if storeMailbox.IsSystem() {
		return fmt.Errorf("cannot rename system mailboxes")
	}

	if storeMailbox.IsFolder() {
		if !strings.HasPrefix(newName, UserFoldersPrefix) {
			return fmt.Errorf("cannot rename folder to non-folder")
		}

		newName = strings.TrimPrefix(newName, UserFoldersPrefix)
	}

	if storeMailbox.IsLabel() {
		if !strings.HasPrefix(newName, UserLabelsPrefix) {
			return fmt.Errorf("cannot rename label to non-label")
		}

		newName = strings.TrimPrefix(newName, UserLabelsPrefix)
	}

	return storeMailbox.storeAddress.updateMailbox(storeMailbox.labelID, newName, storeMailbox.color)
}

// Delete deletes the mailbox by calling an API.
// Deletion has to be propagated to all the same mailboxes in all addresses.
// The propagation is processed by the event loop.
func (storeMailbox *Mailbox) Delete() error {
	storeMailbox.isDeleting.Store(true)
	return storeMailbox.storeAddress.deleteMailbox(storeMailbox.labelID)
}

// GetDelimiter returns the path separator.
func (storeMailbox *Mailbox) GetDelimiter() string {
	return PathDelimiter
}

// deleteMailboxEvent deletes the mailbox bucket.
// This is called from the event loop.
func (storeMailbox *Mailbox) deleteMailboxEvent() error {
	if !storeMailbox.isDeleting.Load().(bool) { //nolint:forcetypeassert
		// Deleting label removes bucket. Any ongoing connection selected
		// in such mailbox then might panic because of non-existing bucket.
		// Closing connetions prevents that panic but if the connection
		// asked for deletion, it should not be closed so it can receive
		// successful response.
		storeMailbox.store.user.CloseAllConnections()
	}
	return storeMailbox.db().Update(func(tx *bolt.Tx) error {
		return tx.Bucket(mailboxesBucket).DeleteBucket(storeMailbox.getBucketName())
	})
}

// txGetIMAPIDsBucket returns the bucket mapping IMAP ID to API ID.
func (storeMailbox *Mailbox) txGetIMAPIDsBucket(tx *bolt.Tx) *bolt.Bucket {
	return storeMailbox.txGetBucket(tx).Bucket(imapIDsBucket)
}

// txGetAPIIDsBucket returns the bucket mapping API ID to IMAP ID.
func (storeMailbox *Mailbox) txGetAPIIDsBucket(tx *bolt.Tx) *bolt.Bucket {
	return storeMailbox.txGetBucket(tx).Bucket(apiIDsBucket)
}

// txGetDeletedIDsBucket returns the bucket with messagesID marked as deleted.
func (storeMailbox *Mailbox) txGetDeletedIDsBucket(tx *bolt.Tx) *bolt.Bucket {
	return storeMailbox.txGetBucket(tx).Bucket(deletedIDsBucket)
}

// txGetBucket returns the bucket of mailbox containing mapping buckets.
func (storeMailbox *Mailbox) txGetBucket(tx *bolt.Tx) *bolt.Bucket {
	return tx.Bucket(mailboxesBucket).Bucket(storeMailbox.getBucketName())
}

func getMailboxBucketName(addressID, labelID string) []byte {
	return []byte(addressID + "-" + labelID)
}

// getBucketName returns the name of mailbox bucket.
func (storeMailbox *Mailbox) getBucketName() []byte {
	return getMailboxBucketName(storeMailbox.storeAddress.addressID, storeMailbox.labelID)
}

// pollNow is a proxy for the store's eventloop's `pollNow()`.
func (storeMailbox *Mailbox) pollNow() {
	storeMailbox.store.eventLoop.pollNow()
}

// api is a proxy for the store's `PMAPIProvider`.
func (storeMailbox *Mailbox) client() pmapi.Client {
	return storeMailbox.store.client()
}

// update is a proxy for the store's db's `Update`.
func (storeMailbox *Mailbox) db() *bolt.DB {
	return storeMailbox.store.db
}
