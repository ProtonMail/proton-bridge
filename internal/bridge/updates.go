// Copyright (c) 2022 Proton AG
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
	"time"

	"github.com/ProtonMail/proton-bridge/v2/internal/constants"
	"github.com/ProtonMail/proton-bridge/v2/internal/events"
	"github.com/ProtonMail/proton-bridge/v2/internal/updater"
)

func (bridge *Bridge) CheckForUpdates() {
	bridge.updateCheckCh <- struct{}{}
}

func (bridge *Bridge) watchForUpdates() error {
	if _, err := bridge.updater.GetVersionInfo(bridge.api, bridge.vault.GetUpdateChannel()); err != nil {
		return err
	}

	ticker := time.NewTicker(constants.UpdateCheckInterval)

	go func() {
		for {
			select {
			case <-bridge.stopCh:
				return

			case <-bridge.updateCheckCh:
				// ...

			case <-ticker.C:
				// ...
			}

			version, err := bridge.updater.GetVersionInfo(bridge.api, bridge.vault.GetUpdateChannel())
			if err != nil {
				continue
			}

			if err := bridge.handleUpdate(version); err != nil {
				continue
			}
		}
	}()

	bridge.updateCheckCh <- struct{}{}

	return nil
}

func (bridge *Bridge) handleUpdate(version updater.VersionInfo) error {
	switch {
	case !version.Version.GreaterThan(bridge.curVersion):
		bridge.publish(events.UpdateNotAvailable{})

	case version.RolloutProportion < bridge.vault.GetUpdateRollout():
		bridge.publish(events.UpdateNotAvailable{})

	case bridge.curVersion.LessThan(version.MinAuto):
		bridge.publish(events.UpdateAvailable{
			Version:    version,
			CanInstall: false,
		})

	case !bridge.vault.GetAutoUpdate():
		bridge.publish(events.UpdateAvailable{
			Version:    version,
			CanInstall: true,
		})

	default:
		if err := bridge.updater.InstallUpdate(bridge.api, version); err != nil {
			return err
		}

		bridge.publish(events.UpdateInstalled{
			Version: version,
		})
	}

	return nil
}
