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
	"sort"
	"strconv"
	"sync"
	"testing"

	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockLister struct {
	err        error
	messageIDs []string
}

func (m *mockLister) ListMessages(_ context.Context, filter *pmapi.MessagesFilter) (msgs []*pmapi.Message, total int, err error) {
	if m.err != nil {
		return nil, 0, m.err
	}
	skipByID := true
	skipByPaging := filter.PageSize * filter.Page
	for idx := 0; idx < len(m.messageIDs); idx++ {
		var messageID string
		if !*filter.Desc {
			messageID = m.messageIDs[idx]
			if filter.BeginID == "" || messageID == filter.BeginID {
				skipByID = false
			}
		} else {
			messageID = m.messageIDs[len(m.messageIDs)-1-idx]
			if filter.EndID == "" || messageID == filter.EndID {
				skipByID = false
			}
		}
		if skipByID {
			continue
		}
		skipByPaging--
		if skipByPaging > 0 {
			continue
		}
		msgs = append(msgs, &pmapi.Message{
			ID: messageID,
		})
		if len(msgs) == filter.PageSize || len(msgs) == filter.Limit {
			break
		}
		if !*filter.Desc {
			if messageID == filter.EndID {
				break
			}
		} else {
			if messageID == filter.BeginID {
				break
			}
		}
	}
	return msgs, len(m.messageIDs), nil
}

type mockStoreSynchronizer struct {
	locker                         sync.Locker
	allMessageIDs                  []string
	errCreateOrUpdateMessagesEvent error
	createdMessageIDsByBatch       [][]string
}

func newSyncer() *mockStoreSynchronizer {
	return &mockStoreSynchronizer{
		locker: &sync.Mutex{},
	}
}

func (m *mockStoreSynchronizer) getAllMessageIDs() ([]string, error) {
	m.locker.Lock()
	defer m.locker.Unlock()

	return m.allMessageIDs, nil
}

func (m *mockStoreSynchronizer) createOrUpdateMessagesEvent(messages []*pmapi.Message) error {
	m.locker.Lock()
	defer m.locker.Unlock()

	if m.errCreateOrUpdateMessagesEvent != nil {
		return m.errCreateOrUpdateMessagesEvent
	}
	createdMessageIDs := []string{}
	for _, message := range messages {
		createdMessageIDs = append(createdMessageIDs, message.ID)
	}
	m.createdMessageIDsByBatch = append(m.createdMessageIDsByBatch, createdMessageIDs)
	return nil
}

func (m *mockStoreSynchronizer) deleteMessagesEvent([]string) error {
	m.locker.Lock()
	defer m.locker.Unlock()

	return nil
}

func (m *mockStoreSynchronizer) saveSyncState(finishTime int64, idRanges []*syncIDRange, idsToBeDeleted []string) {
	m.locker.Lock()
	defer m.locker.Unlock()
}

func newTestSyncState(store storeSynchronizer, splitIDs ...string) *syncState {
	syncState := newSyncState(store, 0, []*syncIDRange{}, []string{})
	syncState.initIDRanges()
	for _, splitID := range splitIDs {
		syncState.addIDRange(splitID)
	}
	return syncState
}

func generateIDs(start, stop int) []string {
	ids := []string{}
	for x := start; x <= stop; x++ {
		ids = append(ids, strconv.Itoa(x))
	}
	return ids
}

func generateIDsR(start, stop int) []string {
	ids := []string{}
	for x := start; x >= stop; x-- {
		ids = append(ids, strconv.Itoa(x))
	}
	return ids
}

// Tests

