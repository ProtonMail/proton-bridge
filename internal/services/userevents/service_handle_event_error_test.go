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
	"encoding/json"
	"errors"
	"io"
	"net"
	"testing"
	"time"

	"github.com/ProtonMail/gluon/async"
	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/proton-bridge/v3/internal/events"
	"github.com/ProtonMail/proton-bridge/v3/internal/events/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestServiceHandleEventError_SubscriberEventUnwrapping(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	eventPublisher := mocks.NewMockEventPublisher(mockCtrl)
	eventIDStore := NewInMemoryEventIDStore()

	service := NewService(
		"foo",
		&NullEventSource{},
		eventIDStore,
		eventPublisher,
		100*time.Millisecond,
		time.Millisecond,
		time.Second,
		async.NoopPanicHandler{},
		events.NewNullSubscription(),
	)

	lastEventID := "PrevEvent"
	event := proton.Event{EventID: "MyEvent"}
	subscriber := &noOpSubscriber[proton.Event]{}

	err := &eventPublishError{
		subscriber: subscriber,
		error:      &proton.NetError{},
	}

	subscriberName, unpackedErr := service.handleEventError(context.Background(), lastEventID, event, err)
	require.Equal(t, subscriber.name(), subscriberName)
	protonNetErr := new(proton.NetError)
	require.True(t, errors.As(unpackedErr, &protonNetErr))

	err2 := &proton.APIError{Status: 500}
	subscriberName2, unpackedErr2 := service.handleEventError(context.Background(), lastEventID, event, err2)
	require.Equal(t, "", subscriberName2)
	protonAPIErr := new(proton.APIError)
	require.True(t, errors.As(unpackedErr2, &protonAPIErr))
}

func TestServiceHandleEventError_BadEventPutsServiceOnPause(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	eventPublisher := mocks.NewMockEventPublisher(mockCtrl)
	eventIDStore := NewInMemoryEventIDStore()

	service := NewService(
		"foo",
		&NullEventSource{},
		eventIDStore,
		eventPublisher,
		100*time.Millisecond,
		time.Millisecond,
		time.Second,
		async.NoopPanicHandler{},
		events.NewNullSubscription(),
	)
	service.Resume()
	lastEventID := "PrevEvent"
	event := proton.Event{EventID: "MyEvent"}

	err := &proton.APIError{}

	eventPublisher.EXPECT().PublishEvent(gomock.Any(), gomock.Eq(events.UserBadEvent{
		UserID:     service.userID,
		OldEventID: lastEventID,
		NewEventID: event.EventID,
		EventInfo:  event.String(),
		Error:      err,
	})).Times(1)

	_, _ = service.handleEventError(context.Background(), lastEventID, event, err)
	require.True(t, service.IsPaused())
}

func TestServiceHandleEventError_BadEventFromPublishTimeout(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	eventPublisher := mocks.NewMockEventPublisher(mockCtrl)
	eventIDStore := NewInMemoryEventIDStore()

	service := NewService(
		"foo",
		&NullEventSource{},
		eventIDStore,
		eventPublisher,
		100*time.Millisecond,
		time.Millisecond,
		time.Second,
		async.NoopPanicHandler{},
		events.NewNullSubscription(),
	)
	lastEventID := "PrevEvent"
	event := proton.Event{EventID: "MyEvent"}
	err := ErrPublishTimeoutExceeded

	eventPublisher.EXPECT().PublishEvent(gomock.Any(), gomock.Eq(events.UserBadEvent{
		UserID:     service.userID,
		OldEventID: lastEventID,
		NewEventID: event.EventID,
		EventInfo:  event.String(),
		Error:      err,
	})).Times(1)

	_, _ = service.handleEventError(context.Background(), lastEventID, event, err)
}

func TestServiceHandleEventError_NoBadEventCheck(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	eventPublisher := mocks.NewMockEventPublisher(mockCtrl)
	eventIDStore := NewInMemoryEventIDStore()

	service := NewService(
		"foo",
		&NullEventSource{},
		eventIDStore,
		eventPublisher,
		100*time.Millisecond,
		time.Millisecond,
		time.Second,
		async.NoopPanicHandler{},
		events.NewNullSubscription(),
	)
	lastEventID := "PrevEvent"
	event := proton.Event{EventID: "MyEvent"}
	_, _ = service.handleEventError(context.Background(), lastEventID, event, context.Canceled)
	_, _ = service.handleEventError(context.Background(), lastEventID, event, &proton.NetError{})
	_, _ = service.handleEventError(context.Background(), lastEventID, event, &net.OpError{})
	_, _ = service.handleEventError(context.Background(), lastEventID, event, io.ErrUnexpectedEOF)
	_, _ = service.handleEventError(context.Background(), lastEventID, event, &proton.APIError{Status: 500})
	_, _ = service.handleEventError(context.Background(), lastEventID, event, &proton.APIError{Status: 429})
}

func TestServiceHandleEventError_JsonUnmarshalEventProducesUncategorizedErrorEvent(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	eventPublisher := mocks.NewMockEventPublisher(mockCtrl)
	eventIDStore := NewInMemoryEventIDStore()

	service := NewService(
		"foo",
		&NullEventSource{},
		eventIDStore,
		eventPublisher,
		100*time.Millisecond,
		time.Millisecond,
		time.Second,
		async.NoopPanicHandler{},
		events.NewNullSubscription(),
	)
	lastEventID := "PrevEvent"
	event := proton.Event{EventID: "MyEvent"}
	err := &json.UnmarshalTypeError{}

	eventPublisher.EXPECT().PublishEvent(gomock.Any(), gomock.Eq(events.UncategorizedEventError{
		UserID: service.userID,
		Error:  err,
	})).Times(1)

	_, _ = service.handleEventError(context.Background(), lastEventID, event, err)
}

type noOpSubscriber[T any] struct{}

func (n noOpSubscriber[T]) name() string { //nolint:unused
	return "NoopSubscriber"
}

func (n noOpSubscriber[T]) handle(_ context.Context, _ T) error { //nolint:unused
	return nil
}

//nolint:unused
func (n noOpSubscriber[T]) close() {} //

//nolint:unused
func (n noOpSubscriber[T]) cancel() {}
