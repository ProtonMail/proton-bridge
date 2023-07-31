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
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ProtonMail/gluon/async"
	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/proton-bridge/v3/internal"
	"github.com/ProtonMail/proton-bridge/v3/internal/events"
	"github.com/ProtonMail/proton-bridge/v3/internal/services/orderedtasks"
	"github.com/bradenaw/juniper/xmaps"
	"github.com/sirupsen/logrus"
)

// Service polls from the given event source and ensures that all the respective subscribers get notified
// before proceeding to the next event. The events are published in the following order:
// * Refresh
// * User
// * Address
// * Label
// * Message
// * UserUsedSpace
// By default this service starts paused, you need to call `Service.Resume` at least one time to begin event polling.
type Service struct {
	userID         string
	eventSource    EventSource
	eventIDStore   EventIDStore
	log            *logrus.Entry
	eventPublisher events.EventPublisher
	timer          *time.Ticker
	eventTimeout   time.Duration
	paused         uint32
	panicHandler   async.PanicHandler

	userSubscriberList      userSubscriberList
	addressSubscribers      addressSubscriberList
	labelSubscribers        labelSubscriberList
	messageSubscribers      messageSubscriberList
	refreshSubscribers      refreshSubscriberList
	userUsedSpaceSubscriber userUsedSpaceSubscriberList

	pendingSubscriptionsLock sync.Mutex
	pendingSubscriptions     []pendingSubscription
}

func NewService(
	userID string,
	eventSource EventSource,
	store EventIDStore,
	eventPublisher events.EventPublisher,
	pollPeriod time.Duration,
	eventTimeout time.Duration,
	panicHandler async.PanicHandler,
) *Service {
	return &Service{
		userID:       userID,
		eventSource:  eventSource,
		eventIDStore: store,
		log: logrus.WithFields(logrus.Fields{
			"service": "user-events",
			"user":    userID,
		}),
		eventPublisher: eventPublisher,
		timer:          time.NewTicker(pollPeriod),
		paused:         1,
		eventTimeout:   eventTimeout,
		panicHandler:   panicHandler,
	}
}

type Subscription struct {
	User          UserSubscriber
	Refresh       RefreshSubscriber
	Address       AddressSubscriber
	Labels        LabelSubscriber
	Messages      MessageSubscriber
	UserUsedSpace UserUsedSpaceSubscriber
}

// cancel subscription subscribers if applicable, see `subscriber.cancel` for more information.
func (s Subscription) cancel() {
	if s.User != nil {
		s.User.cancel()
	}
	if s.Refresh != nil {
		s.Refresh.cancel()
	}
	if s.Address != nil {
		s.Address.cancel()
	}
	if s.Labels != nil {
		s.Labels.cancel()
	}
	if s.Messages != nil {
		s.Messages.cancel()
	}
	if s.UserUsedSpace != nil {
		s.UserUsedSpace.cancel()
	}
}

func (s Subscription) close() {
	if s.User != nil {
		s.User.close()
	}
	if s.Refresh != nil {
		s.Refresh.close()
	}
	if s.Address != nil {
		s.Address.close()
	}
	if s.Labels != nil {
		s.Labels.close()
	}
	if s.Messages != nil {
		s.Messages.close()
	}
	if s.UserUsedSpace != nil {
		s.UserUsedSpace.close()
	}
}

// Subscribe adds new subscribers to the service.
// This method can safely be called during event handling.
func (s *Service) Subscribe(subscription Subscription) {
	s.pendingSubscriptionsLock.Lock()
	defer s.pendingSubscriptionsLock.Unlock()

	s.pendingSubscriptions = append(s.pendingSubscriptions, pendingSubscription{op: pendingOpAdd, sub: subscription})
}

// Unsubscribe removes subscribers from the service.
// This method can safely be called during event handling.
func (s *Service) Unsubscribe(subscription Subscription) {
	subscription.cancel()

	s.pendingSubscriptionsLock.Lock()
	defer s.pendingSubscriptionsLock.Unlock()

	s.pendingSubscriptions = append(s.pendingSubscriptions, pendingSubscription{op: pendingOpRemove, sub: subscription})
}

// Pause pauses the event polling.
func (s *Service) Pause() {
	atomic.StoreUint32(&s.paused, 1)
}

// Resume resumes the event polling.
func (s *Service) Resume() {
	atomic.StoreUint32(&s.paused, 0)
}

// IsPaused return true if the service is paused.
func (s *Service) IsPaused() bool {
	return atomic.LoadUint32(&s.paused) == 1
}