func TestSyncAllMail(t *testing.T) { //nolint:funlen
	m, clear := initMocks(t)
	defer clear()

	numberOfMessages := 10000

	api := &mockLister{
		messageIDs: generateIDs(1, numberOfMessages),
	}

	tests := []struct {
		name              string
		idRanges          []*syncIDRange
		idsToBeDeleted    []string
		wantUpdatedIDs    []string
		wantNotUpdatedIDs []string
	}{
		{
			"full sync",
			[]*syncIDRange{},
			[]string{},
			api.messageIDs,
			[]string{},
		},
		{
			"continue with interrupted sync",
			[]*syncIDRange{
				{StartID: "2000", StopID: "2100"},
				{StartID: "4000", StopID: "4200"},
				{StartID: "9500", StopID: ""},
			},
			mergeArrays(generateIDs(2000, 2100), generateIDs(4000, 4200), generateIDs(9500, 10010)),
			mergeArrays(generateIDs(2000, 2100), generateIDs(4000, 4200), generateIDs(9500, 10000)),
			mergeArrays(generateIDs(1, 1999), generateIDs(2101, 3999), generateIDs(4201, 9459)),
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			store := newSyncer()
			store.allMessageIDs = generateIDs(1, numberOfMessages+10)

			syncState := newSyncState(store, 0, tc.idRanges, tc.idsToBeDeleted)

			err := syncAllMail(m.panicHandler, store, api, syncState)
			require.Nil(t, err)

			// Check all messages were created or updated.
			updateMessageIDsMap := map[string]bool{}
			for _, messageIDs := range store.createdMessageIDsByBatch {
				for _, messageID := range messageIDs {
					updateMessageIDsMap[messageID] = true
				}
			}
			for _, messageID := range tc.wantUpdatedIDs {
				assert.True(t, updateMessageIDsMap[messageID], "Message %s was not created/updated, but should", messageID)
			}
			for _, messageID := range tc.wantNotUpdatedIDs {
				assert.False(t, updateMessageIDsMap[messageID], "Message %s was created/updated, but shouldn't", messageID)
			}

			// Check all messages were deleted.
			idsToBeDeleted := syncState.getIDsToBeDeleted()
			sort.Strings(idsToBeDeleted)
			assert.Equal(t, generateIDs(numberOfMessages+1, numberOfMessages+10), idsToBeDeleted)
		})
	}
}

func mergeArrays(arrays ...[]string) []string {
	result := []string{}
	for _, array := range arrays {
		result = append(result, array...)
	}
	return result
}

func TestSyncAllMail_FailedListing(t *testing.T) {
	m, clear := initMocks(t)
	defer clear()

	numberOfMessages := 10000

	store := newSyncer()
	store.allMessageIDs = generateIDs(1, numberOfMessages+10)

	api := &mockLister{
		err:        errors.New("error"),
		messageIDs: generateIDs(1, numberOfMessages),
	}
	syncState := newTestSyncState(store)

	err := syncAllMail(m.panicHandler, store, api, syncState)
	require.EqualError(t, err, "failed to sync group: failed to list messages: error")
}

func TestSyncAllMail_FailedCreateOrUpdateMessage(t *testing.T) {
	m, clear := initMocks(t)
	defer clear()

	numberOfMessages := 10000

	store := newSyncer()
	store.errCreateOrUpdateMessagesEvent = errors.New("error")
	store.allMessageIDs = generateIDs(1, numberOfMessages+10)

	api := &mockLister{
		messageIDs: generateIDs(1, numberOfMessages),
	}
	syncState := newTestSyncState(store)

	err := syncAllMail(m.panicHandler, store, api, syncState)
	require.EqualError(t, err, "failed to sync group: failed to create or update messages: error")
}

func TestFindIDRanges(t *testing.T) { //nolint:funlen
	store := newSyncer()
	syncState := newTestSyncState(store)

	tests := []struct {
		name        string
		messageIDs  []string
		wantBatches [][]string
	}{
		{
			"1k messages - 1 batch",
			generateIDs(1, 1000),
			[][]string{
				{"", ""},
			},
		},
		{
			"1k messages not starting at 1",
			generateIDs(1000, 2000),
			[][]string{
				{"", ""},
			},
		},
		{
			"2k messages - 2 batches",
			generateIDs(1, 2000),
			[][]string{
				{"", "1050"},
				{"1050", ""},
			},
		},
		{
			"4k messages - 3 batches",
			generateIDs(1, 4000),
			[][]string{
				{"", "1350"},
				{"1350", "2700"},
				{"2700", ""},
			},
		},
		{
			"5k messages - 4 batches",
			generateIDs(1, 5000),
			[][]string{
				{"", "1350"},
				{"1350", "2700"},
				{"2700", "4050"},
				{"4050", ""},
			},
		},
		{
			"10k messages - 5 batches",
			generateIDs(1, 10000),
			[][]string{
				{"", "2100"},
				{"2100", "4200"},
				{"4200", "6300"},
				{"6300", "8400"},
				{"8400", ""},
			},
		},
		{
			"150k messages - 5 batches",
			generateIDs(1, 150000),
			[][]string{
				{"", "30000"},
				{"30000", "60000"},
				{"60000", "90000"},
				{"90000", "120000"},
				{"120000", ""},
			},
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			api := &mockLister{
				messageIDs: tc.messageIDs,
			}

			err := findIDRanges(pmapi.AllMailLabel, api, syncState)

			require.Nil(t, err)
			require.Equal(t, len(tc.wantBatches), len(syncState.idRanges))
			for idx, idRange := range syncState.idRanges {
				want := tc.wantBatches[idx]
				assert.Equal(t, want[0], idRange.StartID, "Start ID for IDs range %d does not match", idx)
				assert.Equal(t, want[1], idRange.StopID, "Stop ID for IDs range %d does not match", idx)
			}
		})
	}
}

