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
	timer          *proton.Ticker
	eventTimeout   time.Duration
	paused         uint32
	panicHandler   async.PanicHandler

	subscriberList eventSubscriberList

	pendingSubscriptionsLock sync.Mutex
	pendingSubscriptions     []pendingSubscription
}

func NewService(
	userID string,
	eventSource EventSource,
	store EventIDStore,
	eventPublisher events.EventPublisher,
	pollPeriod time.Duration,
	jitter time.Duration,
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
		timer:          proton.NewTicker(pollPeriod, jitter, panicHandler),
		paused:         1,
		eventTimeout:   eventTimeout,
		panicHandler:   panicHandler,
	}
}

// Subscribe adds new subscribers to the service.
// This method can safely be called during event handling.
func (s *Service) Subscribe(subscription EventSubscriber) {
	s.pendingSubscriptionsLock.Lock()
	defer s.pendingSubscriptionsLock.Unlock()

	s.pendingSubscriptions = append(s.pendingSubscriptions, pendingSubscription{op: pendingOpAdd, sub: subscription})
}

// Unsubscribe removes subscribers from the service.
// This method can safely be called during event handling.
func (s *Service) Unsubscribe(subscription EventSubscriber) {
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
	defer s.timer.Stop()
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

	processed := xmaps.Set[EventSubscriber]{}

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
		s.log.Info("Received refresh event")
	}

	return s.subscriberList.PublishParallel(ctx, event, s.panicHandler, s.eventTimeout)
}

func unpackPublisherError(err error) (string, error) {
	var publishErr *eventPublishError

	if errors.As(err, &publishErr) {
		return publishErr.subscriber.name(), publishErr.error
	}

	return "", err
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
	if apiErr := new(proton.APIError); errors.As(err, &apiErr) && (apiErr.Status == 429 || apiErr.Status >= 500) {
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

func (s *Service) addSubscription(subscription EventSubscriber) {
	s.subscriberList.Add(subscription)
}

func (s *Service) removeSubscription(subscription EventSubscriber) {
	s.subscriberList.Remove(subscription)
}

type pendingOp int

const (
	pendingOpAdd pendingOp = iota
	pendingOpRemove
)

type pendingSubscription struct {
	op  pendingOp
	sub EventSubscriber
}