func (s *Service) Start(ctx context.Context, group *orderedtasks.OrderedCancelGroup) error {
	lastEventID, err := s.eventIDStore.Load(ctx)
	if err != nil {
		return fmt.Errorf("failed to load last event id: %w", err)
	}

	if lastEventID == "" {
		s.log.Debugf("No event ID present in storage, retrieving latest")
		eventID, err := s.eventSource.GetLatestEventID(ctx)
		if err != nil {
			return fmt.Errorf("failed to get latest event id: %w", err)
		}

		if err := s.eventIDStore.Store(ctx, eventID); err != nil {
			return fmt.Errorf("failed to store event in event id store: %v", err)
		}

		lastEventID = eventID
	}

	group.Go(ctx, s.userID, "event-service", func(ctx context.Context) {
		s.run(ctx, lastEventID)
	})

	return nil
}

func (s *Service) run(ctx context.Context, lastEventID string) {
	s.log.Infof("Starting service Last EventID=%v", lastEventID)
	defer s.close()
	defer s.log.Info("Exiting service")
	defer s.Close()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.timer.C:
			if s.IsPaused() {
				continue
			}
		}

		// Apply any pending subscription changes.
		func() {
			s.pendingSubscriptionsLock.Lock()
			defer s.pendingSubscriptionsLock.Unlock()

			for _, p := range s.pendingSubscriptions {
				if p.op == pendingOpAdd {
					s.addSubscription(p.sub)
				} else {
					s.removeSubscription(p.sub)
				}
			}

			s.pendingSubscriptions = nil
		}()

		newEvents, _, err := s.eventSource.GetEvent(ctx, lastEventID)
		if err != nil {
			s.log.WithError(err).Errorf("Failed to get event (caused by %T)", internal.ErrCause(err))
			continue
		}

		// If the event ID hasn't changed, there are no new events.
		if newEvents[len(newEvents)-1].EventID == lastEventID {
			s.log.Debugf("No new API Events")
			continue
		}

		if event, eventErr := func() (proton.Event, error) {
			for _, event := range newEvents {
				if err := s.handleEvent(ctx, lastEventID, event); err != nil {
					return event, err
				}
			}

			return proton.Event{}, nil
		}(); eventErr != nil {
			subscriberName, err := s.handleEventError(ctx, lastEventID, event, eventErr)
			if subscriberName == "" {
				subscriberName = "?"
			}
			s.log.WithField("subscriber", subscriberName).WithError(err).Errorf("Failed to apply event")
			continue
		}

		newEventID := newEvents[len(newEvents)-1].EventID
		if err := s.eventIDStore.Store(ctx, newEventID); err != nil {
			s.log.WithError(err).Errorf("Failed to store new event ID: %v", err)
			s.onBadEvent(ctx, events.UserBadEvent{
				Error:  fmt.Errorf("failed to store new event ID: %w", err),
				UserID: s.userID,
			})
			continue
		}

		lastEventID = newEventID
	}
}

// Close should be called after the service has been cancelled to clean up any remaining pending operations.
func (s *Service) Close() {
	s.pendingSubscriptionsLock.Lock()
	defer s.pendingSubscriptionsLock.Unlock()

	processed := xmaps.Set[Subscription]{}

	// Cleanup pending removes.
	for _, s := range s.pendingSubscriptions {
		if s.op == pendingOpRemove {
			if !processed.Contains(s.sub) {
				s.sub.close()
			}
		} else {
			s.sub.cancel()
			s.sub.close()
			processed.Add(s.sub)
		}
	}

	s.pendingSubscriptions = nil
}

func (s *Service) handleEvent(ctx context.Context, lastEventID string, event proton.Event) error {
	s.log.WithFields(logrus.Fields{
		"old": lastEventID,
		"new": event,
	}).Info("Received new API event")

	if event.Refresh&proton.RefreshMail != 0 {
		s.log.Info("Handling refresh event")
		if err := s.refreshSubscribers.Publish(ctx, event.Refresh, s.eventTimeout); err != nil {
			return fmt.Errorf("failed to apply refresh event: %w", err)
		}

		return nil
	}

	// Start with user events.
	if event.User != nil {
		if err := s.userSubscriberList.PublishParallel(ctx, *event.User, s.panicHandler, s.eventTimeout); err != nil {
			return fmt.Errorf("failed to apply user event: %w", err)
		}
	}

	// Next Address events
	if err := s.addressSubscribers.PublishParallel(ctx, event.Addresses, s.panicHandler, s.eventTimeout); err != nil {
		return fmt.Errorf("failed to apply address events: %w", err)
	}

	// Next label events
	if err := s.labelSubscribers.PublishParallel(ctx, event.Labels, s.panicHandler, s.eventTimeout); err != nil {
		return fmt.Errorf("failed to apply label events: %w", err)
	}

	// Next message events
	if err := s.messageSubscribers.PublishParallel(ctx, event.Messages, s.panicHandler, s.eventTimeout); err != nil {
		return fmt.Errorf("failed to apply message events: %w", err)
	}

	// Finally user used space events
	if event.UsedSpace != nil {
		if err := s.userUsedSpaceSubscriber.PublishParallel(ctx, *event.UsedSpace, s.panicHandler, s.eventTimeout); err != nil {
			return fmt.Errorf("failed to apply message events: %w", err)
		}
	}

	return nil
}

