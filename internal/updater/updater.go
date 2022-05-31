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
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.

package updater

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/Masterminds/semver/v3"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/v2/internal/config/settings"
	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var ErrManualUpdateRequired = errors.New("manual update is required")

type Installer interface {
	InstallUpdate(*semver.Version, io.Reader) error
}

type Settings interface {
	Get(string) string
	Set(string, string)
	GetFloat64(string) float64
}

type Updater struct {
	cm        pmapi.Manager
	installer Installer
	settings  Settings
	kr        *crypto.KeyRing

	curVer        *semver.Version
	updateURLName string
	platform      string

	locker *locker
}

func New(
	cm pmapi.Manager,
	installer Installer,
	s Settings,
	kr *crypto.KeyRing,
	curVer *semver.Version,
	updateURLName, platform string,
) *Updater {
	// If there's some unexpected value in the preferences, we force it back onto the default channel.
	// This prevents users from screwing up silent updates by modifying their prefs.json file.
	if channel := UpdateChannel(s.Get(settings.UpdateChannelKey)); !(channel == StableChannel || channel == EarlyChannel) {
		s.Set(settings.UpdateChannelKey, string(DefaultUpdateChannel))
	}

	return &Updater{
		cm:            cm,
		installer:     installer,
		settings:      s,
		kr:            kr,
		curVer:        curVer,
		updateURLName: updateURLName,
		platform:      platform,
		locker:        newLocker(),
	}
}

func (u *Updater) Check() (VersionInfo, error) {
	logrus.Info("Checking for updates")

	b, err := u.cm.DownloadAndVerify(
		u.kr,
		u.getVersionFileURL(),
		u.getVersionFileURL()+".sig",
	)
	if err != nil {
		return VersionInfo{}, err
	}

	var versionMap VersionMap

	if err := json.Unmarshal(b, &versionMap); err != nil {
		return VersionInfo{}, err
	}

	version, ok := versionMap[u.settings.Get(settings.UpdateChannelKey)]
	if !ok {
		return VersionInfo{}, errors.New("no updates available for this channel")
	}

	return version, nil
}

func (u *Updater) IsUpdateApplicable(version VersionInfo) bool {
	if !version.Version.GreaterThan(u.curVer) {
		return false
	}

	if u.settings.GetFloat64(settings.RolloutKey) > version.RolloutProportion {
		return false
	}

	return true
}

func (u *Updater) IsDowngrade(version VersionInfo) bool {
	return version.Version.LessThan(u.curVer)
}

func (u *Updater) CanInstall(version VersionInfo) bool {
	if version.MinAuto == nil {
		return true
	}

	return !u.curVer.LessThan(version.MinAuto)
}

func (u *Updater) InstallUpdate(update VersionInfo) error {
	return u.locker.doOnce(func() error {
		logrus.WithField("package", update.Package).Info("Installing update package")

		b, err := u.cm.DownloadAndVerify(u.kr, update.Package, update.Package+".sig")
		if err != nil {
			return errors.Wrap(ErrDownloadVerify, err.Error())
		}

		if err := u.installer.InstallUpdate(update.Version, bytes.NewReader(b)); err != nil {
			return errors.Wrap(ErrInstall, err.Error())
		}

		u.curVer = update.Version

		return nil
	})
}
