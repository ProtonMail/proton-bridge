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
	"sort"
	"testing"

	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadSaveSyncState(t *testing.T) {
	m, clear := initMocks(t)
	defer clear()

	m.newStoreNoEvents(t, true)
	insertMessage(t, m, "msg1", "Test message 1", addrID1, false, []string{pmapi.AllMailLabel, pmapi.InboxLabel})
	insertMessage(t, m, "msg2", "Test message 2", addrID1, false, []string{pmapi.AllMailLabel, pmapi.InboxLabel})

	// Clear everything.

	syncState := m.store.loadSyncState()
	syncState.clearFinishTime()

	// Check everything is empty at the beginning.

	syncState = m.store.loadSyncState()
	checkSyncStateAfterLoad(t, syncState, false, false, []string{})

	// Save IDs ranges and check everything is also properly loaded.

	syncState.initIDRanges()
	syncState.addIDRange("100")
	syncState.addIDRange("200")
	syncState.save()

	syncState = m.store.loadSyncState()
	checkSyncStateAfterLoad(t, syncState, false, true, []string{})

	// Save IDs to be deleted and check everything is properly loaded.

	require.Nil(t, syncState.loadMessageIDsToBeDeleted())

	syncState = m.store.loadSyncState()
	checkSyncStateAfterLoad(t, syncState, false, true, []string{"msg1", "msg2"})

	// Set finish time and check everything is resetted to empty values.

	syncState.setFinishTime()

	syncState = m.store.loadSyncState()
	checkSyncStateAfterLoad(t, syncState, true, false, []string{})
}

func checkSyncStateAfterLoad(t *testing.T, syncState *syncState, wantIsFinished bool, wantIDRanges bool, wantIDsToBeDeleted []string) {
	assert.Equal(t, wantIsFinished, syncState.isFinished())

	if wantIDRanges {
		require.Equal(t, 3, len(syncState.idRanges))
		assert.Equal(t, "", syncState.idRanges[0].StartID)
		assert.Equal(t, "100", syncState.idRanges[0].StopID)
		assert.Equal(t, "100", syncState.idRanges[1].StartID)
		assert.Equal(t, "200", syncState.idRanges[1].StopID)
		assert.Equal(t, "200", syncState.idRanges[2].StartID)
		assert.Equal(t, "", syncState.idRanges[2].StopID)
	} else {
		assert.Empty(t, syncState.idRanges)
	}

	idsToBeDeleted := syncState.getIDsToBeDeleted()
	sort.Strings(idsToBeDeleted)
	assert.Equal(t, wantIDsToBeDeleted, idsToBeDeleted)
}
