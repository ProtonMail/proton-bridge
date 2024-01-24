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
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.

package versioner

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Masterminds/semver/v3"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/v3/pkg/sum"
)

const sumFile = ".sum"
const sigFile = ".sig"

type Version struct {
	version *semver.Version
	path    string
}

type Versions []*Version

func (v Versions) Len() int {
	return len(v)
}

func (v Versions) Less(i, j int) bool {
	return v[i].version.LessThan(v[j].version)
}

func (v Versions) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}

func (v Versions) HasVersion(want *semver.Version) bool {
	for i := range v {
		if v[i].version.Equal(want) {
			return true
		}
	}

	return false
}

func (v *Version) String() string {
	return fmt.Sprintf("%v", v.version)
}

func (v *Version) Equal(version *semver.Version) bool {
	return v.version.Equal(version)
}

func (v *Version) SemVer() *semver.Version {
	return v.version
}

// VerifyFiles verifies all files in the version directory.
func (v *Version) VerifyFiles(kr *crypto.KeyRing) error {
	return VerifyUpdateFolder(kr, v.path)
}

func VerifyUpdateFolder(kr *crypto.KeyRing, path string) error {
	fileBytes, err := os.ReadFile(filepath.Join(path, sumFile)) //nolint:gosec
	if err != nil {
		return err
	}

	sigBytes, err := os.ReadFile(filepath.Join(path, sumFile+sigFile)) //nolint:gosec
	if err != nil {
		return err
	}

	if err := kr.VerifyDetached(
		crypto.NewPlainMessage(fileBytes),
		crypto.NewPGPSignature(sigBytes),
		crypto.GetUnixTime(),
	); err != nil {
		return err
	}

	sum, err := sum.RecursiveSum(path, sumFile)
	if err != nil {
		return err
	}

	if !bytes.Equal(sum, fileBytes) {
		return fmt.Errorf(
			"sum mismatch: %v should be %v",
			base64.RawStdEncoding.EncodeToString(sum),
			base64.RawStdEncoding.EncodeToString(fileBytes),
		)
	}

	return nil
}

// GetExecutable returns the full path to the executable of the given version.
// It returns an error if the executable is missing or does not have executable permissions set.
func (v *Version) GetExecutable(name string) (string, error) {
	return getExecutableInDirectory(name, v.path)
}

// Remove removes this version directory.
func (v *Version) Remove() error {
	return os.RemoveAll(v.path)
}
