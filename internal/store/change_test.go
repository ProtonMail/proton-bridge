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
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestNotifyChangeCreateOrUpdateMessage(t *testing.T) {
	m, clear := initMocks(t)
	defer clear()

	m.changeNotifier.EXPECT().MailboxStatus(addr1, "All Mail", uint32(1), uint32(0), uint32(0))
	m.changeNotifier.EXPECT().MailboxStatus(addr1, "All Mail", uint32(2), uint32(0), uint32(0))
	m.changeNotifier.EXPECT().UpdateMessage(addr1, "All Mail", uint32(1), uint32(1), gomock.Any(), false)
	m.changeNotifier.EXPECT().UpdateMessage(addr1, "All Mail", uint32(2), uint32(2), gomock.Any(), false)

	m.newStoreNoEvents(t, true)
	m.store.SetChangeNotifier(m.changeNotifier)

	insertMessage(t, m, "msg1", "Test message 1", addrID1, false, []string{pmapi.AllMailLabel})
	insertMessage(t, m, "msg2", "Test message 2", addrID1, false, []string{pmapi.AllMailLabel})
}

func TestNotifyChangeCreateOrUpdateMessages(t *testing.T) {
	m, clear := initMocks(t)
	defer clear()

	m.changeNotifier.EXPECT().MailboxStatus(addr1, "All Mail", uint32(2), uint32(0), uint32(0))
	m.changeNotifier.EXPECT().UpdateMessage(addr1, "All Mail", uint32(1), uint32(1), gomock.Any(), false)
	m.changeNotifier.EXPECT().UpdateMessage(addr1, "All Mail", uint32(2), uint32(2), gomock.Any(), false)

	m.newStoreNoEvents(t, true)
	m.store.SetChangeNotifier(m.changeNotifier)

	msg1 := getTestMessage("msg1", "Test message 1", addrID1, false, []string{pmapi.AllMailLabel})
	msg2 := getTestMessage("msg2", "Test message 2", addrID1, false, []string{pmapi.AllMailLabel})
	require.Nil(t, m.store.createOrUpdateMessagesEvent([]*pmapi.Message{msg1, msg2}))
}

func TestNotifyChangeDeleteMessage(t *testing.T) {
	m, clear := initMocks(t)
	defer clear()

	m.newStoreNoEvents(t, true)

	insertMessage(t, m, "msg1", "Test message 1", addrID1, false, []string{pmapi.AllMailLabel})
	insertMessage(t, m, "msg2", "Test message 2", addrID1, false, []string{pmapi.AllMailLabel})

	m.changeNotifier.EXPECT().DeleteMessage(addr1, "All Mail", uint32(2))
	m.changeNotifier.EXPECT().DeleteMessage(addr1, "All Mail", uint32(1))

	m.store.SetChangeNotifier(m.changeNotifier)
	require.Nil(t, m.store.deleteMessageEvent("msg2"))
	require.Nil(t, m.store.deleteMessageEvent("msg1"))
}
