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
	"net/mail"
	"testing"

	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
	a "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type wantID struct {
	appID string
	uid   int
}

func TestGetSequenceNumberAndGetUID(t *testing.T) {
	m, clear := initMocks(t)
	defer clear()

	m.newStoreNoEvents(t, true)

	insertMessage(t, m, "msg1", "Test message 1", addrID1, false, []string{pmapi.AllMailLabel, pmapi.InboxLabel})
	insertMessage(t, m, "msg2", "Test message 2", addrID1, false, []string{pmapi.AllMailLabel, pmapi.ArchiveLabel})
	insertMessage(t, m, "msg3", "Test message 3", addrID1, false, []string{pmapi.AllMailLabel, pmapi.InboxLabel})
	insertMessage(t, m, "msg4", "Test message 4", addrID1, false, []string{pmapi.AllMailLabel})

	checkAllMessageIDs(t, m, []string{"msg1", "msg2", "msg3", "msg4"})

	checkMailboxMessageIDs(t, m, pmapi.InboxLabel, []wantID{{"msg1", 1}, {"msg3", 2}})
	checkMailboxMessageIDs(t, m, pmapi.ArchiveLabel, []wantID{{"msg2", 1}})
	checkMailboxMessageIDs(t, m, pmapi.SpamLabel, []wantID(nil))
	checkMailboxMessageIDs(t, m, pmapi.AllMailLabel, []wantID{{"msg1", 1}, {"msg2", 2}, {"msg3", 3}, {"msg4", 4}})
}

// checkMailboxMessageIDs checks that the mailbox contains all API IDs with correct sequence numbers and UIDs.
// wantIDs is map from IMAP UID to API ID. Sequence number is detected automatically by order of the ID in the map.
func checkMailboxMessageIDs(t *testing.T, m *mocksForStore, mailboxLabel string, wantIDs []wantID) {
	storeAddress := m.store.addresses[addrID1]
	storeMailbox := storeAddress.mailboxes[mailboxLabel]

	ids, err := storeMailbox.GetAPIIDsFromSequenceRange(1, uint32(len(wantIDs)))
	require.Nil(t, err)

	idx := 0
	for _, wantID := range wantIDs {
		id := ids[idx]
		require.Equal(t, wantID.appID, id, "Got IDs: %+v", ids)

		uid, err := storeMailbox.getUID(wantID.appID)
		require.Nil(t, err)
		a.Equal(t, uint32(wantID.uid), uid)

		seqNum, err := storeMailbox.getSequenceNumber(wantID.appID)
		require.Nil(t, err)
		a.Equal(t, uint32(idx+1), seqNum)

		idx++
	}
}

func TestGetUIDByHeader(t *testing.T) { //nolint:funlen
	m, clear := initMocks(t)
	defer clear()

	m.newStoreNoEvents(t, true)

	tstMsg := getTestMessage("msg1", "Without external ID", addrID1, false, []string{pmapi.AllMailLabel, pmapi.SentLabel})
	require.Nil(t, m.store.createOrUpdateMessageEvent(tstMsg))

	tstMsg = getTestMessage("msg2", "External ID with spaces", addrID1, false, []string{pmapi.AllMailLabel, pmapi.SentLabel})
	tstMsg.ExternalID = "   externalID-non-pm-com "
	require.Nil(t, m.store.createOrUpdateMessageEvent(tstMsg))

	tstMsg = getTestMessage("msg3", "External ID with <>", addrID1, false, []string{pmapi.AllMailLabel, pmapi.SentLabel})
	tstMsg.ExternalID = "<externalID@pm.me>"
	tstMsg.Header = mail.Header{"References": []string{"wrongID", "externalID-non-pm-com", "msg2"}}
	require.Nil(t, m.store.createOrUpdateMessageEvent(tstMsg))

	// Not sure if this is a real-world scenario but we should be able to address this properly.
	tstMsg = getTestMessage("msg4", "External ID with <> and spaces and special characters", addrID1, false, []string{pmapi.AllMailLabel, pmapi.SentLabel})
	tstMsg.ExternalID = "   <   external.()+*[]ID@another.pm.me    >    "
	require.Nil(t, m.store.createOrUpdateMessageEvent(tstMsg))

	testDataUIDByHeader := []struct {
		header *mail.Header
		wantID uint32
	}{
		{
			&mail.Header{"Message-Id": []string{"wrongID"}},
			0,
		},
		{
			&mail.Header{"Message-Id": []string{"ext"}},
			0,
		},
		{
			&mail.Header{"Message-Id": []string{"externalID"}},
			0,
		},
		{
			&mail.Header{"Message-Id": []string{"msg1"}},
			1,
		},
		{
			&mail.Header{"Message-Id": []string{"<msg3@pm.me>"}},
			3,
		},
		{
			&mail.Header{"Message-Id": []string{"<externalID-non-pm-com>"}},
			2,
		},
		{
			&mail.Header{"Message-Id": []string{"externalID@pm.me"}},
			3,
		},
		{
			&mail.Header{"Message-Id": []string{"external.()+*[]ID@another.pm.me"}},
			4,
		},
	}

	storeAddress := m.store.addresses[addrID1]
	storeMailbox := storeAddress.mailboxes[pmapi.SentLabel]

	for _, td := range testDataUIDByHeader {
		haveID := storeMailbox.GetUIDByHeader(td.header)
		a.Equal(t, td.wantID, haveID, "testing header: %v", td.header)
	}
}
