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

package useridentity

import (
	"context"
	"testing"

	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/proton-bridge/v3/internal/events"
	mocks2 "github.com/ProtonMail/proton-bridge/v3/internal/events/mocks"
	"github.com/ProtonMail/proton-bridge/v3/internal/services/userevents"
	"github.com/ProtonMail/proton-bridge/v3/internal/services/useridentity/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

const TestUserID = "MyUserID"

func TestService_OnUserEvent(t *testing.T) {
	mockCtrl := gomock.NewController(t)

	service, eventPublisher, _ := newTestService(t, mockCtrl)

	eventPublisher.EXPECT().PublishEvent(gomock.Any(), gomock.Eq(events.UserChanged{UserID: TestUserID})).Times(1)

	require.NoError(t, service.HandleUserEvent(context.Background(), newTestUser()))
}

func TestService_OnUserSpaceChanged(t *testing.T) {
	mockCtrl := gomock.NewController(t)

	service, eventPublisher, _ := newTestService(t, mockCtrl)

	eventPublisher.EXPECT().PublishEvent(gomock.Any(), gomock.Eq(events.UsedSpaceChanged{UserID: TestUserID, UsedSpace: 1024})).Times(1)

	// Original value, no changes.
	require.NoError(t, service.HandleUsedSpaceEvent(context.Background(), 0))

	// New value, event should be published.
	require.NoError(t, service.HandleUsedSpaceEvent(context.Background(), 1024))
	require.Equal(t, uint64(1024), service.identity.User.UsedSpace)
}

func TestService_OnRefreshEvent(t *testing.T) {
	mockCtrl := gomock.NewController(t)

	service, eventPublisher, provider := newTestService(t, mockCtrl)

	eventPublisher.EXPECT().PublishEvent(gomock.Any(), gomock.Eq(events.UserRefreshed{UserID: TestUserID, CancelEventPool: false})).Times(1)

	newUser := newTestUserRefreshed()
	newAddresses := newTestAddressesRefreshed()

	{
		getUserCall := provider.EXPECT().GetUser(gomock.Any()).Times(1).Return(*newUser, nil)
		provider.EXPECT().GetAddresses(gomock.Any()).After(getUserCall).Times(1).Return(newAddresses, nil)
	}

	// Original value, no changes.
	require.NoError(t, service.HandleRefreshEvent(context.Background(), 0))

	require.Equal(t, *newUser, service.identity.User)
	require.Equal(t, newAddresses, service.identity.AddressesSorted)
}

func TestService_OnAddressCreated(t *testing.T) {
	mockCtrl := gomock.NewController(t)

	service, eventPublisher, _ := newTestService(t, mockCtrl)

	newAddress := proton.Address{
		ID:     "NewAddrID",
		Email:  "new@bar.com",
		Status: proton.AddressStatusEnabled,
	}

	eventPublisher.EXPECT().PublishEvent(gomock.Any(), gomock.Eq(events.UserAddressCreated{
		UserID:    TestUserID,
		AddressID: newAddress.ID,
		Email:     newAddress.Email,
	})).Times(1)

	err := service.HandleAddressEvents(context.Background(), []proton.AddressEvent{
		{
			EventItem: proton.EventItem{
				ID:     "",
				Action: proton.EventCreate,
			},
			Address: newAddress,
		},
	})
	require.NoError(t, err)

	require.Contains(t, service.identity.Addresses, newAddress.ID)
}

func TestService_OnAddressCreatedDisabledDoesNotProduceEvent(t *testing.T) {
	mockCtrl := gomock.NewController(t)

	service, _, _ := newTestService(t, mockCtrl)

	newAddress := proton.Address{
		ID:     "Address1",
		Email:  "new@bar.com",
		Status: proton.AddressStatusEnabled,
	}

	err := service.HandleAddressEvents(context.Background(), []proton.AddressEvent{
		{
			EventItem: proton.EventItem{
				ID:     "",
				Action: proton.EventCreate,
			},
			Address: newAddress,
		},
	})
	require.NoError(t, err)

	require.Contains(t, service.identity.Addresses, newAddress.ID)
}

func TestService_OnAddressCreatedDuplicateDoesNotProduceEvent(t *testing.T) {
	mockCtrl := gomock.NewController(t)

	service, _, _ := newTestService(t, mockCtrl)

	newAddress := proton.Address{
		ID:     "NewAddrID",
		Email:  "new@bar.com",
		Status: proton.AddressStatusDisabled,
	}

	err := service.HandleAddressEvents(context.Background(), []proton.AddressEvent{
		{
			EventItem: proton.EventItem{
				ID:     "",
				Action: proton.EventCreate,
			},
			Address: newAddress,
		},
	})
	require.NoError(t, err)

	require.Contains(t, service.identity.Addresses, newAddress.ID)
}

func TestService_OnAddressUpdated(t *testing.T) {
	mockCtrl := gomock.NewController(t)

	service, eventPublisher, _ := newTestService(t, mockCtrl)

	newAddress := proton.Address{
		ID:     "Address1",
		Email:  "new@bar.com",
		Status: proton.AddressStatusEnabled,
	}

	eventPublisher.EXPECT().PublishEvent(gomock.Any(), gomock.Eq(events.UserAddressUpdated{
		UserID:    TestUserID,
		AddressID: newAddress.ID,
		Email:     newAddress.Email,
	})).Times(1)

	err := service.HandleAddressEvents(context.Background(), []proton.AddressEvent{
		{
			EventItem: proton.EventItem{
				ID:     "",
				Action: proton.EventUpdate,
			},
			Address: newAddress,
		},
	})
	require.NoError(t, err)

	require.Equal(t, newAddress, service.identity.Addresses[newAddress.ID])
}

func TestService_OnAddressUpdatedDisableFollowedByEnable(t *testing.T) {
	mockCtrl := gomock.NewController(t)

	service, eventPublisher, _ := newTestService(t, mockCtrl)

	newAddressDisabled := proton.Address{
		ID:     "Address1",
		Email:  "new@bar.com",
		Status: proton.AddressStatusDisabled,
	}
	newAddressEnabled := proton.Address{
		ID:     "Address1",
		Email:  "new@bar.com",
		Status: proton.AddressStatusEnabled,
	}

	{
		disabledCall := eventPublisher.EXPECT().PublishEvent(gomock.Any(), gomock.Eq(events.UserAddressDisabled{
			UserID:    TestUserID,
			AddressID: newAddressDisabled.ID,
			Email:     newAddressDisabled.Email,
		})).Times(1)

		eventPublisher.EXPECT().PublishEvent(gomock.Any(), gomock.Eq(events.UserAddressEnabled{
			UserID:    TestUserID,
			AddressID: newAddressEnabled.ID,
			Email:     newAddressEnabled.Email,
		})).Times(1).After(disabledCall)
	}

	err := service.HandleAddressEvents(context.Background(), []proton.AddressEvent{
		{
			EventItem: proton.EventItem{
				ID:     "",
				Action: proton.EventUpdate,
			},
			Address: newAddressDisabled,
		},
	})
	require.NoError(t, err)

	require.Equal(t, newAddressDisabled, service.identity.Addresses[newAddressEnabled.ID])

	err = service.HandleAddressEvents(context.Background(), []proton.AddressEvent{
		{
			EventItem: proton.EventItem{
				ID:     "",
				Action: proton.EventUpdate,
			},
			Address: newAddressEnabled,
		},
	})
	require.NoError(t, err)

	require.Equal(t, newAddressEnabled, service.identity.Addresses[newAddressEnabled.ID])
}

func TestService_OnAddressUpdateCreatedIfNotExists(t *testing.T) {
	mockCtrl := gomock.NewController(t)

	service, eventPublisher, _ := newTestService(t, mockCtrl)

	newAddress := proton.Address{
		ID:     "NewAddrID",
		Email:  "new@bar.com",
		Status: proton.AddressStatusEnabled,
	}

	eventPublisher.EXPECT().PublishEvent(gomock.Any(), gomock.Eq(events.UserAddressCreated{
		UserID:    TestUserID,
		AddressID: newAddress.ID,
		Email:     newAddress.Email,
	})).Times(1)

	err := service.HandleAddressEvents(context.Background(), []proton.AddressEvent{
		{
			EventItem: proton.EventItem{
				ID:     "",
				Action: proton.EventUpdate,
			},
			Address: newAddress,
		},
	})
	require.NoError(t, err)

	require.Contains(t, service.identity.Addresses, newAddress.ID)
}

func TestService_OnAddressDeleted(t *testing.T) {
	mockCtrl := gomock.NewController(t)

	service, eventPublisher, _ := newTestService(t, mockCtrl)

	address := proton.Address{
		ID:     "Address1",
		Email:  "foo@bar.com",
		Status: proton.AddressStatusEnabled,
	}

	eventPublisher.EXPECT().PublishEvent(gomock.Any(), gomock.Eq(events.UserAddressDeleted{
		UserID:    TestUserID,
		AddressID: address.ID,
		Email:     address.Email,
	})).Times(1)

	err := service.HandleAddressEvents(context.Background(), []proton.AddressEvent{
		{
			EventItem: proton.EventItem{
				ID:     address.ID,
				Action: proton.EventDelete,
			},
		},
	})
	require.NoError(t, err)

	require.NotContains(t, service.identity.Addresses, address.ID)
}

func TestService_OnAddressDeleteDisabledDoesNotProduceEvent(t *testing.T) {
	mockCtrl := gomock.NewController(t)

	service, _, _ := newTestService(t, mockCtrl)

	address := proton.Address{
		ID:     "Address2",
		Email:  "foo2@bar.com",
		Status: proton.AddressStatusDisabled,
	}

	err := service.HandleAddressEvents(context.Background(), []proton.AddressEvent{
		{
			EventItem: proton.EventItem{
				ID:     address.ID,
				Action: proton.EventDelete,
			},
		},
	})
	require.NoError(t, err)

	require.NotContains(t, service.identity.Addresses, address.ID)
}

func TestService_OnAddressDeletedUnknownDoesNotProduceEvent(t *testing.T) {
	mockCtrl := gomock.NewController(t)

	service, _, _ := newTestService(t, mockCtrl)

	address := proton.Address{
		ID:     "UnknownID",
		Email:  "new@bar.com",
		Status: proton.AddressStatusEnabled,
	}

	err := service.HandleAddressEvents(context.Background(), []proton.AddressEvent{
		{
			EventItem: proton.EventItem{
				ID:     address.ID,
				Action: proton.EventDelete,
			},
			Address: address,
		},
	})
	require.NoError(t, err)
}

func newTestService(_ *testing.T, mockCtrl *gomock.Controller) (*Service, *mocks2.MockEventPublisher, *mocks.MockIdentityProvider) {
	subscribable := &userevents.NoOpSubscribable{}
	eventPublisher := mocks2.NewMockEventPublisher(mockCtrl)
	provider := mocks.NewMockIdentityProvider(mockCtrl)
	user := newTestUser()
	telemetry := mocks.NewMockTelemetry(mockCtrl)
	bridgePassProvider := NewFixedBridgePassProvider([]byte("hello"))

	service := NewService(subscribable, eventPublisher, NewState(*user, newTestAddresses(), provider), bridgePassProvider, telemetry)
	return service, eventPublisher, provider
}

func newTestUser() *proton.User {
	return &proton.User{
		ID:          TestUserID,
		Name:        "Foo",
		DisplayName: "Foo",
		Email:       "foo@bar",
		Keys:        nil,
		UsedSpace:   0,
		MaxSpace:    0,
		MaxUpload:   0,
		Credit:      0,
		Currency:    "",
	}
}

func newTestUserRefreshed() *proton.User {
	return &proton.User{
		ID:          TestUserID,
		Name:        "Alternate",
		DisplayName: "Universe",
		Email:       "foo2@bar",
		Keys:        nil,
		UsedSpace:   0,
		MaxSpace:    0,
		MaxUpload:   0,
		Credit:      0,
		Currency:    "USD",
	}
}

func newTestAddresses() []proton.Address {
	return []proton.Address{
		{
			ID:          "Address1",
			Email:       "foo@bar.com",
			Status:      proton.AddressStatusEnabled,
			Type:        0,
			Order:       0,
			DisplayName: "",
			Keys:        nil,
		},
		{
			ID:          "Address2",
			Email:       "foo2@bar.com",
			Status:      proton.AddressStatusDisabled,
			Type:        0,
			Order:       1,
			DisplayName: "",
			Keys:        nil,
		},
	}
}

func newTestAddressesRefreshed() []proton.Address {
	return []proton.Address{
		{
			ID:          "Address1",
			Email:       "foo@bar.com",
			Status:      proton.AddressStatusEnabled,
			Type:        0,
			Order:       2,
			DisplayName: "FOo barish",
			Keys:        nil,
		},
		{
			ID:          "Address2",
			Email:       "foo2@bar.com",
			Status:      proton.AddressStatusDisabled,
			Type:        0,
			Order:       4,
			DisplayName: "New display name",
			Keys:        nil,
		},
	}
}
