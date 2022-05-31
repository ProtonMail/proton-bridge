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
	"net/mail"
	"testing"
	"time"

	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
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
		m.client.EXPECT().GetEvent(gomock.Any(), "latestEventID").Return(&pmapi.Event{
			EventID: "event50",
			More:    true,
		}, nil),
		m.client.EXPECT().GetEvent(gomock.Any(), "event50").Return(&pmapi.Event{
			EventID: "event70",
			More:    false,
		}, nil),
		m.client.EXPECT().GetEvent(gomock.Any(), "event70").Return(&pmapi.Event{
			EventID: "event71",
			More:    false,
		}, nil),
	)
	m.newStoreNoEvents(t, true)

	// Event loop runs in goroutine started during store creation (newStoreNoEvents).
	// Force to run the next event.
	m.store.eventLoop.pollNow()

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

	m.newStoreNoEvents(t, true, &pmapi.Message{
		ID:      "msg1",
		Subject: subject,
	})

	testEvent(t, m, &pmapi.Event{
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
	})

	msg, err := m.store.getMessageFromDB("msg1")
	require.NoError(t, err)
	require.Equal(t, newSubject, msg.Subject)
}

func TestEventLoopDeletionNotPaused(t *testing.T) {
	m, clear := initMocks(t)
	defer clear()

	m.newStoreNoEvents(t, true, &pmapi.Message{
		ID:       "msg1",
		Subject:  "subject",
		LabelIDs: []string{"label"},
	})

	m.changeNotifier.EXPECT().CanDelete("label").Return(true, func() {})
	m.store.SetChangeNotifier(m.changeNotifier)

	testEvent(t, m, &pmapi.Event{
		EventID: "event1",
		Messages: []*pmapi.EventMessage{{
			EventItem: pmapi.EventItem{
				ID:     "msg1",
				Action: pmapi.EventDelete,
			},
		}},
	})

	_, err := m.store.getMessageFromDB("msg1")
	require.Error(t, err)
}

func TestEventLoopDeletionPaused(t *testing.T) {
	m, clear := initMocks(t)
	defer clear()

	m.newStoreNoEvents(t, true, &pmapi.Message{
		ID:       "msg1",
		Subject:  "subject",
		LabelIDs: []string{"label"},
	})

	delay := 5 * time.Second

	m.changeNotifier.EXPECT().CanDelete("label").Return(false, func() {
		time.Sleep(delay)
	})
	m.changeNotifier.EXPECT().CanDelete("label").Return(true, func() {})
	m.store.SetChangeNotifier(m.changeNotifier)

	start := time.Now()

	testEvent(t, m, &pmapi.Event{
		EventID: "event1",
		Messages: []*pmapi.EventMessage{{
			EventItem: pmapi.EventItem{
				ID:     "msg1",
				Action: pmapi.EventDelete,
			},
		}},
	})

	_, err := m.store.getMessageFromDB("msg1")
	require.Error(t, err)
	require.True(t, time.Since(start) > delay)
}

func testEvent(t *testing.T, m *mocksForStore, event *pmapi.Event) {
	eventReceived := make(chan struct{})
	m.client.EXPECT().GetEvent(gomock.Any(), "latestEventID").DoAndReturn(func(_ context.Context, eventID string) (*pmapi.Event, error) {
		defer close(eventReceived)
		return event, nil
	})

	// Event loop runs in goroutine started during store creation (newStoreNoEvents).
	// Force to run the next event.
	m.store.eventLoop.pollNow()

	select {
	case <-eventReceived:
	case <-time.After(5 * time.Second):
		require.Fail(t, "latestEventID was not processed")
	}
}

func TestEventLoopUpdateMessage(t *testing.T) {
	address1 := &mail.Address{Address: "user1@example.com"}
	address2 := &mail.Address{Address: "user2@example.com"}
	msg := &pmapi.Message{
		ID:       "msg1",
		Subject:  "old",
		Unread:   false,
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
		Unread:   true,
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
