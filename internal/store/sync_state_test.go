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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSyncState_IDRanges(t *testing.T) {
	store := newSyncer()
	syncState := newSyncState(store, 0, []*syncIDRange{}, []string{})

	syncState.initIDRanges()
	syncState.addIDRange("100")
	syncState.addIDRange("200")

	r := syncState.idRanges
	assert.Equal(t, "", r[0].StartID)
	assert.Equal(t, "100", r[0].StopID)
	assert.Equal(t, "100", r[1].StartID)
	assert.Equal(t, "200", r[1].StopID)
	assert.Equal(t, "200", r[2].StartID)
	assert.Equal(t, "", r[2].StopID)
}

func TestSyncState_IDRangesLoaded(t *testing.T) {
	store := newSyncer()
	syncState := newSyncState(store, 0, []*syncIDRange{
		{StartID: "", StopID: "100"},
		{StartID: "100", StopID: ""},
	}, []string{})

	r := syncState.idRanges
	assert.Equal(t, "", r[0].StartID)
	assert.Equal(t, "100", r[0].StopID)
	assert.Equal(t, "100", r[1].StartID)
	assert.Equal(t, "", r[1].StopID)
}

func TestSyncState_IDsToBeDeleted(t *testing.T) {
	store := newSyncer()
	store.allMessageIDs = generateIDs(1, 9)

	syncState := newSyncState(store, 0, []*syncIDRange{}, []string{})

	require.Nil(t, syncState.loadMessageIDsToBeDeleted())
	syncState.doNotDeleteMessageID("1")
	syncState.doNotDeleteMessageID("2")
	syncState.doNotDeleteMessageID("3")
	syncState.doNotDeleteMessageID("notthere")

	idsToBeDeleted := syncState.getIDsToBeDeleted()
	sort.Strings(idsToBeDeleted)
	assert.Equal(t, generateIDs(4, 9), idsToBeDeleted)
}

func TestSyncState_IDsToBeDeletedLoaded(t *testing.T) {
	store := newSyncer()
	store.allMessageIDs = generateIDs(1, 9)

	syncState := newSyncState(store, 0, []*syncIDRange{}, generateIDs(4, 9))

	idsToBeDeleted := syncState.getIDsToBeDeleted()
	sort.Strings(idsToBeDeleted)
	assert.Equal(t, generateIDs(4, 9), idsToBeDeleted)
}
