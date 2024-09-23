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
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/proton-bridge/v3/internal/events"
	"github.com/ProtonMail/proton-bridge/v3/internal/logging"
	"github.com/ProtonMail/proton-bridge/v3/internal/services/observability"
	"github.com/ProtonMail/proton-bridge/v3/internal/services/orderedtasks"
	"github.com/ProtonMail/proton-bridge/v3/internal/services/userevents"
	"github.com/ProtonMail/proton-bridge/v3/internal/unleash"
	"github.com/sirupsen/logrus"
)

type Service struct {
	userID string

	log *logrus.Entry

	eventService   userevents.Subscribable
	subscription   *userevents.EventChanneledSubscriber
	eventPublisher events.EventPublisher

	store *Store

	getFlagValueFn unleash.GetFlagValueFn

	observabilitySender observability.Sender
}

const bitfieldRegexPattern = `^\\\d+`
const disableNotificationsKillSwitch = "InboxBridgeEventLoopNotificationDisabled"

func NewService(userID string, service userevents.Subscribable, eventPublisher events.EventPublisher, store *Store,
	getFlagFn unleash.GetFlagValueFn, observabilitySender observability.Sender) *Service {
	return &Service{
		userID: userID,

		log: logrus.WithFields(logrus.Fields{
			"user":    userID,
			"service": "notification",
		}),

		eventService: service,
		subscription: userevents.NewEventSubscriber(
			fmt.Sprintf("notifications-%v", userID)),
		eventPublisher: eventPublisher,

		store: store,

		getFlagValueFn:      getFlagFn,
		observabilitySender: observabilitySender,
	}
}

func (s *Service) Start(ctx context.Context, group *orderedtasks.OrderedCancelGroup) {
	group.Go(ctx, s.userID, "notification-service", s.run)
}

func (s *Service) run(ctx context.Context) {
	s.log.Info("Starting service main loop")
	defer s.log.Info("Exiting service main loop")

	eventHandler := userevents.EventHandler{
		NotificationHandler: s,
	}

	s.eventService.Subscribe(s.subscription)
	defer s.eventService.Unsubscribe(s.subscription)

	for {
		select {
		case <-ctx.Done():
			return
		case e, ok := <-s.subscription.OnEventCh():
			if !ok {
				continue
			}
			e.Consume(func(event proton.Event) error { return eventHandler.OnEvent(ctx, event) })
		}
	}
}

func (s *Service) HandleNotificationEvents(ctx context.Context, notificationEvents []proton.NotificationEvent) error {
	if s.getFlagValueFn(disableNotificationsKillSwitch) {
		s.log.Info("Received notification events. Skipping as kill switch is enabled.")
		return nil
	}

	s.log.Debug("Handling notification events")

	// Publish observability metrics that we've received notifications
	s.observabilitySender.AddMetrics(GenerateReceivedMetric(len(notificationEvents)))

	for _, event := range notificationEvents {
		ctx = logging.WithLogrusField(ctx, "notificationID", event.ID)
		switch strings.ToLower(event.Type) {
		case "bridge_modal":
			{
				// We currently don't support any notification types with bitfields in the body.
				if isBodyBitfieldValid(event) {
					continue
				}

				shouldSend := s.store.shouldSendAndStore(event)
				if !shouldSend {
					s.log.Info("Skipping notification event. Notification was displayed previously")
					continue
				}

				s.log.Info("Publishing notification event. notificationID:", event.ID) // \todo BRIDGE-141 - change this to UID once it is available
				s.eventPublisher.PublishEvent(ctx, events.UserNotification{UserID: s.userID, Title: event.Payload.Title,
					Subtitle: event.Payload.Subtitle, Body: event.Payload.Body})

				// Publish observability metric that we've successfully processed notifications
				s.observabilitySender.AddMetrics(GenerateProcessedMetric(1))
			}

		default:
			s.log.Debug("Skipping notification event. Notification type is not related to bridge:", event.Type)
			continue
		}
	}
	return nil
}

// We will (potentially) encode different notification functionalities based on a starting bitfield "\NUMBER" in the
// payload Body. Currently, we don't support this, but we might in the future. This is so versions of Bridge that don't
// support this functionality won't be display such a message.
func isBodyBitfieldValid(notification proton.NotificationEvent) bool {
	match, err := regexp.MatchString(bitfieldRegexPattern, notification.Payload.Body)
	if err != nil {
		return false
	}
	return match
}
