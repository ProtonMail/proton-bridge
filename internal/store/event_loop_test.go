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
	"time"

	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	mail "github.com/ProtonMail/proton-bridge/pkg/net/mail"
)

func TestEventLoopProcessMoreEvents(t *testing.T) {
	m, clear := initMocks(t)
	defer clear()

	// Event expectations need to be defined before calling `newStoreNoEvents`
	// to force to use these for this particular test.
	// Also, event loop calls ListMessages again and we need to place it after
	// calling `newStoreNoEvents` to not break expectations for the first sync.
	gomock.InOrder(
		// Doesn't matter which IDs are used.
		// This test is trying to see whether event loop will immediately process
		// next event if there is `More` of them.
		m.client.EXPECT().GetEvent("latestEventID").Return(&pmapi.Event{
			EventID: "event50",
			More:    1,
		}, nil),
		m.client.EXPECT().GetEvent("event50").Return(&pmapi.Event{
			EventID: "event70",
			More:    0,
		}, nil),
		m.client.EXPECT().GetEvent("event70").Return(&pmapi.Event{
			EventID: "event71",
			More:    0,
		}, nil),
	)
	m.newStoreNoEvents(true)
	m.client.EXPECT().ListMessages(gomock.Any()).Return([]*pmapi.Message{}, 0, nil).AnyTimes()

	// Event loop runs in goroutine and will be stopped by deferred mock clearing.
	go m.store.eventLoop.start()

	// More events are processed right away.
	require.Eventually(t, func() bool {
		return m.store.eventLoop.currentEventID == "event70"
	}, time.Second, 10*time.Millisecond)

	// For normal event we need to wait to next polling.
	time.Sleep(pollInterval + pollIntervalSpread)
	require.Eventually(t, func() bool {
		return m.store.eventLoop.currentEventID == "event71"
	}, time.Second, 10*time.Millisecond)
}

func TestEventLoopUpdateMessageFromLoop(t *testing.T) {
	m, clear := initMocks(t)
	defer clear()

	subject := "old subject"
	newSubject := "new subject"

	// First sync will add message with old subject to database.
	m.client.EXPECT().GetMessage("msg1").Return(&pmapi.Message{
		ID:      "msg1",
		Subject: subject,
	}, nil)
	// Event will update the subject.
	m.client.EXPECT().GetEvent("latestEventID").Return(&pmapi.Event{
		EventID: "event1",
		Messages: []*pmapi.EventMessage{{
			EventItem: pmapi.EventItem{
				ID:     "msg1",
				Action: pmapi.EventUpdate,
			},
			Updated: &pmapi.EventMessageUpdated{
				ID:      "msg1",
				Subject: &newSubject,
			},
		}},
	}, nil)

	m.newStoreNoEvents(true)

	// Event loop runs in goroutine and will be stopped by deferred mock clearing.
	go m.store.eventLoop.start()

	var err error
	assert.Eventually(t, func() bool {
		var msg *pmapi.Message
		msg, err = m.store.getMessageFromDB("msg1")
		return err == nil && msg.Subject == newSubject
	}, time.Second, 10*time.Millisecond)
	require.NoError(t, err)
}

func TestEventLoopUpdateMessage(t *testing.T) {
	address1 := &mail.Address{Address: "user1@example.com"}
	address2 := &mail.Address{Address: "user2@example.com"}
	msg := &pmapi.Message{
		ID:       "msg1",
		Subject:  "old",
		Unread:   0,
		Flags:    10,
		Sender:   address1,
		ToList:   []*mail.Address{address2},
		CCList:   []*mail.Address{address1},
		BCCList:  []*mail.Address{},
		Time:     20,
		LabelIDs: []string{"old"},
	}
	newMsg := &pmapi.Message{
		ID:       "msg1",
		Subject:  "new",
		Unread:   1,
		Flags:    11,
		Sender:   address2,
		ToList:   []*mail.Address{address1},
		CCList:   []*mail.Address{address2},
		BCCList:  []*mail.Address{address1},
		Time:     21,
		LabelIDs: []string{"new"},
	}

	updateMessage(log, msg, &pmapi.EventMessageUpdated{
		ID:       "msg1",
		Subject:  &newMsg.Subject,
		Unread:   &newMsg.Unread,
		Flags:    &newMsg.Flags,
		Sender:   newMsg.Sender,
		ToList:   &newMsg.ToList,
		CCList:   &newMsg.CCList,
		BCCList:  &newMsg.BCCList,
		Time:     newMsg.Time,
		LabelIDs: newMsg.LabelIDs,
	})

	require.Equal(t, newMsg, msg)
}
