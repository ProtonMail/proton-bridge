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

package telemetry

import (
	"context"
	"fmt"

	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/proton-bridge/v3/internal/services/orderedtasks"
	"github.com/ProtonMail/proton-bridge/v3/internal/services/userevents"
	"github.com/ProtonMail/proton-bridge/v3/pkg/cpc"
	"github.com/sirupsen/logrus"
)

type SettingsGetter interface {
	GetUserSettings(ctx context.Context) (proton.UserSettings, error)
}

type Service struct {
	cpc *cpc.CPC
	log *logrus.Entry

	eventService userevents.Subscribable
	subscription *userevents.EventChanneledSubscriber

	userID         string
	settingsGetter SettingsGetter

	isTelemetryEnabled bool
	isInitialised      bool
}

func NewService(
	userID string,
	settingsGetter SettingsGetter,
	eventService userevents.Subscribable,
) *Service {
	return &Service{
		cpc: cpc.NewCPC(),
		log: logrus.WithFields(logrus.Fields{
			"user":    userID,
			"service": "telemetry",
		}),

		eventService: eventService,
		subscription: userevents.NewEventSubscriber(
			fmt.Sprintf("telemetry-%v", userID),
		),

		userID:         userID,
		settingsGetter: settingsGetter,
	}
}

func (s *Service) initialise(ctx context.Context) {
	settings, err := s.settingsGetter.GetUserSettings(ctx)
	if err != nil {
		logrus.WithError(err).Error("Cannot get telemetry settings, asuming off")
		s.isInitialised = false
		s.isTelemetryEnabled = false

		return
	}

	logrus.WithField("telemetry", settings.Telemetry).Debug("Telemetry initialised")
	s.isInitialised = true
	s.isTelemetryEnabled = settings.Telemetry == proton.SettingEnabled
}

func (s *Service) Start(ctx context.Context, group *orderedtasks.OrderedCancelGroup) {
	s.initialise(ctx)

	group.Go(ctx, s.userID, "telemetry-service", s.run)
}

func (s *Service) run(ctx context.Context) {
	s.log.Info("Starting service main loop")
	defer s.log.Info("Exiting service main loop")
	defer s.cpc.Close()

	eventHandler := userevents.EventHandler{
		RefreshHandler:      s,
		UserSettingsHandler: s,
	}

	s.eventService.Subscribe(s.subscription)
	defer s.eventService.Unsubscribe(s.subscription)

	for {
		select {
		case <-ctx.Done():
			return

		case request, ok := <-s.cpc.ReceiveCh():
			if !ok {
				return
			}

			switch request.Value().(type) {
			case *isTelemetryEnabledReq:
				s.log.Debug("Received is telemetry enabled request")
				if !s.isInitialised {
					s.initialise(ctx)
				}

				request.Reply(ctx, s.isTelemetryEnabled, nil)

			default:
				s.log.Error("Received unknown request")
			}
		case e, ok := <-s.subscription.OnEventCh():
			if !ok {
				continue
			}
			e.Consume(func(event proton.Event) error {
				return eventHandler.OnEvent(ctx, event)
			})
		}
	}
}

func (s *Service) HandleRefreshEvent(ctx context.Context, _ proton.RefreshFlag) error {
	s.initialise(ctx)
	return nil
}

func (s *Service) HandleUserSettingsEvent(_ context.Context, settings *proton.UserSettings) error {
	s.isTelemetryEnabled = settings.Telemetry == proton.SettingEnabled
	s.isInitialised = true
	return nil
}

type isTelemetryEnabledReq struct{}

func (s *Service) IsTelemetryEnabled(ctx context.Context) bool {
	enabled, err := cpc.SendTyped[bool](ctx, s.cpc, &isTelemetryEnabledReq{})
	if err != nil {
		s.log.WithError(err).Error("Failed to retrieve IsTelemeteryEnabled, assuming no")
		return false
	}

	return enabled
}
