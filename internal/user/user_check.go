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

package user

import (
	"context"
	"fmt"

	"github.com/ProtonMail/proton-bridge/v3/internal/events"
	"github.com/ProtonMail/proton-bridge/v3/internal/services/imapservice"
)

func checkIrrecoverableEventID(
	ctx context.Context,
	lastEventID,
	userID,
	syncConfigDir string,
	publisher events.EventPublisher,
) error {
	// If we detect that the event ID stored in the vault got reset, the user is not a new account and
	// we have started or finished syncing: this is an irrecoverable state and we should produce a bad event.
	if lastEventID != "" {
		return nil
	}

	syncConfigPath := imapservice.GetSyncConfigPath(syncConfigDir, userID)

	syncState, err := imapservice.NewSyncState(syncConfigPath)
	if err != nil {
		return fmt.Errorf("failed to read imap sync state: %w", err)
	}

	syncStatus, err := syncState.GetSyncStatus(ctx)
	if err != nil {
		return fmt.Errorf("failed to imap sync status: %w", err)
	}

	if syncStatus.IsComplete() || syncStatus.InProgress() {
		publisher.PublishEvent(ctx, newEmptyEventIDBadEvent(userID))
	}

	return nil
}

func newEmptyEventIDBadEvent(userID string) events.UserBadEvent {
	return events.UserBadEvent{
		UserID:     userID,
		OldEventID: "",
		NewEventID: "",
		EventInfo:  "EventID missing from vault",
		Error:      fmt.Errorf("eventID in vault is empty, when it shouldn't be"),
	}
}
