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

package bridge

import (
	"strings"

	"github.com/ProtonMail/gluon/reporter"
	"github.com/ProtonMail/proton-bridge/v3/internal/vault"
	"github.com/sirupsen/logrus"
)

func (bridge *Bridge) databaseResyncNeeded() bool {
	if strings.HasPrefix(bridge.lastVersion.String(), "3.4.0") &&
		strings.HasPrefix(bridge.curVersion.String(), "3.4.1") {
		logrus.WithFields(logrus.Fields{
			"lastVersion": bridge.lastVersion.String(),
			"currVersion": bridge.curVersion.String(),
		}).Warning("Database re-synchronisation needed")
		return true
	}
	return false
}

func (bridge *Bridge) TryMigrate(vault *vault.User) {
	if bridge.databaseResyncNeeded() {
		if err := bridge.reporter.ReportMessage("Database need to be re-sync for migration."); err != nil {
			logrus.WithError(err).Error("Failed to report database re-sync for migration.")
		}
		if err := vault.ClearSyncStatus(); err != nil {
			logrus.WithError(err).Error("Failed reset to SyncStatus.")
			if err2 := bridge.reporter.ReportMessageWithContext("Failed to reset SyncStatus for Database migration.",
				reporter.Context{
					"error": err,
				}); err2 != nil {
				logrus.WithError(err2).Error("Failed to report reset SyncStatus error.")
			}
		}
	}
}
