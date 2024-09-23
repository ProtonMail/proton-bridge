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

package imapservice

import (
	"context"
	"fmt"

	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/proton-bridge/v3/internal/logging"
	obsMetrics "github.com/ProtonMail/proton-bridge/v3/internal/services/imapservice/observabilitymetrics/syncmsgevents"
	"github.com/ProtonMail/proton-bridge/v3/internal/services/observability"
	"github.com/ProtonMail/proton-bridge/v3/internal/services/userevents"
)

func (s *Service) newSyncEventHandler() userevents.EventHandler {
	return userevents.EventHandler{
		RefreshHandler:      s,
		AddressHandler:      s,
		UserHandler:         s,
		LabelHandler:        nil,
		MessageHandler:      &syncMessageEventHandler{service: s},
		UsedSpaceHandler:    nil,
		UserSettingsHandler: nil,
	}
}

type syncMessageEventHandler struct {
	service *Service
}

func (s syncMessageEventHandler) HandleMessageEvents(ctx context.Context, events []proton.MessageEvent) error {
	s.service.log.Debug("handling message events (sync)")
	for _, event := range events {
		//nolint:exhaustive
		switch event.Action {
		case proton.EventCreate:
			updates, err := onMessageCreated(
				logging.WithLogrusField(ctx, "action", "create message (sync)"),
				s.service,
				event.Message,
				true,
			)
			if err != nil {
				s.service.observabilitySender.AddDistinctMetrics(observability.SyncError, obsMetrics.GenerateSyncFailureCreateMessageEventMetric())
				return fmt.Errorf("failed to handle create message event: %w", err)
			}

			if err := waitOnIMAPUpdates(ctx, updates); err != nil {
				return err
			}

		case proton.EventDelete:
			updates := onMessageDeleted(
				logging.WithLogrusField(ctx, "action", "delete message (sync)"),
				s.service,
				event,
			)

			if err := waitOnIMAPUpdates(ctx, updates); err != nil {
				s.service.observabilitySender.AddDistinctMetrics(observability.SyncError, obsMetrics.GenerateSyncFailureDeleteMessageEventMetric())
				return fmt.Errorf("failed to handle delete message event in gluon: %w", err)
			}
		default:
			continue
		}
	}

	return nil
}
