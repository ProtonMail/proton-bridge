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
	"encoding/json"
	"sort"

	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
	"github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"
)

// GetCounts returns numbers of total and unread messages in this mailbox bucket.
func (storeMailbox *Mailbox) GetCounts() (total, unread, unseenSeqNum uint, err error) {
	err = storeMailbox.db().View(func(tx *bolt.Tx) error {
		total, unread, unseenSeqNum, err = storeMailbox.txGetCounts(tx)
		return err
	})
	return
}

func (storeMailbox *Mailbox) txGetCounts(tx *bolt.Tx) (total, unread, unseenSeqNum uint, err error) {
	// For total it would be enough to use `bolt.Bucket.Stats().KeyN` but
	// we also need to retrieve the count of unread emails therefore we are
	// looping all messages in this mailbox by `bolt.Cursor`
	metaBucket := tx.Bucket(metadataBucket)
	b := storeMailbox.txGetIMAPIDsBucket(tx)
	c := b.Cursor()
	imapID, apiID := c.First()
	for ; imapID != nil; imapID, apiID = c.Next() {
		total++
		rawMsg := metaBucket.Get(apiID)
		if rawMsg == nil {
			return 0, 0, 0, ErrNoSuchAPIID
		}
		// Do not unmarshal whole JSON to speed up the looping.
		// Instead, we assume it will contain JSON int field `Unread`
		// where `1` means true (i.e. message is unread)
		if bytes.Contains(rawMsg, []byte(`"Unread":1`)) {
			if unseenSeqNum == 0 {
				unseenSeqNum = total
			}
			unread++
		}
	}
	return total, unread, unseenSeqNum, err
}

type mailboxCounts struct {
	LabelID     string
	LabelName   string
	Color       string
	Order       int
	IsFolder    bool
	TotalOnAPI  uint
	UnreadOnAPI uint
}

func txGetCountsFromBucketOrNew(bkt *bolt.Bucket, labelID string) (*mailboxCounts, error) {
	mc := &mailboxCounts{}
	if mcJSON := bkt.Get([]byte(labelID)); mcJSON != nil {
		if err := json.Unmarshal(mcJSON, mc); err != nil {
			return nil, err
		}
	}
	mc.LabelID = labelID // if it was empty before we need to set labelID

	return mc, nil
}

func (mc *mailboxCounts) txWriteToBucket(bucket *bolt.Bucket) error {
	mcJSON, err := json.Marshal(mc)
	if err != nil {
		return err
	}
	return bucket.Put([]byte(mc.LabelID), mcJSON)
}

func getSystemFolders() []*mailboxCounts {
	return []*mailboxCounts{
		{pmapi.InboxLabel, "INBOX", "#000", -1000, true, 0, 0},
		{pmapi.SentLabel, "Sent", "#000", -9, true, 0, 0},
		{pmapi.ArchiveLabel, "Archive", "#000", -8, true, 0, 0},
		{pmapi.SpamLabel, "Spam", "#000", -7, true, 0, 0},
		{pmapi.TrashLabel, "Trash", "#000", -6, true, 0, 0},
		{pmapi.AllMailLabel, "All Mail", "#000", -5, true, 0, 0},
		{pmapi.DraftLabel, "Drafts", "#000", -4, true, 0, 0},
	}
}

// skipThisLabel decides to skip labelIDs that *are* pmapi system labels but *aren't* local system labels
// (i.e. if it's in `pmapi.SystemLabels` but not in `getSystemFolders` then we skip it, otherwise we don't).
func skipThisLabel(labelID string) bool {
	switch labelID {
	case pmapi.StarredLabel, pmapi.AllSentLabel, pmapi.AllDraftsLabel:
		return true
	}
	return false
}

func sortByOrder(labels []*pmapi.Label) {
	sort.Slice(labels, func(i, j int) bool {
		return labels[i].Order < labels[j].Order
	})
}

