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

package useridentity

import (
	"context"
	"fmt"

	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/proton-bridge/v3/internal/events"
	"github.com/ProtonMail/proton-bridge/v3/internal/logging"
	"github.com/ProtonMail/proton-bridge/v3/internal/services/orderedtasks"
	"github.com/ProtonMail/proton-bridge/v3/internal/services/userevents"
	"github.com/ProtonMail/proton-bridge/v3/internal/usertypes"
	"github.com/ProtonMail/proton-bridge/v3/pkg/cpc"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

type IdentityProvider interface {
	GetUser(ctx context.Context) (proton.User, error)
	GetAddresses(ctx context.Context) ([]proton.Address, error)
}

// Service contains all the data required to establish the user identity. This
// includes all the user's information as well as mail addresses and keys.
type Service struct {
	cpc            *cpc.CPC
	eventService   userevents.Subscribable
	eventPublisher events.EventPublisher
	log            *logrus.Entry
	identity       State

	subscription *userevents.EventChanneledSubscriber

	bridgePassProvider BridgePassProvider
	telemetry          Telemetry
}

func NewService(
	service userevents.Subscribable,
	eventPublisher events.EventPublisher,
	state *State,
	bridgePassProvider BridgePassProvider,
	telemetry Telemetry,
) *Service {
	subscriberName := fmt.Sprintf("identity-%v", state.User.ID)

	return &Service{
		cpc:            cpc.NewCPC(),
		eventService:   service,
		identity:       *state,
		eventPublisher: eventPublisher,
		log: logrus.WithFields(logrus.Fields{
			"service": "user-identity",
			"user":    state.User.ID,
		}),
		subscription:       userevents.NewEventSubscriber(subscriberName),
		bridgePassProvider: bridgePassProvider,
		telemetry:          telemetry,
	}
}

func (s *Service) Start(ctx context.Context, group *orderedtasks.OrderedCancelGroup) {
	group.Go(ctx, s.identity.User.ID, "identity-service", s.run)
}

func (s *Service) Resync(ctx context.Context) error {
	_, err := s.cpc.Send(ctx, &resyncReq{})

	return err
}

func (s *Service) GetAPIUser(ctx context.Context) (proton.User, error) {
	return cpc.SendTyped[proton.User](ctx, s.cpc, &getUserReq{})
}

func (s *Service) GetAddresses(ctx context.Context) (map[string]proton.Address, error) {
	return cpc.SendTyped[map[string]proton.Address](ctx, s.cpc, &getAddressesReq{})
}

func (s *Service) CheckAuth(ctx context.Context, email string, password []byte) (string, error) {
	return cpc.SendTyped[string](ctx, s.cpc, &checkAuthReq{
		email:    email,
		password: password,
	})
}

func (s *Service) HandleUsedSpaceEvent(ctx context.Context, newSpace int64) error {
	s.log.Info("Handling User Space Changed event")

	if s.identity.OnUserSpaceChanged(uint64(newSpace)) { //nolint:gosec // disable G115
		s.eventPublisher.PublishEvent(ctx, events.UsedSpaceChanged{
			UserID:    s.identity.User.ID,
			UsedSpace: uint64(newSpace), //nolint:gosec // disable G115
		})
	}

	return nil
}

func (s *Service) HandleUserEvent(ctx context.Context, user *proton.User) error {
	s.log.WithField("username", logging.Sensitive(user.Name)).Info("Handling user event")
	s.identity.OnUserEvent(*user)
	s.eventPublisher.PublishEvent(ctx, events.UserChanged{
		UserID: user.ID,
	})

	return nil
}

func (s *Service) HandleAddressEvents(ctx context.Context, addressEvents []proton.AddressEvent) error {
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

func (s *Service) HandleRefreshEvent(ctx context.Context, _ proton.RefreshFlag) error {
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

func (s *Service) run(ctx context.Context) {
	s.log.WithFields(logrus.Fields{
		"numAddr": len(s.identity.Addresses),
	}).Info("Starting user identity service")
	defer s.log.Info("Exiting Service")

	eventHandler := userevents.EventHandler{
		UserHandler:      s,
		AddressHandler:   s,
		UsedSpaceHandler: s,
		RefreshHandler:   s,
	}

	s.registerSubscription()
	defer s.unregisterSubscription()

	defer s.cpc.Close()

	for {
		select {
		case <-ctx.Done():
			return
		case r, ok := <-s.cpc.ReceiveCh():
			if !ok {
				continue
			}
			switch req := r.Value().(type) {
			case *resyncReq:
				err := s.identity.OnRefreshEvent(ctx)
				r.Reply(ctx, nil, err)

			case *getUserReq:
				r.Reply(ctx, s.identity.User, nil)

			case *getAddressesReq:
				r.Reply(ctx, maps.Clone(s.identity.Addresses), nil)

			case *checkAuthReq:
				id, err := s.identity.CheckAuth(req.email, req.password, s.bridgePassProvider)
				r.Reply(ctx, id, err)

			default:
				s.log.Error("Invalid request")
			}

		case evt, ok := <-s.subscription.OnEventCh():
			if !ok {
				continue
			}
			evt.Consume(func(event proton.Event) error {
				return eventHandler.OnEvent(ctx, event)
			})
		}
	}
}

func (s *Service) registerSubscription() {
	s.eventService.Subscribe(s.subscription)
}

func (s *Service) unregisterSubscription() {
	s.eventService.Unsubscribe(s.subscription)
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

type resyncReq struct{}

type getUserReq struct{}

type checkAuthReq struct {
	email    string
	password []byte
}

type getAddressesReq struct{}