func TestFindIDRanges_FailedListing(t *testing.T) {
	store := newSyncer()
	api := &mockLister{
		err: errors.New("error"),
	}

	syncState := newTestSyncState(store)

	err := findIDRanges(pmapi.AllMailLabel, api, syncState)
	require.EqualError(t, err, "failed to get first ID and count: failed to list messages: error")
}

func TestGetSplitIDAndCount(t *testing.T) { //nolint:funlen
	tests := []struct {
		name       string
		err        error
		messageIDs []string
		page       int
		wantID     string
		wantTotal  int
		wantErr    string
	}{
		{
			"1000 messages, first page",
			nil,
			generateIDs(1, 1000),
			0,
			"1",
			1000,
			"",
		},
		{
			"1000 messages, 5th page",
			nil,
			generateIDs(1, 1000),
			4,
			"600",
			1000,
			"",
		},
		{
			"one message, first page",
			nil,
			[]string{"1"},
			0,
			"1",
			1,
			"",
		},
		{
			"no message, first page",
			nil,
			[]string{},
			0,
			"",
			0,
			"",
		},
		{
			"listing error",
			errors.New("error"),
			generateIDs(1, 1000),
			0,
			"",
			0,
			"failed to list messages: error",
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			api := &mockLister{
				err:        tc.err,
				messageIDs: tc.messageIDs,
			}

			id, total, err := getSplitIDAndCount(pmapi.AllMailLabel, api, tc.page)

			if tc.wantErr == "" {
				require.Nil(t, err)
			} else {
				require.EqualError(t, err, tc.wantErr)
			}
			assert.Equal(t, tc.wantID, id)
			assert.Equal(t, tc.wantTotal, total)
		})
	}
}

func TestSyncBatch(t *testing.T) {
	tests := []struct {
		name                         string
		batchIdx                     int
		wantCreatedMessageIDsByBatch [][]string
	}{
		{
			"first-batch",
			0,
			[][]string{generateIDsR(200, 51), generateIDsR(51, 1)},
		},
		{
			"second-batch",
			1,
			[][]string{generateIDsR(400, 251), generateIDsR(251, 200)},
		},
		{
			"third-batch",
			2,
			[][]string{generateIDsR(600, 451), generateIDsR(451, 400)},
		},
		{
			"fourth-batch",
			3,
			[][]string{generateIDsR(800, 651), generateIDsR(651, 600)},
		},
		{
			"fifth-batch",
			4,
			[][]string{generateIDsR(1000, 851), generateIDsR(851, 800)},
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			store := newSyncer()
			api := &mockLister{
				messageIDs: generateIDs(1, 1000),
			}

			err := testSyncBatch(t, store, api, tc.batchIdx, "200", "400", "600", "800")
			require.Nil(t, err)
			require.Equal(t, tc.wantCreatedMessageIDsByBatch, store.createdMessageIDsByBatch)
		})
	}
}

func TestSyncBatch_FailedListing(t *testing.T) {
	store := newSyncer()
	api := &mockLister{
		err:        errors.New("error"),
		messageIDs: generateIDs(1, 1000),
	}

	err := testSyncBatch(t, store, api, 0)
	require.EqualError(t, err, "failed to list messages: error")
}

func TestSyncBatch_FailedCreateOrUpdateMessage(t *testing.T) {
	store := newSyncer()
	store.errCreateOrUpdateMessagesEvent = errors.New("error")
	api := &mockLister{
		messageIDs: generateIDs(1, 1000),
	}

	err := testSyncBatch(t, store, api, 0)
	require.EqualError(t, err, "failed to create or update messages: error")
}

func testSyncBatch(t *testing.T, store storeSynchronizer, api messageLister, rangeIdx int, splitIDs ...string) error { //nolint:unparam
	syncState := newTestSyncState(store, splitIDs...)
	idRange := syncState.idRanges[rangeIdx]
	shouldStop := 0
	return syncBatch(pmapi.AllMailLabel, store, api, syncState, idRange, &shouldStop)
}
