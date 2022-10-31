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
	"testing"

	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
	a "github.com/stretchr/testify/assert"
)

func newLabel(order int, id, name string) *pmapi.Label {
	return &pmapi.Label{
		ID:    id,
		Name:  name,
		Order: order,
	}
}

func TestSortByOrder(t *testing.T) {
	want := []*pmapi.Label{
		newLabel(-1000, pmapi.InboxLabel, "INBOX"),
		newLabel(-5, pmapi.SentLabel, "Sent"),
		newLabel(-4, pmapi.ArchiveLabel, "Archive"),
		newLabel(-3, pmapi.SpamLabel, "Spam"),
		newLabel(-2, pmapi.TrashLabel, "Trash"),
		newLabel(-1, pmapi.AllMailLabel, "All Mail"),
		newLabel(100, "labelID1", "custom_label"),
		newLabel(1000, "folderID1", "custom_folder"),
	}
	labels := []*pmapi.Label{
		want[6],
		want[4],
		want[3],
		want[7],
		want[5],
		want[0],
		want[2],
		want[1],
	}

	sortByOrder(labels)
	a.Equal(t, want, labels)
}

func TestMailboxNames(t *testing.T) {
	want := map[string]string{
		pmapi.InboxLabel:   "INBOX",
		pmapi.SentLabel:    "Sent",
		pmapi.ArchiveLabel: "Archive",
		pmapi.SpamLabel:    "Spam",
		pmapi.TrashLabel:   "Trash",
		pmapi.AllMailLabel: "All Mail",
		pmapi.DraftLabel:   "Drafts",
		"labelID1":         "Labels/Label1",
		"folderID1":        "Folders/Folder1",
	}

	foldersAndLabels := []*pmapi.Label{
		newLabel(100, "labelID1", "Label1"),
		newLabel(1000, "folderID1", "Folder1"),
	}
	foldersAndLabels[1].Exclusive = true

	for _, counts := range getSystemFolders() {
		foldersAndLabels = append(foldersAndLabels, counts.getPMLabel())
	}

	got := map[string]string{}
	for _, m := range foldersAndLabels {
		got[m.ID] = getLabelPrefix(m) + m.Name
	}
	a.Equal(t, want, got)
}

func TestAddSystemLabels(t *testing.T) {}

func checkCounts(t testing.TB, wantCounts []*pmapi.MessagesCount, haveStore *Store) {
	nSystemFolders := 7
	haveCounts, err := haveStore.getOnAPICounts()
	a.NoError(t, err)
	a.Len(t, haveCounts, len(wantCounts)+nSystemFolders)
	for iWant, wantCount := range wantCounts {
		iHave := iWant + nSystemFolders
		haveCount := haveCounts[iHave]
		a.Equal(t, wantCount.LabelID, haveCount.LabelID, "iHave:%d\niWant:%d\nHave:%v\nWant:%v", iHave, iWant, haveCount, wantCount)
		a.Equal(t, wantCount.Total, int(haveCount.TotalOnAPI), "iHave:%d\niWant:%d\nHave:%v\nWant:%v", iHave, iWant, haveCount, wantCount)
		a.Equal(t, wantCount.Unread, int(haveCount.UnreadOnAPI), "iHave:%d\niWant:%d\nHave:%v\nWant:%v", iHave, iWant, haveCount, wantCount)
	}
}

func TestMailboxCountRemove(t *testing.T) {
	m, clear := initMocks(t)
	defer clear()
	m.newStoreNoEvents(t, true)

	testCounts := []*pmapi.MessagesCount{
		{LabelID: "label1", Total: 100, Unread: 0},
		{LabelID: "label2", Total: 100, Unread: 30},
		{LabelID: "label4", Total: 100, Unread: 100},
	}
	a.NoError(t, m.store.createOrUpdateOnAPICounts(testCounts))

	a.NoError(t, m.store.removeMailboxCount("not existing"))
	checkCounts(t, testCounts, m.store)

	var pop *pmapi.MessagesCount
	pop, testCounts = testCounts[2], testCounts[0:2]
	a.NoError(t, m.store.removeMailboxCount(pop.LabelID))
	checkCounts(t, testCounts, m.store)
}
