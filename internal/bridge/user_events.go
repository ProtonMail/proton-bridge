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

package bridge

import (
	"context"

	"github.com/ProtonMail/gluon/reporter"
	"github.com/ProtonMail/proton-bridge/v3/internal"
	"github.com/ProtonMail/proton-bridge/v3/internal/events"
	"github.com/ProtonMail/proton-bridge/v3/internal/safe"
	"github.com/ProtonMail/proton-bridge/v3/internal/user"
	"github.com/sirupsen/logrus"
)

func (bridge *Bridge) handleUserEvent(ctx context.Context, user *user.User, event events.Event) {
	switch event := event.(type) {
	case events.UserDeauth:
		bridge.handleUserDeauth(ctx, user)

	case events.UserBadEvent:
		bridge.handleUserBadEvent(ctx, user, event)

	case events.UserLoadedCheckResync:
		user.VerifyResyncAndExecute()

	case events.UncategorizedEventError:
		bridge.handleUncategorizedErrorEvent(event)
	}
}

func (bridge *Bridge) handleUserDeauth(ctx context.Context, user *user.User) {
	safe.Lock(func() {
		bridge.logoutUser(ctx, user, false, false, false)
		user.ReportConfigStatusFailure("User deauth.")
	}, bridge.usersLock)
}

func (bridge *Bridge) handleUserBadEvent(ctx context.Context, user *user.User, event events.UserBadEvent) {
	safe.RLock(func() {
		if rerr := bridge.reporter.ReportMessageWithContext("Failed to handle event", reporter.Context{
			"user_id":      user.ID(),
			"old_event_id": event.OldEventID,
			"new_event_id": event.NewEventID,
			"event_info":   event.EventInfo,
			"error":        event.Error,
			"error_type":   internal.ErrCauseType(event.Error),
		}); rerr != nil {
			logrus.WithField("pkg", "bridge/event").WithError(rerr).Error("Failed to report failed event handling")
		}

		user.OnBadEvent(ctx)
	}, bridge.usersLock)
}

func (bridge *Bridge) handleUncategorizedErrorEvent(event events.UncategorizedEventError) {
	if rerr := bridge.reporter.ReportMessageWithContext("Failed to handle due to uncategorized error", reporter.Context{
		"error_type": internal.ErrCauseType(event.Error),
		"error":      event.Error,
	}); rerr != nil {
		logrus.WithField("pkg", "bridge/event").WithError(rerr).Error("Failed to report failed event handling")
	}
}
