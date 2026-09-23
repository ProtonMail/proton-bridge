// Copyright (c) 2026 Proton AG
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
	"fmt"

	"github.com/ProtonMail/proton-bridge/v3/internal/services/imapservice"
	"github.com/ProtonMail/proton-bridge/v3/internal/vault"
)

// migrateSyncStatusFromVaultToSyncStateFile - The sync status used to be in the vault; migrate the data over
// to the dedicated sync file.
func migrateSyncStatusFromVaultToSyncStateFile(encVault *vault.User, syncConfigDir string, userID string) error {
	syncStatus := encVault.SyncStatus()

	migrated, err := imapservice.MigrateVaultSettings(syncConfigDir, userID, syncStatus.HasLabels, syncStatus.HasMessages, syncStatus.FailedMessageIDs)
	if err != nil {
		return fmt.Errorf("failed to migrate user sync settings: %w", err)
	}

	if migrated {
		if err := encVault.ClearSyncStatusWithoutEventID(); err != nil {
			return fmt.Errorf("failed to clear sync settings from vault: %w", err)
		}
	}

	return nil
}

// cleanupStaleSyncStateFile a new user may indicate that the vault was wiped; in such a case we should
// clean up the old sync state file.
func cleanupStaleSyncStateFile(isNew bool, syncConfigDir string, userID string) error {
	if !isNew {
		return nil
	}

	if err := imapservice.DeleteSyncState(syncConfigDir, userID); err != nil {
		return fmt.Errorf("failed to clear outdated sync state file: %w", err)
	}
	return nil
}
