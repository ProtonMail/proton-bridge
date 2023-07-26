// Copyright (c) 2023 Proton AG
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
package userevents

import (
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/ProtonMail/gluon/async"
	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/proton-bridge/v3/internal/events"
	mocks2 "github.com/ProtonMail/proton-bridge/v3/internal/events/mocks"
	"github.com/ProtonMail/proton-bridge/v3/internal/services/userevents/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestService_EventIDLoadStore(t *testing.T) {
	// Simulate the flow of data when we start without any event id in the event id store:
	// * Load event id from store
	// * Get latest event id since
	// * Store that latest event id
	// * Start event poll loop
	// * Get new event id, store it in vault
	// * Try to poll new event it, but context is cancelled
	group := async.NewGroup(context.Background(), &async.NoopPanicHandler{})
	mockCtrl := gomock.NewController(t)
	eventPublisher := mocks2.NewMockEventPublisher(mockCtrl)
	eventIDStore := mocks.NewMockEventIDStore(mockCtrl)
	eventSource := mocks.NewMockEventSource(mockCtrl)

	firstEventID := "EVENT01"
	secondEventID := "EVENT02"
	secondEvent := []proton.Event{{
		EventID: secondEventID,
	}}

	// Event id store expectations.
	eventIDStore.EXPECT().Load(gomock.Any()).Times(1).Return("", nil)
	eventIDStore.EXPECT().Store(gomock.Any(), gomock.Eq(firstEventID)).Times(1).Return(nil)
	eventIDStore.EXPECT().Store(gomock.Any(), gomock.Eq(secondEventID)).Times(1).DoAndReturn(func(_ context.Context, _ string) error {
		// Force exit, we have finished executing what we expected.
		group.Cancel()
		return nil
	})

	// Event Source expectations.
	eventSource.EXPECT().GetLatestEventID(gomock.Any()).Times(1).Return(firstEventID, nil)
	eventSource.EXPECT().GetEvent(gomock.Any(), gomock.Eq(firstEventID)).MinTimes(1).Return(secondEvent, false, nil)

	service := NewService("foo", eventSource, eventIDStore, eventPublisher, 1*time.Millisecond, time.Second, async.NoopPanicHandler{})
	require.NoError(t, service.Start(context.Background(), group))
	service.Resume()
	group.WaitToFinish()
}

func TestService_RetryEventOnNonCatastrophicFailure(t *testing.T) {
	group := async.NewGroup(context.Background(), &async.NoopPanicHandler{})
	mockCtrl := gomock.NewController(t)
	eventPublisher := mocks2.NewMockEventPublisher(mockCtrl)
	eventIDStore := mocks.NewMockEventIDStore(mockCtrl)
	eventSource := mocks.NewMockEventSource(mockCtrl)
	subscriber := NewMockMessageSubscriber(mockCtrl)

	firstEventID := "EVENT01"
	secondEventID := "EVENT02"
	messageEvents := []proton.MessageEvent{
		{
			EventItem: proton.EventItem{ID: "Message"},
		},
	}
	secondEvent := []proton.Event{{
		EventID:  secondEventID,
		Messages: messageEvents,
	}}

	// Event id store expectations.
	eventIDStore.EXPECT().Load(gomock.Any()).Times(1).Return(firstEventID, nil)
	eventIDStore.EXPECT().Store(gomock.Any(), gomock.Eq(secondEventID)).Times(1).DoAndReturn(func(_ context.Context, _ string) error {
		// Force exit, we have finished executing what we expected.
		group.Cancel()
		return nil
	})

	// Event Source expectations.
	eventSource.EXPECT().GetEvent(gomock.Any(), gomock.Eq(firstEventID)).MinTimes(1).Return(secondEvent, false, nil)

	// Subscriber expectations.
	subscriber.EXPECT().name().AnyTimes().Return("Foo")
	{
		firstCall := subscriber.EXPECT().handle(gomock.Any(), gomock.Eq(messageEvents)).Times(1).Return(io.ErrUnexpectedEOF)
		subscriber.EXPECT().handle(gomock.Any(), gomock.Eq(messageEvents)).After(firstCall).Times(1).Return(nil)
	}

	service := NewService("foo", eventSource, eventIDStore, eventPublisher, 1*time.Millisecond, time.Second, async.NoopPanicHandler{})
	service.Subscribe(Subscription{Messages: subscriber})

	require.NoError(t, service.Start(context.Background(), group))
	service.Resume()
	group.WaitToFinish()
}

