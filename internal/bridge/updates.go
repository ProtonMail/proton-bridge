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
