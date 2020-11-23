// Copyright (c) 2020 Proton Technologies AG
//
// This file is part of ProtonMail Bridge.
//
// ProtonMail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ProtonMail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ProtonMail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package updater

import (
	"encoding/json"
	"io"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type clientProvider interface {
	GetAnonymousClient() pmapi.Client
}

type installer interface {
	InstallUpdate(*semver.Version, io.Reader) error
}

type Updater struct {
	cm        clientProvider
	installer installer
	kr        *crypto.KeyRing

	curVer        *semver.Version
	updateURLName string
	platform      string
	rollout       float64

	locker *locker
}

func New(
	cm clientProvider,
	installer installer,
	kr *crypto.KeyRing,
	curVer *semver.Version,
	updateURLName, platform string,
	rollout float64,
) *Updater {
	return &Updater{
		cm:            cm,
		installer:     installer,
		kr:            kr,
		curVer:        curVer,
		updateURLName: updateURLName,
		platform:      platform,
		rollout:       rollout,
		locker:        newLocker(),
	}
}

func (u *Updater) Watch(
	period time.Duration,
	handleUpdate func(VersionInfo) error,
	handleError func(error),
) func() {
	logrus.WithField("period", period).Info("Watching for updates")

	ticker := time.NewTicker(period)

	go func() {
		for {
			u.watch(handleUpdate, handleError)
			<-ticker.C
		}
	}()

	return ticker.Stop
}

func (u *Updater) watch(
	handleUpdate func(VersionInfo) error,
	handleError func(error),
) {
	logrus.Info("Checking for updates")

	latest, err := u.fetchVersionInfo()
	if err != nil {
		handleError(errors.Wrap(err, "failed to fetch version info"))
		return
	}

	if !latest.Version.GreaterThan(u.curVer) || u.rollout > latest.Rollout {
		logrus.WithError(err).Debug("No need to update")
		return
	}

	if u.curVer.LessThan(latest.MinAuto) {
		logrus.Debug("A manual update is required")
		// NOTE: Need to notify user that they must update manually.
		return
	}

	logrus.
		WithField("latest", latest.Version).
		WithField("current", u.curVer).
		Info("An update is available")

	if err := handleUpdate(latest); err != nil {
		handleError(errors.Wrap(err, "failed to handle update"))
	}
}

func (u *Updater) InstallUpdate(update VersionInfo) error {
	return u.locker.doOnce(func() error {
		logrus.WithField("package", update.Package).Info("Installing update package")

		client := u.cm.GetAnonymousClient()
		defer client.Logout()

		r, err := client.DownloadAndVerify(update.Package, update.Package+".sig", u.kr)
		if err != nil {
			return errors.Wrap(err, "failed to download and verify update package")
		}

		if err := u.installer.InstallUpdate(update.Version, r); err != nil {
			return errors.Wrap(err, "failed to install update package")
		}

		u.curVer = update.Version

		return nil
	})
}

func (u *Updater) fetchVersionInfo() (VersionInfo, error) {
	client := u.cm.GetAnonymousClient()
	defer client.Logout()

	r, err := client.DownloadAndVerify(
		u.getVersionFileURL(),
		u.getVersionFileURL()+".sig",
		u.kr,
	)
	if err != nil {
		return VersionInfo{}, err
	}

	var versionMap VersionMap

	if err := json.NewDecoder(r).Decode(&versionMap); err != nil {
		return VersionInfo{}, err
	}

	return versionMap[Channel], nil
}
