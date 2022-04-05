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
	"sync"
	"time"

	"github.com/pkg/errors"
)

type syncState struct {
	lock  *sync.RWMutex
	store storeSynchronizer

	// finishTime is the time, when the sync was finished for the last time.
	// When it's zero, it was never finished or the sync is ongoing.
	finishTime int64

	// idRanges are ID ranges which are used to split work in several workers.
	// On the beginning of the sync it will find split IDs which are used to
	// create this ranges. If we have 10000 messages and five workers, it will
	// find IDs around 2000, 4000, 6000 and 8000 and then first worker will
	// sync IDs 0-2000, second 2000-4000 and so on.
	idRanges []*syncIDRange

	// idsToBeDeletedMap is map with keys as message IDs. On the beginning
	// of the sync, it will load all message IDs in database. During the sync,
	// it will delete all messages from the map which were sycned. The rest
	// at the end of the sync will be removed as those messages were not synced
	// again. We do that because we don't want to remove everything on the
	// beginning of the sync to keep client synced.
	idsToBeDeletedMap map[string]bool
}

func newSyncState(store storeSynchronizer, finishTime int64, idRanges []*syncIDRange, idsToBeDeleted []string) *syncState {
	idsToBeDeletedMap := map[string]bool{}
	for _, id := range idsToBeDeleted {
		idsToBeDeletedMap[id] = true
	}

	syncState := &syncState{
		lock:  &sync.RWMutex{},
		store: store,

		finishTime:        finishTime,
		idRanges:          idRanges,
		idsToBeDeletedMap: idsToBeDeletedMap,
	}

	for _, idRange := range idRanges {
		idRange.syncState = syncState
	}

	return syncState
}

func (s *syncState) save() {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.store.saveSyncState(s.finishTime, s.idRanges, s.getIDsToBeDeleted())
}

// isIncomplete returns whether the sync is in progress (no matter whether
// the sync is running or just not finished by info from database).
func (s *syncState) isIncomplete() bool {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.finishTime == 0 && len(s.idRanges) != 0
}

// isFinished returns whether the sync was finished.
func (s *syncState) isFinished() bool {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.finishTime != 0
}

// clearFinishTime sets finish time to zero.
func (s *syncState) clearFinishTime() {
	s.lock.Lock()
	defer s.save()
	defer s.lock.Unlock()

	s.finishTime = 0
}

// setFinishTime sets finish time to current time.
func (s *syncState) setFinishTime() {
	s.lock.Lock()
	defer s.save()
	defer s.lock.Unlock()

	s.finishTime = time.Now().UnixNano()
}

// initIDRanges inits the main full range. Then each range is added
// by `addIDRange`.
func (s *syncState) initIDRanges() {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.idRanges = []*syncIDRange{{
		syncState: s,
		StartID:   "",
		StopID:    "",
	}}
}

// addIDRange sets `splitID` as stopID for last range and adds new one
// starting with `splitID`.
func (s *syncState) addIDRange(splitID string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	lastGroup := s.idRanges[len(s.idRanges)-1]
	lastGroup.StopID = splitID

	s.idRanges = append(s.idRanges, &syncIDRange{
		syncState: s,
		StartID:   splitID,
		StopID:    "",
	})
}

// loadMessageIDsToBeDeleted loads all message IDs from database
// and by default all IDs are meant for deletion. During sync for
// each ID `doNotDeleteMessageID` has to be called to remove that
// message from being deleted by `deleteMessagesToBeDeleted`.
func (s *syncState) loadMessageIDsToBeDeleted() error {
	idsToBeDeletedMap := make(map[string]bool)
	ids, err := s.store.getAllMessageIDs()
	if err != nil {
		return err
	}
	for _, id := range ids {
		idsToBeDeletedMap[id] = true
	}

	s.lock.Lock()
	defer s.save()
	defer s.lock.Unlock()

	s.idsToBeDeletedMap = idsToBeDeletedMap
	return nil
}

func (s *syncState) doNotDeleteMessageID(id string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	delete(s.idsToBeDeletedMap, id)
}

func (s *syncState) deleteMessagesToBeDeleted() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	idsToBeDeleted := s.getIDsToBeDeleted()
	log.Infof("Deleting %v messages after sync", len(idsToBeDeleted))
	if err := s.store.deleteMessagesEvent(idsToBeDeleted); err != nil {
		return errors.Wrap(err, "failed to delete messages")
	}
	return nil
}

// getIDsToBeDeleted is helper to convert internal map for easier
// manipulation to array.
func (s *syncState) getIDsToBeDeleted() []string {
	keys := []string{}
	for key := range s.idsToBeDeletedMap {
		keys = append(keys, key)
	}
	return keys
}

// syncIDRange holds range which IDs need to be synced.
type syncIDRange struct {
	syncState *syncState
	StartID   string
	StopID    string
}

func (r *syncIDRange) setStartID(startID string) {
	r.StartID = startID
	r.syncState.save()
}

func (r *syncIDRange) setStopID(stopID string) {
	r.StopID = stopID
	r.syncState.save()
}

// isFinished returns syncIDRange is finished when StartID and StopID
// are the same. But it cannot be full range, full range cannot be
// determined in other way than asking API.
func (r *syncIDRange) isFinished() bool {
	return r.StartID == r.StopID && r.StartID != ""
}