func unpackPublisherError(err error) (string, error) {
	var addressErr *addressPublishError
	var labelErr *labelPublishError
	var messageErr *messagePublishError
	var refreshErr *refreshPublishError
	var userErr *userPublishError
	var usedSpaceErr *userUsedEventPublishError

	switch {
	case errors.As(err, &userErr):
		return userErr.subscriber.name(), userErr.error
	case errors.As(err, &addressErr):
		return addressErr.subscriber.name(), addressErr.error
	case errors.As(err, &labelErr):
		return labelErr.subscriber.name(), labelErr.error
	case errors.As(err, &messageErr):
		return messageErr.subscriber.name(), messageErr.error
	case errors.As(err, &refreshErr):
		return refreshErr.subscriber.name(), refreshErr.error
	case errors.As(err, &usedSpaceErr):
		return usedSpaceErr.subscriber.name(), usedSpaceErr.error
	default:
		return "", err
	}
}

func (s *Service) handleEventError(ctx context.Context, lastEventID string, event proton.Event, err error) (string, error) {
	// Unpack the error so we can proceed to handle the real issue.
	subscriberName, err := unpackPublisherError(err)

	// If the error is a context cancellation, return error to retry later.
	if errors.Is(err, context.Canceled) {
		return subscriberName, fmt.Errorf("failed to handle event due to context cancellation: %w", err)
	}

	// If the error is a network error, return error to retry later.
	if netErr := new(proton.NetError); errors.As(err, &netErr) {
		return subscriberName, fmt.Errorf("failed to handle event due to network issue: %w", err)
	}

	// Catch all for uncategorized net errors that may slip through.
	if netErr := new(net.OpError); errors.As(err, &netErr) {
		return subscriberName, fmt.Errorf("failed to handle event due to network issues (uncategorized): %w", err)
	}

	// In case a json decode error slips through.
	if jsonErr := new(json.UnmarshalTypeError); errors.As(err, &jsonErr) {
		s.eventPublisher.PublishEvent(ctx, events.UncategorizedEventError{
			UserID: s.userID,
			Error:  err,
		})

		return subscriberName, fmt.Errorf("failed to handle event due to JSON issue: %w", err)
	}

	// If the error is an unexpected EOF, return error to retry later.
	if errors.Is(err, io.ErrUnexpectedEOF) {
		return subscriberName, fmt.Errorf("failed to handle event due to EOF: %w", err)
	}

	// If the error is a server-side issue, return error to retry later.
	if apiErr := new(proton.APIError); errors.As(err, &apiErr) && apiErr.Status >= 500 {
		return subscriberName, fmt.Errorf("failed to handle event due to server error: %w", err)
	}

	// Otherwise, the error is a client-side issue; notify bridge to handle it.
	s.log.WithField("event", event).Warn("Failed to handle API event")

	s.onBadEvent(ctx, events.UserBadEvent{
		UserID:     s.userID,
		OldEventID: lastEventID,
		NewEventID: event.EventID,
		EventInfo:  event.String(),
		Error:      err,
	})

	return subscriberName, fmt.Errorf("failed to handle event due to client error: %w", err)
}

func (s *Service) onBadEvent(ctx context.Context, event events.UserBadEvent) {
	s.Pause()
	s.eventPublisher.PublishEvent(ctx, event)
}

func (s *Service) addSubscription(subscription Subscription) {
	if subscription.User != nil {
		s.userSubscriberList.Add(subscription.User)
	}

	if subscription.Refresh != nil {
		s.refreshSubscribers.Add(subscription.Refresh)
	}

	if subscription.Address != nil {
		s.addressSubscribers.Add(subscription.Address)
	}

	if subscription.Labels != nil {
		s.labelSubscribers.Add(subscription.Labels)
	}

	if subscription.Messages != nil {
		s.messageSubscribers.Add(subscription.Messages)
	}

	if subscription.UserUsedSpace != nil {
		s.userUsedSpaceSubscriber.Add(subscription.UserUsedSpace)
	}
}

func (s *Service) removeSubscription(subscription Subscription) {
	if subscription.User != nil {
		s.userSubscriberList.Remove(subscription.User)
	}

	if subscription.Refresh != nil {
		s.refreshSubscribers.Remove(subscription.Refresh)
	}

	if subscription.Address != nil {
		s.addressSubscribers.Remove(subscription.Address)
	}

	if subscription.Labels != nil {
		s.labelSubscribers.Remove(subscription.Labels)
	}

	if subscription.Messages != nil {
		s.messageSubscribers.Remove(subscription.Messages)
	}

	if subscription.UserUsedSpace != nil {
		s.userUsedSpaceSubscriber.Remove(subscription.UserUsedSpace)
	}
}

func (s *Service) close() {
	s.timer.Stop()
}

type pendingOp int

const (
	pendingOpAdd pendingOp = iota
	pendingOpRemove
)

type pendingSubscription struct {
	op  pendingOp
	sub Subscription
}
