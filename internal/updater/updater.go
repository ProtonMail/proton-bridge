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

	"github.com/Masterminds/semver/v3"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var ErrManualUpdateRequired = errors.New("manual update is required")

type ClientProvider interface {
	GetAnonymousClient() pmapi.Client
}

type Installer interface {
	InstallUpdate(*semver.Version, io.Reader) error
}

type Updater struct {
	cm        ClientProvider
	installer Installer
	kr        *crypto.KeyRing

	curVer        *semver.Version
	updateURLName string
	platform      string
	rollout       float64

	locker *locker
}

func New(
	cm ClientProvider,
	installer Installer,
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

func (u *Updater) Check() (VersionInfo, error) {
	logrus.Info("Checking for updates")

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

func (u *Updater) IsUpdateApplicable(version VersionInfo) bool {
	if !version.Version.GreaterThan(u.curVer) {
		return false
	}

	if u.rollout > version.Rollout {
		return false
	}

	return true
}

func (u *Updater) CanInstall(version VersionInfo) bool {
	return !u.curVer.LessThan(version.MinAuto)
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