func (mc *mailboxCounts) getPMLabel() *pmapi.Label {
	return &pmapi.Label{
		ID:        mc.LabelID,
		Name:      mc.LabelName,
		Path:      mc.LabelName,
		Color:     mc.Color,
		Order:     mc.Order,
		Type:      pmapi.LabelTypeMailBox,
		Exclusive: pmapi.Boolean(mc.IsFolder),
	}
}

// createOrUpdateMailboxCountsBuckets will not change the on-API-counts.
func (store *Store) createOrUpdateMailboxCountsBuckets(labels []*pmapi.Label) error {
	// Don't forget about system folders.
	// It should set label id, name, color, isFolder, total, unread.
	tx := func(tx *bolt.Tx) error {
		countsBkt := tx.Bucket(countsBucket)
		for _, label := range labels {
			// Skipping is probably not necessary.
			if skipThisLabel(label.ID) {
				continue
			}

			// Get current data.
			mailbox, err := txGetCountsFromBucketOrNew(countsBkt, label.ID)
			if err != nil {
				return err
			}

			// Update mailbox info, but dont change on-API-counts.
			mailbox.LabelName = label.Path
			mailbox.Color = label.Color
			mailbox.Order = label.Order
			mailbox.IsFolder = bool(label.Exclusive)

			// Write.
			if err = mailbox.txWriteToBucket(countsBkt); err != nil {
				return err
			}
		}
		return nil
	}

	return store.db.Update(tx)
}

func (store *Store) getLabelsFromLocalStorage() ([]*pmapi.Label, error) {
	countsOnAPI, err := store.getOnAPICounts()
	if err != nil {
		return nil, err
	}
	labels := []*pmapi.Label{}
	for _, counts := range countsOnAPI {
		labels = append(labels, counts.getPMLabel())
	}
	sortByOrder(labels)

	return labels, nil
}

func (store *Store) getOnAPICounts() (counts []*mailboxCounts, err error) {
	err = store.db.View(func(tx *bolt.Tx) error {
		counts, err = store.txGetOnAPICounts(tx)
		return err
	})
	return
}

func (store *Store) txGetOnAPICounts(tx *bolt.Tx) ([]*mailboxCounts, error) {
	counts := []*mailboxCounts{}
	c := tx.Bucket(countsBucket).Cursor()
	for k, countsB := c.First(); k != nil; k, countsB = c.Next() {
		l := store.log.WithField("key", string(k))
		if countsB == nil {
			err := errors.New("empty counts in DB")
			l.WithError(err).Error("While getting local labels")
			return nil, err
		}

		mbCounts := &mailboxCounts{}
		if err := json.Unmarshal(countsB, mbCounts); err != nil {
			l.WithError(err).Error("While unmarshaling local labels")
			return nil, err
		}

		counts = append(counts, mbCounts)
	}
	return counts, nil
}

// createOrUpdateOnAPICounts will change only on-API-counts.
func (store *Store) createOrUpdateOnAPICounts(mailboxCountsOnAPI []*pmapi.MessagesCount) error {
	store.log.Debug("Updating API counts")

	tx := func(tx *bolt.Tx) error {
		countsBkt := tx.Bucket(countsBucket)
		for _, countsOnAPI := range mailboxCountsOnAPI {
			if skipThisLabel(countsOnAPI.LabelID) {
				continue
			}

			// Get current data.
			counts, err := txGetCountsFromBucketOrNew(countsBkt, countsOnAPI.LabelID)
			if err != nil {
				return err
			}

			// Update only counts.
			counts.TotalOnAPI = uint(countsOnAPI.Total)
			counts.UnreadOnAPI = uint(countsOnAPI.Unread)

			if err = counts.txWriteToBucket(countsBkt); err != nil {
				return err
			}
		}

		return nil
	}

	return store.db.Update(tx)
}

func (store *Store) removeMailboxCount(labelID string) error {
	err := store.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(countsBucket).Delete([]byte(labelID))
	})
	if err != nil {
		store.log.WithError(err).
			WithField("labelID", labelID).
			Warning("Cannot remove counts")
	}
	return err
}
