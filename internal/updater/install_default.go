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

// +build !darwin

package updater

import (
	"io"

	"github.com/Masterminds/semver/v3"
	"github.com/ProtonMail/proton-bridge/internal/versioner"
)

type Installer struct {
	versioner *versioner.Versioner
}

func NewInstaller(versioner *versioner.Versioner) *Installer {
	return &Installer{
		versioner: versioner,
	}
}

func (i *Installer) InstallUpdate(version *semver.Version, r io.Reader) error {
	return i.versioner.InstallNewVersion(version, r)
}
