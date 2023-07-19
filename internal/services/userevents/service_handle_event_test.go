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
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/ProtonMail/gluon/async"
	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/proton-bridge/v3/internal/events/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestServiceHandleEvent_CheckEventCategoriesHandledInOrder(t *testing.T) {
	mockCtrl := gomock.NewController(t)

	eventPublisher := mocks.NewMockEventPublisher(mockCtrl)
	eventIDStore := NewInMemoryEventIDStore()

	refreshHandler := NewMockRefreshSubscriber(mockCtrl)
	refreshHandler.EXPECT().handle(gomock.Any(), gomock.Any()).Times(2).Return(nil)

	userHandler := NewMockUserSubscriber(mockCtrl)
	userCall := userHandler.EXPECT().handle(gomock.Any(), gomock.Any()).Times(1).Return(nil)

	addressHandler := NewMockAddressSubscriber(mockCtrl)
	addressCall := addressHandler.EXPECT().handle(gomock.Any(), gomock.Any()).After(userCall).Times(1).Return(nil)

	labelHandler := NewMockLabelSubscriber(mockCtrl)
	labelCall := labelHandler.EXPECT().handle(gomock.Any(), gomock.Any()).After(addressCall).Times(1).Return(nil)

	messageHandler := NewMockMessageSubscriber(mockCtrl)
	messageCall := messageHandler.EXPECT().handle(gomock.Any(), gomock.Any()).After(labelCall).Times(1).Return(nil)

	userSpaceHandler := NewMockUserUsedSpaceSubscriber(mockCtrl)
	userSpaceCall := userSpaceHandler.EXPECT().handle(gomock.Any(), gomock.Any()).After(messageCall).Times(1).Return(nil)

	secondRefreshHandler := NewMockRefreshSubscriber(mockCtrl)
	secondRefreshHandler.EXPECT().handle(gomock.Any(), gomock.Any()).After(userSpaceCall).Times(1).Return(nil)

	service := NewService("foo", &NullEventSource{}, eventIDStore, eventPublisher, 100*time.Millisecond, time.Second, async.NoopPanicHandler{})

	service.addSubscription(Subscription{
		User:          userHandler,
		Refresh:       refreshHandler,
		Address:       addressHandler,
		Labels:        labelHandler,
		Messages:      messageHandler,
		UserUsedSpace: userSpaceHandler,
	})

	// Simulate 1st refresh.
	require.NoError(t, service.handleEvent(context.Background(), "", proton.Event{Refresh: proton.RefreshMail}))

	// Simulate Regular event.
	usedSpace := 20
	require.NoError(t, service.handleEvent(context.Background(), "", proton.Event{
		User:      new(proton.User),
		Addresses: []proton.AddressEvent{},
		Labels: []proton.LabelEvent{
			{},
		},
		Messages: []proton.MessageEvent{
			{},
		},
		UsedSpace: &usedSpace,
	}))

	service.addSubscription(Subscription{
		Refresh: secondRefreshHandler,
	})

	// Simulate 2nd refresh.
	require.NoError(t, service.handleEvent(context.Background(), "", proton.Event{Refresh: proton.RefreshMail}))
}

func TestServiceHandleEvent_CheckEventFailureCausesError(t *testing.T) {
	mockCtrl := gomock.NewController(t)

	eventPublisher := mocks.NewMockEventPublisher(mockCtrl)
	eventIDStore := NewInMemoryEventIDStore()

	addressHandler := NewMockAddressSubscriber(mockCtrl)
	addressHandler.EXPECT().name().MinTimes(1).Return("Hello")
	addressHandler.EXPECT().handle(gomock.Any(), gomock.Any()).Times(1).Return(fmt.Errorf("failed"))

	messageHandler := NewMockMessageSubscriber(mockCtrl)

	service := NewService("foo", &NullEventSource{}, eventIDStore, eventPublisher, 100*time.Millisecond, time.Second, async.NoopPanicHandler{})

	service.addSubscription(Subscription{
		Address:  addressHandler,
		Messages: messageHandler,
	})

	err := service.handleEvent(context.Background(), "", proton.Event{Addresses: []proton.AddressEvent{{}}})
	require.Error(t, err)
	publisherErr := new(addressPublishError)
	require.True(t, errors.As(err, &publisherErr))
	require.Equal(t, publisherErr.subscriber, addressHandler)
}

func TestServiceHandleEvent_CheckEventFailureCausesErrorParallel(t *testing.T) {
	mockCtrl := gomock.NewController(t)

	eventPublisher := mocks.NewMockEventPublisher(mockCtrl)
	eventIDStore := NewInMemoryEventIDStore()

	addressHandler := NewMockAddressSubscriber(mockCtrl)
	addressHandler.EXPECT().name().MinTimes(1).Return("Hello")
	addressHandler.EXPECT().handle(gomock.Any(), gomock.Any()).Times(1).Return(fmt.Errorf("failed"))

	addressHandler2 := NewMockAddressSubscriber(mockCtrl)
	addressHandler2.EXPECT().handle(gomock.Any(), gomock.Any()).MaxTimes(1).Return(nil)

	service := NewService("foo", &NullEventSource{}, eventIDStore, eventPublisher, 100*time.Millisecond, time.Second, async.NoopPanicHandler{})

	service.addSubscription(Subscription{
		Address: addressHandler,
	})

	service.addSubscription(Subscription{
		Address: addressHandler2,
	})

	err := service.handleEvent(context.Background(), "", proton.Event{Addresses: []proton.AddressEvent{{}}})
	require.Error(t, err)
	publisherErr := new(addressPublishError)
	require.True(t, errors.As(err, &publisherErr))
	require.Equal(t, publisherErr.subscriber, addressHandler)
}

func TestServiceHandleEvent_SubscriberTimeout(t *testing.T) {
	mockCtrl := gomock.NewController(t)

	eventPublisher := mocks.NewMockEventPublisher(mockCtrl)
	eventIDStore := NewInMemoryEventIDStore()

	addressHandler := NewMockAddressSubscriber(mockCtrl)
	addressHandler.EXPECT().name().AnyTimes().Return("Ok")
	addressHandler.EXPECT().handle(gomock.Any(), gomock.Any()).MaxTimes(1).Return(nil)

	addressHandler2 := NewMockAddressSubscriber(mockCtrl)
	addressHandler2.EXPECT().name().AnyTimes().Return("Timeout")
	addressHandler2.EXPECT().handle(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, _ []proton.AddressEvent) error {
		timer := time.NewTimer(time.Second)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			return nil
		}
	}).MaxTimes(1)

	service := NewService("foo", &NullEventSource{}, eventIDStore, eventPublisher, 100*time.Millisecond, 500*time.Millisecond, async.NoopPanicHandler{})

	service.addSubscription(Subscription{
		Address: addressHandler,
	})

	service.addSubscription(Subscription{
		Address: addressHandler2,
	})

	// Simulate 1st refresh.
	err := service.handleEvent(context.Background(), "", proton.Event{Addresses: []proton.AddressEvent{{}}})
	require.Error(t, err)
	if publisherErr := new(addressPublishError); errors.As(err, &publisherErr) {
		require.Equal(t, publisherErr.subscriber, addressHandler)
		require.True(t, errors.Is(publisherErr.error, ErrPublishTimeoutExceeded))
	} else {
		require.True(t, errors.Is(err, ErrPublishTimeoutExceeded))
	}
}
