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

package vault

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

const keychainSettingsFileName = "keychain.json"

// KeychainSettings holds settings related to the keychain. It is serialized in the vault directory.
type KeychainSettings struct {
	Helper      string // The helper used for keychain.
	DisableTest bool   // Is the keychain test on startup disabled?
}

// LoadKeychainSettings load keychain settings from the vaultDir folder, or returns a default one if the file
// does not exists or is invalid.
func LoadKeychainSettings(vaultDir string) (KeychainSettings, error) {
	path := filepath.Join(vaultDir, keychainSettingsFileName)
	bytes, err := os.ReadFile(path) //nolint:gosec
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			logrus.
				WithFields(logrus.Fields{"pkg": "vault", "path": path}).
				Trace("Keychain settings file does not exists, default values will be used")
			return KeychainSettings{}, nil
		}
		return KeychainSettings{}, err
	}

	var result KeychainSettings
	if err := json.Unmarshal(bytes, &result); err != nil {
		return KeychainSettings{}, fmt.Errorf("keychain settings file is invalid settings: %w", err)
	}

	return result, nil
}

// Save saves the keychain settings in a file in the vaultDir folder.
func (k KeychainSettings) Save(vaultDir string) error {
	bytes, err := json.MarshalIndent(k, "", "  ")
	if err != nil {
		return err
	}

	if err = os.MkdirAll(vaultDir, 0o700); err != nil {
		return err
	}

	path := filepath.Join(vaultDir, keychainSettingsFileName)
	return os.WriteFile(path, bytes, 0o600)
}
