// Copyright (c) 2025 Proton AG
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

	"github.com/Masterminds/semver/v3"
	"github.com/ProtonMail/gluon/reporter"
	"github.com/ProtonMail/proton-bridge/v3/internal/events"
	"github.com/ProtonMail/proton-bridge/v3/internal/safe"
	"github.com/ProtonMail/proton-bridge/v3/internal/updater"
	"github.com/elastic/go-sysinfo"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
)

func (bridge *Bridge) CheckForUpdates() {
	bridge.goUpdate()
}

func (bridge *Bridge) InstallUpdateLegacy(version updater.VersionInfoLegacy) {
	bridge.installChLegacy <- installJobLegacy{version: version, silent: false}
}

func (bridge *Bridge) InstallUpdate(release updater.Release) {
	bridge.installCh <- installJob{Release: release, Silent: false}
}

func (bridge *Bridge) handleUpdate(version updater.VersionInfo) {
	updateChannel := bridge.vault.GetUpdateChannel()
	updateRollout := bridge.vault.GetUpdateRollout()
	autoUpdateEnabled := bridge.vault.GetAutoUpdate()

	checkSystemVersion := true
	hostInfo, err := sysinfo.Host()
	// If we're unable to get host system information we skip the update's minimum/maximum OS version checks
	if err != nil {
		checkSystemVersion = false
		logrus.WithError(err).Error("Failed to obtain host system info while handling updates")
		if reporterErr := bridge.reporter.ReportMessageWithContext(
			"Failed to obtain host system info while handling updates",
			reporter.Context{"error": err},
		); reporterErr != nil {
			logrus.WithError(reporterErr).Error("Failed to report update error")
		}
	}

	if len(version.Releases) > 0 {
		// Update latest is only used to update the release notes and landing page URL
		bridge.publish(events.UpdateLatest{Release: version.Releases[0]})
	}

	// minAutoUpdateEvent - used to determine the highest compatible update that satisfies the Minimum Bridge version
	minAutoUpdateEvent := events.UpdateAvailable{
		Release:    updater.Release{Version: &semver.Version{}},
		Compatible: false,
		Silent:     false,
	}

	// We assume that the version file is always created in descending order
	// where newer versions are prepended to the top of the releases
	// The logic for checking update eligibility is as follows:
	// 1. Check release channel.
	// 2. Check whether release version is greater.
	// 3. Check if rollout is larger.
	// 4. Check OS Version restrictions (provided that restrictions are provided, and we can extract the OS version).
	// 5. Check Minimum Compatible Bridge Version.
	// 6. Check if an update package is provided.
	// 7. Check auto-update.
	for _, release := range version.Releases {
		log := logrus.WithFields(logrus.Fields{
			"current":               bridge.curVersion,
			"channel":               updateChannel,
			"update_version":        release.Version,
			"update_channel":        release.ReleaseCategory,
			"update_min_auto":       release.MinAuto,
			"update_rollout":        release.RolloutProportion,
			"update_min_os_version": release.SystemVersion.Minimum,
			"update_max_os_version": release.SystemVersion.Maximum,
		})

		log.Debug("Checking update release")

		if !release.ReleaseCategory.UpdateEligible(updateChannel) {
			log.Debug("Update does not satisfy update channel requirement")
			continue
		}

		if !release.Version.GreaterThan(bridge.curVersion) {
			log.Debug("Update version is not greater than current version")
			continue
		}

		if release.RolloutProportion < updateRollout {
			log.Debug("Update has not been rolled out yet")
			continue
		}

		if checkSystemVersion {
			shouldContinue, err := release.SystemVersion.IsHostVersionEligible(log, hostInfo, bridge.getHostVersion)
			if err != nil && shouldContinue {
				log.WithError(err).Error(
					"Failed to verify host system version compatibility during release check." +
						"Error is non-fatal continuing with checks",
				)
			} else if err != nil {
				log.WithError(err).Error("Failed to verify host system version compatibility during update check")
				continue
			}

			if !shouldContinue {
				log.Debug("Host version does not satisfy system requirements for update")
				continue
			}
		}

		if release.MinAuto != nil && bridge.curVersion.LessThan(release.MinAuto) {
			log.Debug("Update is available but is incompatible with this Bridge version")
			if release.Version.GreaterThan(minAutoUpdateEvent.Release.Version) {
				minAutoUpdateEvent.Release = release
			}
			continue
		}

		// Check if we have a provided installer package
		if found := slices.IndexFunc(release.File, func(file updater.File) bool {
			return file.Identifier == updater.PackageIdentifier
		}); found == -1 {
			log.Error("Update is available but does not contain update package")

			if reporterErr := bridge.reporter.ReportMessageWithContext(
				"Available update does not contain update package",
				reporter.Context{"update_version": release.Version},
			); reporterErr != nil {
				log.WithError(reporterErr).Error("Failed to report update error")
			}

			continue
		}

		if !autoUpdateEnabled {
			log.Info("An update is available but auto-update is disabled")
			bridge.publish(events.UpdateAvailable{
				Release:    release,
				Compatible: true,
				Silent:     false,
			})
			return
		}

		// If we've gotten to this point that means an automatic update is available and we should install it
		safe.RLock(func() {
			bridge.installCh <- installJob{Release: release, Silent: true}
		}, bridge.newVersionLock)

		return
	}

	// If there's a release with a minAuto requirement that we satisfy (alongside all other checks)
	// then notify the user that a manual update is needed
	if !minAutoUpdateEvent.Release.Version.Equal(&semver.Version{}) {
		bridge.publish(minAutoUpdateEvent)
	}

	bridge.publish(events.UpdateNotAvailable{})
}

