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
// along with Proton Mail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package vault

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/v3/pkg/keychain"
	"golang.org/x/exp/slices"
)

const vaultSecretName = "bridge-vault-key"

type Keychain struct {
	Helper string
}

func getKeychainPrefPath(vaultDir string) string {
	return filepath.Clean(filepath.Join(vaultDir, "keychain.json"))
}

func GetHelper(vaultDir string) (string, error) {
	if _, err := os.Stat(getKeychainPrefPath(vaultDir)); errors.Is(err, fs.ErrNotExist) {
		return "", nil
	}

	b, err := os.ReadFile(getKeychainPrefPath(vaultDir))
	if err != nil {
		return "", err
	}

	var keychain Keychain

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

	return os.WriteFile(getKeychainPrefPath(vaultDir), b, 0o600)
}

func HasVaultKey(kc *keychain.Keychain) (bool, error) {
	secrets, err := kc.List()
	if err != nil {
		return false, fmt.Errorf("could not list keychain: %w", err)
	}

	return slices.Contains(secrets, vaultSecretName), nil
}

func GetVaultKey(kc *keychain.Keychain) ([]byte, error) {
	_, keyEnc, err := kc.Get(vaultSecretName)
	if err != nil {
		return nil, fmt.Errorf("could not get keychain item: %w", err)
	}

	keyDec, err := base64.StdEncoding.DecodeString(keyEnc)
	if err != nil {
		return nil, fmt.Errorf("could not decode keychain item: %w", err)
	}

	return keyDec, nil
}

func SetVaultKey(kc *keychain.Keychain, key []byte) error {
	return kc.Put(vaultSecretName, base64.StdEncoding.EncodeToString(key))
}

func NewVaultKey(kc *keychain.Keychain) ([]byte, error) {
	tok, err := crypto.RandomToken(32)
	if err != nil {
		return nil, fmt.Errorf("could not generate random token: %w", err)
	}

	if err := kc.Put(vaultSecretName, base64.StdEncoding.EncodeToString(tok)); err != nil {
		return nil, fmt.Errorf("could not put keychain item: %w", err)
	}

	return tok, nil
}
