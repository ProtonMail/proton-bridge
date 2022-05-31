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
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	bolt "go.etcd.io/bbolt"
)

const (
	syncFinishTimeKey     = "sync_state" // The original key was sync_state and we want to keep compatibility.
	syncIDRangesKey       = "id_ranges"
	syncIDsToBeDeletedKey = "ids_to_be_deleted"
)

// updateCountsFromServer will download and set the counts.
func (store *Store) updateCountsFromServer() error {
	counts, err := store.client().CountMessages(context.Background(), "")
	if err != nil {
		return errors.Wrap(err, "cannot update counts from server")
	}

	return store.createOrUpdateOnAPICounts(counts)
}

// isSynced checks whether DB counts are synced with provided counts from API.
func (store *Store) isSynced(countsOnAPI []*pmapi.MessagesCount) (bool, error) {
	store.log.WithField("apiCounts", countsOnAPI).Debug("Checking whether store is synced")

	// IMPORTANT: The countsOnAPI can contain duplicates due to event merge
	// (ie one label can be present multiple times). It is important to
	// process all counts before checking whether they are synced.
	if err := store.createOrUpdateOnAPICounts(countsOnAPI); err != nil {
		store.log.WithError(err).Error("Cannot update counts before check sync")
		return false, err
	}

	allCounts, err := store.getOnAPICounts()
	if err != nil {
		return false, err
	}

	store.lock.Lock()
	defer store.lock.Unlock()

	countsAreOK := true
	for _, counts := range allCounts {
		total, unread := uint(0), uint(0)
		for _, address := range store.addresses {
			mbox, err := address.getMailboxByID(counts.LabelID)
			if err != nil {
				return false, errors.Wrapf(
					err,
					"cannot find mailbox for address %q",
					address.addressID,
				)
			}

			mboxTot, mboxUnread, _, err := mbox.GetCounts()
			if err != nil {
				errW := errors.Wrap(err, "cannot count messages")
				store.log.
					WithError(errW).
					WithField("label", counts.LabelID).
					WithField("address", address.addressID).
					Error("IsSynced failed")
				return false, err
			}
			total += mboxTot
			unread += mboxUnread
		}

		if total != counts.TotalOnAPI || unread != counts.UnreadOnAPI {
			store.log.WithFields(logrus.Fields{
				"label":      counts.LabelID,
				"db-total":   total,
				"db-unread":  unread,
				"api-total":  counts.TotalOnAPI,
				"api-unread": counts.UnreadOnAPI,
			}).Warning("counts differ")
			countsAreOK = false
		}
	}

	return countsAreOK, nil
}

// triggerSync starts a sync of complete user by syncing All Mail mailbox.
// All Mail mailbox contains all messages, so we download all meta data needed
// to generate any address/mailbox IMAP UIDs.
// Sync state can be in three states:
//  * Nothing in database. For example when user logs in for the first time.
//    `triggerSync` will start full sync.
//  * Database has syncIDRangesKey and syncIDsToBeDeletedKey keys with data.
//    Sync is in progress or was interrupted. In later case when, `triggerSync`
//    will continue where it left off.
//  * Database has only syncStateKey with time when database was last synced.
//    `triggerSync` will reset it and start full sync again.
func (store *Store) triggerSync() {
	syncState := store.loadSyncState()

	// We first clear the last sync state in case this sync fails.
	syncState.clearFinishTime()

	// We don't want sync to block.
	go func() {
		defer store.panicHandler.HandlePanic()

		store.log.Debug("Store sync triggered")

		store.lock.Lock()

		if store.isSyncRunning {
			store.lock.Unlock()
			store.log.Info("Store sync is already ongoing")
			return
		}

		if store.syncCooldown.isTooSoon() {
			store.lock.Unlock()
			store.log.Info("Skipping sync: store tries to resync too often")
			return
		}

		store.isSyncRunning = true
		store.lock.Unlock()

		defer func() {
			store.lock.Lock()
			store.isSyncRunning = false
			store.lock.Unlock()
		}()

		store.log.WithField("isIncomplete", syncState.isIncomplete()).Info("Store sync started")

		err := syncAllMail(store.panicHandler, store, store.client(), syncState)
		if err != nil {
			log.WithError(err).Error("Store sync failed")
			store.syncCooldown.increaseWaitTime()
			return
		}

		store.syncCooldown.reset()
		syncState.setFinishTime()
	}()
}

// isSyncFinished returns whether the database has finished a sync.
func (store *Store) isSyncFinished() (isSynced bool) {
	return store.loadSyncState().isFinished()
}

// loadSyncState loads information about sync from database.
// See `triggerSync` to learn more about possible states.
func (store *Store) loadSyncState() *syncState {
	finishTime := int64(0)
	idRanges := []*syncIDRange{}
	idsToBeDeleted := []string{}

	err := store.db.View(func(tx *bolt.Tx) (err error) {
		b := tx.Bucket(syncStateBucket)

		finishTimeByte := b.Get([]byte(syncFinishTimeKey))
		if finishTimeByte != nil {
			finishTime, err = strconv.ParseInt(string(finishTimeByte), 10, 64)
			if err != nil {
				store.log.WithError(err).Error("Failed to unmarshal sync IDs ranges")
			}
		}

		idRangesData := b.Get([]byte(syncIDRangesKey))
		if idRangesData != nil {
			if err := json.Unmarshal(idRangesData, &idRanges); err != nil {
				store.log.WithError(err).Error("Failed to unmarshal sync IDs ranges")
			}
		}

		idsToBeDeletedData := b.Get([]byte(syncIDsToBeDeletedKey))
		if idsToBeDeletedData != nil {
			if err := json.Unmarshal(idsToBeDeletedData, &idsToBeDeleted); err != nil {
				store.log.WithError(err).Error("Failed to unmarshal sync IDs to be deleted")
			}
		}

		return
	})
	if err != nil {
		store.log.WithError(err).Error("Failed to load sync state")
	}

	return newSyncState(store, finishTime, idRanges, idsToBeDeleted)
}

// saveSyncState saves information about sync to database.
// See `triggerSync` to learn more about possible states.
func (store *Store) saveSyncState(finishTime int64, idRanges []*syncIDRange, idsToBeDeleted []string) {
	idRangesData, err := json.Marshal(idRanges)
	if err != nil {
		store.log.WithError(err).Error("Failed to marshall sync IDs ranges")
	}

	idsToBeDeletedData, err := json.Marshal(idsToBeDeleted)
	if err != nil {
		store.log.WithError(err).Error("Failed to marshall sync IDs to be deleted")
	}

	err = store.db.Update(func(tx *bolt.Tx) (err error) {
		b := tx.Bucket(syncStateBucket)
		if finishTime != 0 {
			curTime := []byte(fmt.Sprintf("%v", finishTime))
			if err := b.Put([]byte(syncFinishTimeKey), curTime); err != nil {
				return err
			}
			if err := b.Delete([]byte(syncIDRangesKey)); err != nil {
				return err
			}
			if err := b.Delete([]byte(syncIDsToBeDeletedKey)); err != nil {
				return err
			}
		} else {
			if err := b.Delete([]byte(syncFinishTimeKey)); err != nil {
				return err
			}
			if err := b.Put([]byte(syncIDRangesKey), idRangesData); err != nil {
				return err
			}
			if err := b.Put([]byte(syncIDsToBeDeletedKey), idsToBeDeletedData); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		store.log.WithError(err).Error("Failed to set sync state")
	}
}