func (bridge *Bridge) handleUpdateLegacy(version updater.VersionInfoLegacy) {
	log := logrus.WithFields(logrus.Fields{
		"version": version.Version,
		"current": bridge.curVersion,
		"channel": bridge.vault.GetUpdateChannel(),
	})

	bridge.publish(events.UpdateLatest{
		VersionLegacy: version,
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
			VersionLegacy: version,
			Compatible:    false,
			Silent:        false,
		})

	case !bridge.vault.GetAutoUpdate():
		log.Info("An update is available but auto-update is disabled")

		bridge.publish(events.UpdateAvailable{
			VersionLegacy: version,
			Compatible:    true,
			Silent:        false,
		})

	default:
		safe.RLock(func() {
			bridge.installChLegacy <- installJobLegacy{version: version, silent: true}
		}, bridge.newVersionLock)
	}
}

type installJobLegacy struct {
	version updater.VersionInfoLegacy
	silent  bool
}

func (bridge *Bridge) installUpdateLegacy(ctx context.Context, job installJobLegacy) {
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
			VersionLegacy: job.version,
			Compatible:    true,
			Silent:        job.silent,
		})

		err := bridge.updater.InstallUpdateLegacy(ctx, bridge.api, job.version)

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
				VersionLegacy: job.version,
				Silent:        job.silent,
				Error:         err,
			})

		default:
			log.Info("The update was installed successfully")

			bridge.publish(events.UpdateInstalled{
				VersionLegacy: job.version,
				Silent:        job.silent,
			})

			bridge.newVersion = job.version.Version
		}
	}, bridge.newVersionLock)
}

type installJob struct {
	Release updater.Release
	Silent  bool
}

func (bridge *Bridge) installUpdate(ctx context.Context, job installJob) {
	safe.Lock(func() {
		log := logrus.WithFields(logrus.Fields{
			"version": job.Release.Version,
			"current": bridge.curVersion,
			"channel": bridge.vault.GetUpdateChannel(),
		})

		if !job.Release.Version.GreaterThan(bridge.newVersion) {
			return
		}

		log.WithField("silent", job.Silent).Info("An update is available")

		bridge.publish(events.UpdateAvailable{
			Release:    job.Release,
			Compatible: true,
			Silent:     job.Silent,
		})

		err := bridge.updater.InstallUpdate(ctx, bridge.api, job.Release)
		switch {
		case errors.Is(err, updater.ErrReleaseUpdatePackageMissing):
			log.WithError(err).Error("The update could not be installed but we will fail silently")
			if reporterErr := bridge.reporter.ReportExceptionWithContext(
				"Cannot download update, update package is missing",
				reporter.Context{"error": err},
			); reporterErr != nil {
				log.WithError(reporterErr).Error("Failed to report update error")
			}
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
				Release: job.Release,
				Silent:  job.Silent,
				Error:   err,
			})

		default:
			log.Info("The update was installed successfully")

			bridge.publish(events.UpdateInstalled{
				Release: job.Release,
				Silent:  job.Silent,
			})

			bridge.newVersion = job.Release.Version
		}
	}, bridge.newVersionLock)
}

func (bridge *Bridge) RemoveOldUpdates() {
	if err := bridge.updater.RemoveOldUpdates(); err != nil {
		logrus.WithError(err).Error("Remove old updates fails")
	}
}
