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
	"fmt"

	"github.com/ProtonMail/gluon/async"
	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/proton-bridge/v3/internal/events"
	"github.com/ProtonMail/proton-bridge/v3/internal/logging"
	"github.com/ProtonMail/proton-bridge/v3/internal/services/userevents"
	"github.com/ProtonMail/proton-bridge/v3/internal/usertypes"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
)

type IdentityProvider interface {
	GetUser(ctx context.Context) (proton.User, error)
	GetAddresses(ctx context.Context) ([]proton.Address, error)
}

// Service contains all the data required to establish the user identity. This
// includes all the user's information as well as mail addresses and keys.
type Service struct {
	eventService   userevents.Subscribable
	eventPublisher events.EventPublisher
	log            *logrus.Entry
	identity       State

	userSubscriber      *userevents.UserChanneledSubscriber
	addressSubscriber   *userevents.AddressChanneledSubscriber
	usedSpaceSubscriber *userevents.UserUsedSpaceChanneledSubscriber
	refreshSubscriber   *userevents.RefreshChanneledSubscriber
}

func NewService(
	service userevents.Subscribable,
	eventPublisher events.EventPublisher,
	state *State,
) *Service {
	subscriberName := fmt.Sprintf("identity-%v", state.User.ID)

	return &Service{
		eventService:   service,
		identity:       *state,
		eventPublisher: eventPublisher,
		log: logrus.WithFields(logrus.Fields{
			"service": "user-identity",
			"user":    state.User.ID,
		}),
		userSubscriber:      userevents.NewUserSubscriber(subscriberName),
		refreshSubscriber:   userevents.NewRefreshSubscriber(subscriberName),
		addressSubscriber:   userevents.NewAddressSubscriber(subscriberName),
		usedSpaceSubscriber: userevents.NewUserUsedSpaceSubscriber(subscriberName),
	}
}

func (s *Service) Start(group *async.Group) {
	group.Once(func(ctx context.Context) {
		s.run(ctx)
	})
}

func (s *Service) run(ctx context.Context) {
	s.log.WithFields(logrus.Fields{
		"numAddr": len(s.identity.Addresses),
	}).Info("Starting user identity service")

	s.registerSubscription()
	defer s.unregisterSubscription()

	for {
		select {
		case <-ctx.Done():
			return
		case evt, ok := <-s.userSubscriber.OnEventCh():
			if !ok {
				continue
			}
			evt.Consume(func(user proton.User) error {
				s.onUserEvent(ctx, user)
				return nil
			})
		case evt, ok := <-s.refreshSubscriber.OnEventCh():
			if !ok {
				continue
			}
			evt.Consume(func(_ proton.RefreshFlag) error {
				return s.onRefreshEvent(ctx)
			})
		case evt, ok := <-s.usedSpaceSubscriber.OnEventCh():
			if !ok {
				continue
			}
			evt.Consume(func(usedSpace int) error {
				s.onUserSpaceChanged(ctx, usedSpace)

				return nil
			})
		case evt, ok := <-s.addressSubscriber.OnEventCh():
			if !ok {
				continue
			}
			evt.Consume(func(events []proton.AddressEvent) error {
				return s.onAddressEvent(ctx, events)
			})
		}
	}
}

func (s *Service) registerSubscription() {
	s.eventService.Subscribe(userevents.Subscription{
		Refresh:       s.refreshSubscriber,
		User:          s.userSubscriber,
		Address:       s.addressSubscriber,
		UserUsedSpace: s.usedSpaceSubscriber,
	})
}

func (s *Service) unregisterSubscription() {
	s.eventService.Unsubscribe(userevents.Subscription{
		Refresh:       s.refreshSubscriber,
		User:          s.userSubscriber,
		Address:       s.addressSubscriber,
		UserUsedSpace: s.usedSpaceSubscriber,
	})
}

func (s *Service) onUserEvent(ctx context.Context, user proton.User) {
	s.log.WithField("username", logging.Sensitive(user.Name)).Info("Handling user event")
	s.identity.OnUserEvent(user)
	s.eventPublisher.PublishEvent(ctx, events.UserChanged{
		UserID: user.ID,
	})
}

func (s *Service) onRefreshEvent(ctx context.Context) error {
	s.log.Info("Handling refresh event")

	if err := s.identity.OnRefreshEvent(ctx); err != nil {
		s.log.WithError(err).Error("Failed to handle refresh event")
		return err
	}

	s.eventPublisher.PublishEvent(ctx, events.UserRefreshed{
		UserID:          s.identity.User.ID,
		CancelEventPool: false,
	})

	return nil
}

func (s *Service) onUserSpaceChanged(ctx context.Context, value int) {
	s.log.Info("Handling User Space Changed event")
	if !s.identity.OnUserSpaceChanged(value) {
		return
	}

	s.eventPublisher.PublishEvent(ctx, events.UsedSpaceChanged{
		UserID:    s.identity.User.ID,
		UsedSpace: value,
	})
}

func (s *Service) onAddressEvent(ctx context.Context, addressEvents []proton.AddressEvent) error {
	s.log.Infof("Handling Address Events (%v)", len(addressEvents))

	for idx, event := range addressEvents {
		switch event.Action {
		case proton.EventCreate:
			s.log.WithFields(logrus.Fields{
				"index":     idx,
				"addressID": event.ID,
				"email":     logging.Sensitive(event.Address.Email),
			}).Info("Handling address created event")
			if s.identity.OnAddressCreated(event) == AddressUpdateCreated {
				s.eventPublisher.PublishEvent(ctx, events.UserAddressCreated{
					UserID:    s.identity.User.ID,
					AddressID: event.Address.ID,
					Email:     event.Address.Email,
				})
			}

		case proton.EventUpdate, proton.EventUpdateFlags:
			addr, status := s.identity.OnAddressUpdated(event)
			switch status {
			case AddressUpdateCreated:
				s.eventPublisher.PublishEvent(ctx, events.UserAddressCreated{
					UserID:    s.identity.User.ID,
					AddressID: addr.ID,
					Email:     addr.Email,
				})
			case AddressUpdateUpdated:
				s.eventPublisher.PublishEvent(ctx, events.UserAddressUpdated{
					UserID:    s.identity.User.ID,
					AddressID: addr.ID,
					Email:     addr.Email,
				})

			case AddressUpdateDisabled:
				s.eventPublisher.PublishEvent(ctx, events.UserAddressDisabled{
					UserID:    s.identity.User.ID,
					AddressID: addr.ID,
					Email:     addr.Email,
				})

			case AddressUpdateEnabled:
				s.eventPublisher.PublishEvent(ctx, events.UserAddressEnabled{
					UserID:    s.identity.User.ID,
					AddressID: addr.ID,
					Email:     addr.Email,
				})

			case AddressUpdateNoop:
				continue

			case AddressUpdateDeleted:
				s.log.Warnf("Unexpected address update status after update event %v", status)
				continue
			}

		case proton.EventDelete:
			if addr, status := s.identity.OnAddressDeleted(event); status == AddressUpdateDeleted {
				s.eventPublisher.PublishEvent(ctx, events.UserAddressDeleted{
					UserID:    s.identity.User.ID,
					AddressID: event.ID,
					Email:     addr.Email,
				})
			}
		}
	}
	return nil
}

func sortAddresses(addr []proton.Address) []proton.Address {
	slices.SortFunc(addr, func(a, b proton.Address) bool {
		return a.Order < b.Order
	})

	return addr
}

func buildAddressMapFromSlice(addr []proton.Address) map[string]proton.Address {
	return usertypes.GroupBy(addr, func(addr proton.Address) string { return addr.ID })
}
