// Copyright (c) 2020 Proton Technologies AG
//
// This file is part of ProtonMail Bridge.
//
// ProtonMail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ProtonMail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ProtonMail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package store

import (
	"testing"

	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
	imapBackend "github.com/emersion/go-imap/backend"
	"github.com/stretchr/testify/require"
)

func TestCreateOrUpdateMessageIMAPUpdates(t *testing.T) {
	m, clear := initMocks(t)
	defer clear()

	updates := make(chan interface{})

	m.newStoreNoEvents(true)
	m.store.SetIMAPUpdateChannel(updates)

	go checkIMAPUpdates(t, updates, []func(interface{}) bool{
		checkMessageUpdate(addr1, "All Mail", 1, 1),
		checkMessageUpdate(addr1, "All Mail", 2, 2),
	})

	insertMessage(t, m, "msg1", "Test message 1", addrID1, 0, []string{pmapi.AllMailLabel})
	insertMessage(t, m, "msg2", "Test message 2", addrID1, 0, []string{pmapi.AllMailLabel})

	close(updates)
}

func TestCreateOrUpdateMessageIMAPUpdatesBulkUpdate(t *testing.T) {
	m, clear := initMocks(t)
	defer clear()

	updates := make(chan interface{})

	m.newStoreNoEvents(true)
	m.store.SetIMAPUpdateChannel(updates)

	go checkIMAPUpdates(t, updates, []func(interface{}) bool{
		checkMessageUpdate(addr1, "All Mail", 1, 1),
		checkMessageUpdate(addr1, "All Mail", 2, 2),
	})

	msg1 := getTestMessage("msg1", "Test message 1", addrID1, 0, []string{pmapi.AllMailLabel})
	msg2 := getTestMessage("msg2", "Test message 2", addrID1, 0, []string{pmapi.AllMailLabel})
	require.Nil(t, m.store.createOrUpdateMessagesEvent([]*pmapi.Message{msg1, msg2}))

	close(updates)
}

func TestDeleteMessageIMAPUpdate(t *testing.T) {
	m, clear := initMocks(t)
	defer clear()

	m.newStoreNoEvents(true)

	insertMessage(t, m, "msg1", "Test message 1", addrID1, 0, []string{pmapi.AllMailLabel})
	insertMessage(t, m, "msg2", "Test message 2", addrID1, 0, []string{pmapi.AllMailLabel})

	updates := make(chan interface{})
	m.store.SetIMAPUpdateChannel(updates)
	go checkIMAPUpdates(t, updates, []func(interface{}) bool{
		checkMessageDelete(addr1, "All Mail", 2),
		checkMessageDelete(addr1, "All Mail", 1),
	})

	require.Nil(t, m.store.deleteMessageEvent("msg2"))
	require.Nil(t, m.store.deleteMessageEvent("msg1"))
	close(updates)
}

func checkIMAPUpdates(t *testing.T, updates chan interface{}, checkFunctions []func(interface{}) bool) {
	idx := 0
	for update := range updates {
		if idx >= len(checkFunctions) {
			continue
		}
		if !checkFunctions[idx](update) {
			continue
		}
		idx++
	}
	require.True(t, idx == len(checkFunctions), "Less updates than expected: %+v of %+v", idx, len(checkFunctions))
}

func checkMessageUpdate(username, mailbox string, seqNum, uid int) func(interface{}) bool { //nolint[unparam]
	return func(update interface{}) bool {
		switch u := update.(type) {
		case *imapBackend.MessageUpdate:
			return (u.Update.Username() == username &&
				u.Update.Mailbox() == mailbox &&
				u.Message.SeqNum == uint32(seqNum) &&
				u.Message.Uid == uint32(uid))
		default:
			return false
		}
	}
}

func checkMessageDelete(username, mailbox string, seqNum int) func(interface{}) bool { //nolint[unparam]
	return func(update interface{}) bool {
		switch u := update.(type) {
		case *imapBackend.ExpungeUpdate:
			return (u.Update.Username() == username &&
				u.Update.Mailbox() == mailbox &&
				u.SeqNum == uint32(seqNum))
		default:
			return false
		}
	}
}
