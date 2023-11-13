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
	"fmt"

	"github.com/ProtonMail/proton-bridge/v3/internal/services/imapservice"
	"github.com/ProtonMail/proton-bridge/v3/internal/vault"
)

func migrateSyncStatusFromVault(encVault *vault.User, syncConfigDir string, userID string) error {
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
