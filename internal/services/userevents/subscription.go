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

package userevents

import (
	"context"
	"fmt"

	"github.com/ProtonMail/go-proton-api"
)

type Subscription = EventSubscriber

func NewEventSubscriber(name string) *EventChanneledSubscriber {
	return newSubscriber(name)
}

type EventHandler struct {
	RefreshHandler      RefreshEventHandler
	AddressHandler      AddressEventHandler
	UserHandler         UserEventHandler
	LabelHandler        LabelEventHandler
	MessageHandler      MessageEventHandler
	UsedSpaceHandler    UserUsedSpaceEventHandler
	UserSettingsHandler UserSettingsHandler
	NotificationHandler NotificationEventHandler
}

func (e EventHandler) OnEvent(ctx context.Context, event proton.Event) error {
	if event.Refresh&proton.RefreshMail != 0 && e.RefreshHandler != nil {
		return e.RefreshHandler.HandleRefreshEvent(ctx, event.Refresh)
	}

	// Start with user settings because of telemetry.
	if event.UserSettings != nil && e.UserSettingsHandler != nil {
		if err := e.UserSettingsHandler.HandleUserSettingsEvent(ctx, event.UserSettings); err != nil {
			return fmt.Errorf("failed to apply user event: %w", err)
		}
	}

	// Continue with user events.
	if event.User != nil && e.UserHandler != nil {
		if err := e.UserHandler.HandleUserEvent(ctx, event.User); err != nil {
			return fmt.Errorf("failed to apply user event: %w", err)
		}
	}

	// Next Address events
	if len(event.Addresses) != 0 && e.AddressHandler != nil {
		if err := e.AddressHandler.HandleAddressEvents(ctx, event.Addresses); err != nil {
			return fmt.Errorf("failed to apply address events: %w", err)
		}
	}

	// Next label events
	if len(event.Labels) != 0 && e.LabelHandler != nil {
		if err := e.LabelHandler.HandleLabelEvents(ctx, event.Labels); err != nil {
			return fmt.Errorf("failed to apply label events: %w", err)
		}
	}

	// Next message events
	if len(event.Messages) != 0 && e.MessageHandler != nil {
		if err := e.MessageHandler.HandleMessageEvents(ctx, event.Messages); err != nil {
			return fmt.Errorf("failed to apply message events: %w", err)
		}
	}

	// Finally user used space events
	if event.UsedSpace != nil && e.UsedSpaceHandler != nil {
		if err := e.UsedSpaceHandler.HandleUsedSpaceEvent(ctx, *event.UsedSpace); err != nil {
			return fmt.Errorf("failed to apply message events: %w", err)
		}
	}

	if len(event.Notifications) != 0 && e.NotificationHandler != nil {
		if err := e.NotificationHandler.HandleNotificationEvents(ctx, event.Notifications); err != nil {
			return fmt.Errorf("failed to apply notification events: %w", err)
		}
	}

	return nil
}

type RefreshEventHandler interface {
	HandleRefreshEvent(ctx context.Context, flag proton.RefreshFlag) error
}
type UserEventHandler interface {
	HandleUserEvent(ctx context.Context, user *proton.User) error
}

type UserUsedSpaceEventHandler interface {
	HandleUsedSpaceEvent(ctx context.Context, newSpace int64) error
}

type UserSettingsHandler interface {
	HandleUserSettingsEvent(ctx context.Context, settings *proton.UserSettings) error
}

type AddressEventHandler interface {
	HandleAddressEvents(ctx context.Context, events []proton.AddressEvent) error
}

type LabelEventHandler interface {
	HandleLabelEvents(ctx context.Context, events []proton.LabelEvent) error
}

type MessageEventHandler interface {
	HandleMessageEvents(ctx context.Context, events []proton.MessageEvent) error
}

type NotificationEventHandler interface {
	HandleNotificationEvents(ctx context.Context, events []proton.NotificationEvent) error
}