func TestService_OnBadEventServiceIsPaused(t *testing.T) {
	group := async.NewGroup(context.Background(), &async.NoopPanicHandler{})
	mockCtrl := gomock.NewController(t)
	eventPublisher := mocks2.NewMockEventPublisher(mockCtrl)
	eventIDStore := mocks.NewMockEventIDStore(mockCtrl)
	eventSource := mocks.NewMockEventSource(mockCtrl)
	subscriber := NewMockMessageSubscriber(mockCtrl)

	firstEventID := "EVENT01"
	secondEventID := "EVENT02"
	messageEvents := []proton.MessageEvent{
		{
			EventItem: proton.EventItem{ID: "Message"},
		},
	}
	secondEvent := []proton.Event{{
		EventID:  secondEventID,
		Messages: messageEvents,
	}}

	// Event id store expectations.
	eventIDStore.EXPECT().Load(gomock.Any()).Times(1).Return(firstEventID, nil)

	// Event Source expectations.
	eventSource.EXPECT().GetEvent(gomock.Any(), gomock.Eq(firstEventID)).MinTimes(1).Return(secondEvent, false, nil)

	// Subscriber expectations.
	badEventErr := fmt.Errorf("I will cause bad event")
	subscriber.EXPECT().name().AnyTimes().Return("Foo")
	subscriber.EXPECT().handle(gomock.Any(), gomock.Eq(messageEvents)).Times(1).Return(badEventErr)

	service := NewService("foo", eventSource, eventIDStore, eventPublisher, 1*time.Millisecond, time.Second, async.NoopPanicHandler{})

	// Event publisher expectations.
	eventPublisher.EXPECT().PublishEvent(gomock.Any(), events.UserBadEvent{
		UserID:     "foo",
		OldEventID: firstEventID,
		NewEventID: secondEventID,
		EventInfo:  secondEvent[0].String(),
		Error:      badEventErr,
	}).Do(func(_ context.Context, event events.Event) {
		group.Once(func(_ context.Context) {
			// Use background context to avoid having the request cancelled
			require.True(t, service.IsPaused())
			group.Cancel()
		})
	})

	service.Subscribe(Subscription{Messages: subscriber})
	require.NoError(t, service.Start(context.Background(), group))
	service.Resume()
	group.WaitToFinish()
}

func TestService_UnsubscribeDuringEventHandlingDoesNotCauseDeadlock(t *testing.T) {
	group := async.NewGroup(context.Background(), &async.NoopPanicHandler{})
	mockCtrl := gomock.NewController(t)
	eventPublisher := mocks2.NewMockEventPublisher(mockCtrl)
	eventIDStore := mocks.NewMockEventIDStore(mockCtrl)
	eventSource := mocks.NewMockEventSource(mockCtrl)
	subscriber := NewMockMessageSubscriber(mockCtrl)

	firstEventID := "EVENT01"
	secondEventID := "EVENT02"
	messageEvents := []proton.MessageEvent{
		{
			EventItem: proton.EventItem{ID: "Message"},
		},
	}
	secondEvent := []proton.Event{{
		EventID:  secondEventID,
		Messages: messageEvents,
	}}

	// Event id store expectations.
	eventIDStore.EXPECT().Load(gomock.Any()).Times(1).Return(firstEventID, nil)
	eventIDStore.EXPECT().Store(gomock.Any(), gomock.Eq(secondEventID)).Times(1).DoAndReturn(func(_ context.Context, _ string) error {
		// Force exit, we have finished executing what we expected.
		group.Cancel()
		return nil
	})

	// Event Source expectations.
	eventSource.EXPECT().GetEvent(gomock.Any(), gomock.Eq(firstEventID)).MinTimes(1).Return(secondEvent, false, nil)

	service := NewService("foo", eventSource, eventIDStore, eventPublisher, 1*time.Millisecond, time.Second, async.NoopPanicHandler{})

	// Subscriber expectations.
	subscriber.EXPECT().name().AnyTimes().Return("Foo")
	subscriber.EXPECT().cancel().Times(1)
	subscriber.EXPECT().handle(gomock.Any(), gomock.Eq(messageEvents)).Times(1).DoAndReturn(func(_ context.Context, _ []proton.MessageEvent) error {
		service.Unsubscribe(Subscription{Messages: subscriber})
		return nil
	})

	service.Subscribe(Subscription{Messages: subscriber})
	require.NoError(t, service.Start(context.Background(), group))
	service.Resume()
	group.WaitToFinish()
}

func TestService_UnsubscribeBeforeHandlingEventIsNotConsideredError(t *testing.T) {
	group := async.NewGroup(context.Background(), &async.NoopPanicHandler{})
	mockCtrl := gomock.NewController(t)
	eventPublisher := mocks2.NewMockEventPublisher(mockCtrl)
	eventIDStore := mocks.NewMockEventIDStore(mockCtrl)
	eventSource := mocks.NewMockEventSource(mockCtrl)
	subscriber := NewMessageSubscriber("My subscriber")

	firstEventID := "EVENT01"
	secondEventID := "EVENT02"
	messageEvents := []proton.MessageEvent{
		{
			EventItem: proton.EventItem{ID: "Message"},
		},
	}
	secondEvent := []proton.Event{{
		EventID:  secondEventID,
		Messages: messageEvents,
	}}

	// Event id store expectations.
	eventIDStore.EXPECT().Load(gomock.Any()).Times(1).Return(firstEventID, nil)
	eventIDStore.EXPECT().Store(gomock.Any(), gomock.Eq(secondEventID)).Times(1).DoAndReturn(func(_ context.Context, _ string) error {
		// Force exit, we have finished executing what we expected.
		group.Cancel()
		return nil
	})

	// Event Source expectations.
	eventSource.EXPECT().GetEvent(gomock.Any(), gomock.Eq(firstEventID)).MinTimes(1).Return(secondEvent, false, nil)
	eventSource.EXPECT().GetEvent(gomock.Any(), gomock.Eq(secondEventID)).AnyTimes().Return(secondEvent, false, nil)

	service := NewService("foo", eventSource, eventIDStore, eventPublisher, 1*time.Millisecond, time.Second, async.NoopPanicHandler{})

	// start subscriber
	group.Once(func(_ context.Context) {
		defer service.Unsubscribe(Subscription{Messages: subscriber})

		// Simulate the reception of an event, but it is never handled due to unexpected exit
		<-time.NewTicker(500 * time.Millisecond).C
	})

	service.Subscribe(Subscription{Messages: subscriber})
	require.NoError(t, service.Start(context.Background(), group))
	service.Resume()
	group.WaitToFinish()
}
