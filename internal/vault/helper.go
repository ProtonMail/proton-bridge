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
// along with Proton Mail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package vault

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

type Keychain struct {
	Helper string
}

func GetHelper(vaultDir string) (string, error) {
	var keychain Keychain

	if _, err := os.Stat(filepath.Join(vaultDir, "keychain.json")); errors.Is(err, fs.ErrNotExist) {
		return "", nil
	}

	b, err := os.ReadFile(filepath.Join(vaultDir, "keychain.json"))
	if err != nil {
		return "", err
	}

	if err := json.Unmarshal(b, &keychain); err != nil {
		return "", err
	}

	return keychain.Helper, nil
}

func SetHelper(vaultDir, helper string) error {
	b, err := json.MarshalIndent(Keychain{Helper: helper}, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(vaultDir, "keychain.json"), b, 0o600)
}
