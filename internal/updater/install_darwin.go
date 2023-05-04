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
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.

package updater

import (
	"compress/gzip"
	"io"
	"os"
	"path/filepath"

	"github.com/Masterminds/semver/v3"
	"github.com/ProtonMail/proton-bridge/v3/internal/versioner"
	"github.com/ProtonMail/proton-bridge/v3/pkg/tar"
	"github.com/pkg/errors"
)

type InstallerDarwin struct{}

func NewInstaller(*versioner.Versioner) *InstallerDarwin {
	return &InstallerDarwin{}
}

func (i *InstallerDarwin) InstallUpdate(_ *semver.Version, r io.Reader) error {
	gr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer func() { _ = gr.Close() }()

	tempDir, err := os.MkdirTemp("", "proton-update-source")
	if err != nil {
		return errors.Wrap(err, "failed to get temporary update directory")
	}

	if err := tar.UntarToDir(gr, tempDir); err != nil {
		return errors.Wrap(err, "failed to unpack update package")
	}

	exePath, err := os.Executable()
	if err != nil {
		return errors.Wrap(err, "failed to determine current executable path")
	}

	oldBundle := filepath.Dir(filepath.Dir(filepath.Dir(exePath)))
	newBundle := filepath.Join(tempDir, filepath.Base(oldBundle))

	return syncFolders(oldBundle, newBundle)
}

func (i *InstallerDarwin) IsAlreadyInstalled(_ *semver.Version) bool {
	return false
}
