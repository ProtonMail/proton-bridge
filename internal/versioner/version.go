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

package versioner

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/Masterminds/semver/v3"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
)

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

// VerifyFiles verifies all files in the version directory.
func (v *Version) VerifyFiles(kr *crypto.KeyRing) error {
	return filepath.Walk(v.path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if filepath.Ext(path) == ".sig" || info.IsDir() {
			return nil
		}

		fileBytes, err := ioutil.ReadFile(path) // nolint[gosec]
		if err != nil {
			return err
		}

		sigBytes, err := ioutil.ReadFile(path + ".sig") // nolint[gosec]
		if err != nil {
			return err
		}

		return kr.VerifyDetached(
			crypto.NewPlainMessage(fileBytes),
			crypto.NewPGPSignature(sigBytes),
			crypto.GetUnixTime(),
		)
	})
}

// GetExecutable returns the full path to the executable of the given version.
// It returns an error if the executable is missing or does not have executable permissions set.
func (v *Version) GetExecutable(name string) (string, error) {
	exe := filepath.Join(v.path, getExeName(name))

	if !fileExists(exe) || !fileIsExecutable(exe) {
		return "", ErrNoExecutable
	}

	return exe, nil
}
