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
// along with Proton Mail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package vault

import (
	"encoding/base64"
	"fmt"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/v3/pkg/keychain"
	"github.com/sirupsen/logrus"
)

const vaultSecretName = "bridge-vault-key"

func GetShouldSkipKeychainTest(vaultDir string) (bool, error) {
	settings, err := LoadKeychainSettings(vaultDir)
	if err != nil {
		return false, err
	}

	return settings.DisableTest, nil
}

func SetShouldSkipKeychainTest(vaultDir string, skip bool) error {
	settings, err := LoadKeychainSettings(vaultDir)
	if err != nil {
		return err
	}

	log := logrus.WithFields(logrus.Fields{"pkg": "vault", "skipKeychainTest": skip})
	if skip == settings.DisableTest {
		log.Info("Skipping change of keychain test setting as value is not modified")
		return nil
	}

	logrus.WithFields(logrus.Fields{"pkg": "vault", "skipKeychainTest": skip}).Info("Setting keychain test skip option")
	settings.DisableTest = skip
	return settings.Save(vaultDir)
}

func GetHelper(vaultDir string) (string, error) {
	settings, err := LoadKeychainSettings(vaultDir)
	if err != nil {
		return "", err
	}
	return settings.Helper, nil
}

func SetHelper(vaultDir, helper string) error {
	settings, err := LoadKeychainSettings(vaultDir)
	if err != nil {
		return err
	}

	settings.Helper = helper
	return settings.Save(vaultDir)
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
