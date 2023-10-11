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
	"github.com/ProtonMail/proton-bridge/v3/internal/events"
	"github.com/ProtonMail/proton-bridge/v3/internal/events/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestServiceHandleEvent_CheckEventCategoriesHandledInOrder(t *testing.T) {
	mockCtrl := gomock.NewController(t)

	eventPublisher := mocks.NewMockEventPublisher(mockCtrl)
	eventIDStore := NewInMemoryEventIDStore()

	refreshHandler := NewMockRefreshEventHandler(mockCtrl)
	refreshHandler.EXPECT().HandleRefreshEvent(gomock.Any(), gomock.Any()).Times(2).Return(nil)

	userHandler := NewMockUserEventHandler(mockCtrl)
	userCall := userHandler.EXPECT().HandleUserEvent(gomock.Any(), gomock.Any()).Times(1).Return(nil)

	addressHandler := NewMockAddressEventHandler(mockCtrl)
	addressCall := addressHandler.EXPECT().HandleAddressEvents(gomock.Any(), gomock.Any()).After(userCall).Times(1).Return(nil)

	labelHandler := NewMockLabelEventHandler(mockCtrl)
	labelCall := labelHandler.EXPECT().HandleLabelEvents(gomock.Any(), gomock.Any()).After(addressCall).Times(1).Return(nil)

	messageHandler := NewMockMessageEventHandler(mockCtrl)
	messageCall := messageHandler.EXPECT().HandleMessageEvents(gomock.Any(), gomock.Any()).After(labelCall).Times(1).Return(nil)

	userSpaceHandler := NewMockUserUsedSpaceEventHandler(mockCtrl)
	userSpaceCall := userSpaceHandler.EXPECT().HandleUsedSpaceEvent(gomock.Any(), gomock.Any()).After(messageCall).Times(1).Return(nil)

	secondRefreshHandler := NewMockRefreshEventHandler(mockCtrl)
	secondRefreshHandler.EXPECT().HandleRefreshEvent(gomock.Any(), gomock.Any()).After(userSpaceCall).Times(1).Return(nil)

	service := NewService(
		"foo",
		&NullEventSource{},
		eventIDStore,
		eventPublisher,
		100*time.Millisecond,
		time.Millisecond,
		10*time.Second,
		async.NoopPanicHandler{},
		events.NewNullSubscription(),
	)

	subscription := NewCallbackSubscriber("test", EventHandler{
		UserHandler:      userHandler,
		RefreshHandler:   refreshHandler,
		AddressHandler:   addressHandler,
		LabelHandler:     labelHandler,
		MessageHandler:   messageHandler,
		UsedSpaceHandler: userSpaceHandler,
	})

	service.addSubscription(subscription)

	// Simulate 1st refresh.
	require.NoError(t, service.handleEvent(context.Background(), "", proton.Event{Refresh: proton.RefreshMail}))

	// Simulate Regular event.
	usedSpace := int64(20)
	require.NoError(t, service.handleEvent(context.Background(), "", proton.Event{
		User: new(proton.User),
		Addresses: []proton.AddressEvent{
			{},
		},
		Labels: []proton.LabelEvent{
			{},
		},
		Messages: []proton.MessageEvent{
			{},
		},
		UsedSpace: &usedSpace,
	}))

	service.addSubscription(NewCallbackSubscriber("test", EventHandler{
		RefreshHandler: secondRefreshHandler,
	}))

	// Simulate 2nd refresh.
	require.NoError(t, service.handleEvent(context.Background(), "", proton.Event{Refresh: proton.RefreshMail}))
}

func TestServiceHandleEvent_CheckEventFailureCausesError(t *testing.T) {
	mockCtrl := gomock.NewController(t)

	eventPublisher := mocks.NewMockEventPublisher(mockCtrl)
	eventIDStore := NewInMemoryEventIDStore()

	addressHandler := NewMockAddressEventHandler(mockCtrl)
	addressHandler.EXPECT().HandleAddressEvents(gomock.Any(), gomock.Any()).Times(1).Return(fmt.Errorf("failed"))

	messageHandler := NewMockMessageEventHandler(mockCtrl)

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

	subscription := NewCallbackSubscriber("test", EventHandler{
		AddressHandler: addressHandler,
		MessageHandler: messageHandler,
	})

	service.addSubscription(subscription)

	err := service.handleEvent(context.Background(), "", proton.Event{Addresses: []proton.AddressEvent{{}}})
	require.Error(t, err)
	publisherErr := new(eventPublishError)
	require.True(t, errors.As(err, &publisherErr))
	require.Equal(t, publisherErr.subscriber, subscription)
}

func TestServiceHandleEvent_CheckEventFailureCausesErrorParallel(t *testing.T) {
	mockCtrl := gomock.NewController(t)

	eventPublisher := mocks.NewMockEventPublisher(mockCtrl)
	eventIDStore := NewInMemoryEventIDStore()

	addressHandler := NewMockAddressEventHandler(mockCtrl)
	addressHandler.EXPECT().HandleAddressEvents(gomock.Any(), gomock.Any()).Times(1).Return(fmt.Errorf("failed"))

	addressHandler2 := NewMockAddressEventHandler(mockCtrl)
	addressHandler2.EXPECT().HandleAddressEvents(gomock.Any(), gomock.Any()).MaxTimes(1).Return(nil)

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

	subscription := NewCallbackSubscriber("test", EventHandler{
		AddressHandler: addressHandler,
	})

	service.addSubscription(subscription)

	service.addSubscription(NewCallbackSubscriber("test2", EventHandler{
		AddressHandler: addressHandler2,
	}))

	err := service.handleEvent(context.Background(), "", proton.Event{Addresses: []proton.AddressEvent{{}}})
	require.Error(t, err)
	publisherErr := new(eventPublishError)
	require.True(t, errors.As(err, &publisherErr))
	require.Equal(t, publisherErr.subscriber, subscription)
}
