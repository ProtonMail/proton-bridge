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

package updater

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/Masterminds/semver/v3"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/v3/internal/versioner"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
)

const updateFileVersion = 1

var (
	ErrDownloadVerify              = errors.New("failed to download or verify the update")
	ErrInstall                     = errors.New("failed to install the update")
	ErrUpdateAlreadyInstalled      = errors.New("update is already installed")
	ErrVersionFileDownloadOrVerify = errors.New("failed to download or verify the version file")
	ErrReleaseUpdatePackageMissing = errors.New("release update package is missing")
)

type Downloader interface {
	DownloadAndVerify(ctx context.Context, kr *crypto.KeyRing, url, sig string) ([]byte, error)
}

type Installer interface {
	IsAlreadyInstalled(*semver.Version) bool
	InstallUpdate(*semver.Version, io.Reader) error
}

type Updater struct {
	versioner *versioner.Versioner
	installer Installer
	verifier  *crypto.KeyRing
	product   string
	platform  string
	version   uint
}

func NewUpdater(ver *versioner.Versioner, verifier *crypto.KeyRing, product, platform string) *Updater {
	return &Updater{
		versioner: ver,
		installer: NewInstaller(ver),
		verifier:  verifier,
		product:   product,
		platform:  platform,
		version:   updateFileVersion,
	}
}

func (u *Updater) GetVersionInfoLegacy(ctx context.Context, downloader Downloader, channel Channel) (VersionInfoLegacy, error) {
	b, err := downloader.DownloadAndVerify(
		ctx,
		u.verifier,
		u.getVersionFileURLLegacy(),
		u.getVersionFileURLLegacy()+".sig",
	)
	if err != nil {
		return VersionInfoLegacy{}, fmt.Errorf("%w: %w", ErrVersionFileDownloadOrVerify, err)
	}

	var versionMap VersionMap

	if err := json.Unmarshal(b, &versionMap); err != nil {
		return VersionInfoLegacy{}, err
	}

	version, ok := versionMap[channel]
	if !ok {
		return VersionInfoLegacy{}, errors.New("no updates available for this channel")
	}

	return version, nil
}

func (u *Updater) GetVersionInfo(ctx context.Context, downloader Downloader) (VersionInfo, error) {
	b, err := downloader.DownloadAndVerify(
		ctx,
		u.verifier,
		u.getVersionFileURL(),
		u.getVersionFileURL()+".sig",
	)
	if err != nil {
		return VersionInfo{}, fmt.Errorf("%w: %w", ErrVersionFileDownloadOrVerify, err)
	}

	var releases VersionInfo

	if err := json.Unmarshal(b, &releases); err != nil {
		return VersionInfo{}, err
	}

	return releases, nil
}

func (u *Updater) InstallUpdateLegacy(ctx context.Context, downloader Downloader, update VersionInfoLegacy) error {
	if u.installer.IsAlreadyInstalled(update.Version) {
		return ErrUpdateAlreadyInstalled
	}

	b, err := downloader.DownloadAndVerify(
		ctx,
		u.verifier,
		update.Package,
		update.Package+".sig",
	)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrDownloadVerify, err)
	}

	if err := u.installer.InstallUpdate(update.Version, bytes.NewReader(b)); err != nil {
		logrus.WithError(err).Error("Failed to install update")
		return ErrInstall
	}

	return nil
}

func (u *Updater) InstallUpdate(ctx context.Context, downloader Downloader, release Release) error {
	if u.installer.IsAlreadyInstalled(release.Version) {
		return ErrUpdateAlreadyInstalled
	}

	// Find update package
	idx := slices.IndexFunc(release.File, func(file File) bool {
		return file.Identifier == PackageIdentifier
	})

	if idx == -1 {
		logrus.WithFields(logrus.Fields{
			"release_version": release.Version,
		}).Error("Update release does not contain update package")
		return ErrReleaseUpdatePackageMissing
	}

	releaseUpdatePackage := release.File[idx]

	b, err := downloader.DownloadAndVerify(
		ctx,
		u.verifier,
		releaseUpdatePackage.URL,
		releaseUpdatePackage.URL+".sig",
	)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrDownloadVerify, err)
	}

	if err := u.installer.InstallUpdate(release.Version, bytes.NewReader(b)); err != nil {
		logrus.WithError(err).Error("Failed to install update")
		return ErrInstall
	}

	return nil
}

func (u *Updater) RemoveOldUpdates() error {
	return u.versioner.RemoveOldVersions()
}

// getVersionFileURLLegacy returns the URL of the version file.
// For example:
//   - https://protonmail.com/download/bridge/version_linux.json
func (u *Updater) getVersionFileURLLegacy() string {
	return fmt.Sprintf("%v/%v/version_%v.json", Host, u.product, u.platform)
}

// getVersionFileURL returns the URL of the version file.
// For example:
//   - https://protonmail.com/download/windows/x86/v1/version.json
//   - https://protonmail.com/download/linux/x86/v1/version.json
//   - https://protonmail.com/download/darwin/universal/v1/version.json
func (u *Updater) getVersionFileURL() string {
	switch u.platform {
	case "darwin":
		return fmt.Sprintf("%v/%v/%v/universal/v%v/version.json", Host, u.product, u.platform, u.version)
	default:
		return fmt.Sprintf("%v/%v/%v/x86/v%v/version.json", Host, u.product, u.platform, u.version)
	}
}
