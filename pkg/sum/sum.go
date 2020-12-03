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

package sum

import (
	"crypto/sha512"
	"io"
	"os"
	"path/filepath"
)

// RecursiveSum computes the sha512 sum of all files in the root directory and descendents.
// If a skipFile is provided (e.g. the path of a checksum file relative to rootDir), it (and its signature) is ignored.
func RecursiveSum(rootDir, skipFileName string) ([]byte, error) {
	hash := sha512.New()

	skipFile := filepath.Join(rootDir, skipFileName)
	skipFileSig := skipFile + ".sig"

	if err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// The hashfile itself isn't included in the hash.
		if path == skipFile || path == skipFileSig {
			return nil
		}

		rel, err := filepath.Rel(rootDir, path)
		if err != nil {
			return err
		}

		if _, err := hash.Write([]byte(rel)); err != nil {
			return err
		}

		f, err := os.Open(path) // nolint[gosec]
		if err != nil {
			return err
		}

		if _, err := io.Copy(hash, f); err != nil {
			return err
		}

		return f.Close()
	}); err != nil {
		return nil, err
	}

	return hash.Sum([]byte{}), nil
}
