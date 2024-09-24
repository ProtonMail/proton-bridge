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
	"errors"

	"github.com/ProtonMail/gluon/reporter"
	"github.com/ProtonMail/proton-bridge/v3/internal/events"
	"github.com/ProtonMail/proton-bridge/v3/internal/safe"
	"github.com/ProtonMail/proton-bridge/v3/internal/updater"
	"github.com/sirupsen/logrus"
)

func (bridge *Bridge) CheckForUpdates() {
	bridge.goUpdate()
}

func (bridge *Bridge) InstallUpdate(version updater.VersionInfo) {
	bridge.installCh <- installJob{version: version, silent: false}
}

func (bridge *Bridge) handleUpdate(version updater.VersionInfo) {
	log := logrus.WithFields(logrus.Fields{
		"version": version.Version,
		"current": bridge.curVersion,
		"channel": bridge.vault.GetUpdateChannel(),
	})

	bridge.publish(events.UpdateLatest{
		Version: version,
	})

	switch {
	case !version.Version.GreaterThan(bridge.curVersion):
		log.Debug("No update available")

		bridge.publish(events.UpdateNotAvailable{})

	case version.RolloutProportion < bridge.vault.GetUpdateRollout():
		log.Info("An update is available but has not been rolled out yet")

		bridge.publish(events.UpdateNotAvailable{})

	case bridge.curVersion.LessThan(version.MinAuto):
		log.Info("An update is available but is incompatible with this version")

		bridge.publish(events.UpdateAvailable{
			Version:    version,
			Compatible: false,
			Silent:     false,
		})

	case !bridge.vault.GetAutoUpdate():
		log.Info("An update is available but auto-update is disabled")

		bridge.publish(events.UpdateAvailable{
			Version:    version,
			Compatible: true,
			Silent:     false,
		})

	default:
		safe.RLock(func() {
			bridge.installCh <- installJob{version: version, silent: true}
		}, bridge.newVersionLock)
	}
}

type installJob struct {
	version updater.VersionInfo
	silent  bool
}

func (bridge *Bridge) installUpdate(ctx context.Context, job installJob) {
	safe.Lock(func() {
		log := logrus.WithFields(logrus.Fields{
			"version": job.version.Version,
			"current": bridge.curVersion,
			"channel": bridge.vault.GetUpdateChannel(),
		})

		if !job.version.Version.GreaterThan(bridge.newVersion) {
			return
		}

		log.WithField("silent", job.silent).Info("An update is available")

		bridge.publish(events.UpdateAvailable{
			Version:    job.version,
			Compatible: true,
			Silent:     job.silent,
		})

		bridge.publish(events.UpdateInstalling{
			Version: job.version,
			Silent:  job.silent,
		})

		err := bridge.updater.InstallUpdate(ctx, bridge.api, job.version)

		switch {
		case errors.Is(err, updater.ErrDownloadVerify):
			// BRIDGE-207: if download or verification fails, we do not want to trigger a manual update. We report in the log and to Sentry
			// and we fail silently.
			log.WithError(err).Error("The update could not be installed, but we will fail silently")
			if reporterErr := bridge.reporter.ReportMessageWithContext(
				"Cannot download or verify update",
				reporter.Context{"error": err},
			); reporterErr != nil {
				log.WithError(reporterErr).Error("Failed to report update error")
			}

		case errors.Is(err, updater.ErrUpdateAlreadyInstalled):
			log.Info("The update was already installed")

		case err != nil:
			log.WithError(err).Error("The update could not be installed")

			bridge.publish(events.UpdateFailed{
				Version: job.version,
				Silent:  job.silent,
				Error:   err,
			})

		default:
			log.Info("The update was installed successfully")

			bridge.publish(events.UpdateInstalled{
				Version: job.version,
				Silent:  job.silent,
			})

			bridge.newVersion = job.version.Version
		}
	}, bridge.newVersionLock)
}

func (bridge *Bridge) RemoveOldUpdates() {
	if err := bridge.updater.RemoveOldUpdates(); err != nil {
		logrus.WithError(err).Error("Remove old updates fails")
	}
}
