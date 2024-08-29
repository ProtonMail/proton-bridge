// Copyright (c) 2024 Proton AG
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
// along with Proton Mail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package notifications

import (
	"testing"
	"time"

	"github.com/ProtonMail/go-proton-api"
	"github.com/stretchr/testify/require"
)

func TestIsBodyBitfieldValid(t *testing.T) {
	tests := []struct {
		notification proton.NotificationEvent
		isValid      bool
	}{
		{notification: proton.NotificationEvent{Payload: proton.NotificationPayload{Body: ""}}, isValid: true},
		{notification: proton.NotificationEvent{Payload: proton.NotificationPayload{Body: "HELLO"}}, isValid: true},
		{notification: proton.NotificationEvent{Payload: proton.NotificationPayload{Body: "What is up?"}}, isValid: true},
		{notification: proton.NotificationEvent{Payload: proton.NotificationPayload{Body: "123 Hello"}}, isValid: true},
		{notification: proton.NotificationEvent{Payload: proton.NotificationPayload{Body: "\\123Hello"}}, isValid: false},
		{notification: proton.NotificationEvent{Payload: proton.NotificationPayload{Body: "\\123 Hello"}}, isValid: false},
		{notification: proton.NotificationEvent{Payload: proton.NotificationPayload{Body: "\\1test"}}, isValid: false},
		{notification: proton.NotificationEvent{Payload: proton.NotificationPayload{Body: "\\1 test"}}, isValid: false},
	}

	for _, test := range tests {
		isValid := isBodyBitfieldValid(test.notification)
		require.Equal(t, isValid, !test.isValid)
	}
}

// The notification TTL is defined as timestamp by server + predefined duration.
func TestShouldSendAndStore(t *testing.T) {
	getDirFn := func(dir string) func() (string, error) {
		return func() (string, error) {
			return dir, nil
		}
	}

	dir1 := t.TempDir()
	dir2 := t.TempDir()

	store := NewStore(getDirFn(dir1))

	notification1 := proton.NotificationEvent{ID: "1", Payload: proton.NotificationPayload{Title: "test1", Subtitle: "test1", Body: "test1"}, Time: time.Now().Unix()}
	notification2 := proton.NotificationEvent{ID: "2", Payload: proton.NotificationPayload{Title: "test2", Subtitle: "test2", Body: "test2"}, Time: time.Now().Unix()}
	notification3 := proton.NotificationEvent{ID: "3", Payload: proton.NotificationPayload{Title: "test3", Subtitle: "test3", Body: "test3"}, Time: time.Now().Unix()}
	notificationAlt1 := proton.NotificationEvent{ID: "1", Payload: proton.NotificationPayload{Title: "testAlt1", Subtitle: "test1", Body: "test1"}, Time: time.Now().Unix()}
	notificationAlt2 := proton.NotificationEvent{ID: "1", Payload: proton.NotificationPayload{Title: "test2", Subtitle: "testAlt2", Body: "test2"}, Time: time.Now().Unix()}

	require.Equal(t, true, store.shouldSendAndStore(notification1))
	require.Equal(t, true, store.shouldSendAndStore(notification2))
	require.Equal(t, true, store.shouldSendAndStore(notification3))
	require.Equal(t, true, store.shouldSendAndStore(notificationAlt1))
	require.Equal(t, true, store.shouldSendAndStore(notificationAlt2))

	require.Equal(t, false, store.shouldSendAndStore(notification1))
	require.Equal(t, false, store.shouldSendAndStore(notification2))
	require.Equal(t, false, store.shouldSendAndStore(notification3))

	store = NewStore(getDirFn(dir1))

	// These should be cached in the file
	require.Equal(t, false, store.shouldSendAndStore(notification1))
	require.Equal(t, false, store.shouldSendAndStore(notification2))
	require.Equal(t, false, store.shouldSendAndStore(notification3))

	store = NewStore(getDirFn(dir2))
	timeOffset = 1 * time.Second

	// We're basing the time based on when the notification is sent
	// Let's reset it.
	notification1 = proton.NotificationEvent{ID: "1", Payload: proton.NotificationPayload{Title: "test1", Subtitle: "test1", Body: "test1"}, Time: time.Now().Unix()}
	notification2 = proton.NotificationEvent{ID: "2", Payload: proton.NotificationPayload{Title: "test2", Subtitle: "test2", Body: "test2"}, Time: time.Now().Unix()}
	notification3 = proton.NotificationEvent{ID: "3", Payload: proton.NotificationPayload{Title: "test3", Subtitle: "test3", Body: "test3"}, Time: time.Now().Unix()}

	require.Equal(t, true, store.shouldSendAndStore(notification1))
	require.Equal(t, true, store.shouldSendAndStore(notification2))
	require.Equal(t, true, store.shouldSendAndStore(notification3))

	require.Equal(t, false, store.shouldSendAndStore(notification1))
	require.Equal(t, false, store.shouldSendAndStore(notification2))
	require.Equal(t, false, store.shouldSendAndStore(notification3))

	time.Sleep(1200 * time.Millisecond)

	require.Equal(t, true, store.shouldSendAndStore(notification1))
	require.Equal(t, true, store.shouldSendAndStore(notification2))
	require.Equal(t, true, store.shouldSendAndStore(notification3))

	store = NewStore(getDirFn(dir2))
	require.Equal(t, true, store.shouldSendAndStore(notification1))
	require.Equal(t, true, store.shouldSendAndStore(notification2))
	require.Equal(t, true, store.shouldSendAndStore(notification3))
}
