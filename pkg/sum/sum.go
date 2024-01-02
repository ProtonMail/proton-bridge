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

package sum

import (
	"crypto/sha512"
	"encoding/base64"
	"io"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

// RecursiveSum computes the sha512 sum of all files in the root directory and descendents.
// If a skipFile is provided (e.g. the path of a checksum file relative to
// rootDir), it (and its signature) is ignored.
func RecursiveSum(rootDir, skipFileName string) ([]byte, error) {
	hash := sha512.New()
	// In windows filepath accepts both delimiters `\` and `/`. In order to
	// to properly skip file we have to choose one native delimiter.
	rootDir = filepath.FromSlash(rootDir)
	skipFile := filepath.Join(rootDir, skipFileName)
	skipFileSig := skipFile + ".sig"

	if err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		log := logrus.
			WithField("path", path).
			WithField("sum", base64.StdEncoding.EncodeToString(hash.Sum([]byte{})))
		log.Debug("Next file")
		if err != nil {
			log.WithError(err).Error("Walk failed")
			return err
		}
		if info.IsDir() {
			log.Debug("Skip dir")
			return nil
		}

		// The hashfile itself isn't included in the hash.
		if path == skipFile || path == skipFileSig {
			log.Debug("Skip file")
			return nil
		}

		rel, err := filepath.Rel(rootDir, path)
		if err != nil {
			log.WithError(err).Error("Failed to find relative path")
			return err
		}
		if _, err := hash.Write([]byte(rel)); err != nil {
			log.WithError(err).Error("Failed to write path")
			return err
		}
		f, err := os.Open(path) //nolint:gosec
		if err != nil {
			log.WithError(err).Error("Failed to open file")
			return err
		}
		if _, err := io.Copy(hash, f); err != nil {
			log.WithError(err).Error("Copy to hash failed")
			return err
		}
		return f.Close()
	}); err != nil {
		return nil, err
	}

	return hash.Sum([]byte{}), nil
}
