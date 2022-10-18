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

package app

import (
	"encoding/base64"
	"fmt"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/v2/internal/certs"
	"github.com/ProtonMail/proton-bridge/v2/internal/constants"
	"github.com/ProtonMail/proton-bridge/v2/internal/locations"
	"github.com/ProtonMail/proton-bridge/v2/internal/vault"
	"github.com/ProtonMail/proton-bridge/v2/pkg/keychain"
	"golang.org/x/exp/slices"
)

func withVault(locations *locations.Locations, fn func(*vault.Vault, bool, bool) error) error {
	// Create the encVault.
	encVault, insecure, corrupt, err := newVault(locations)
	if err != nil {
		return fmt.Errorf("could not create vault: %w", err)
	}

	// Install the certificates if needed.
	if installed := encVault.GetCertsInstalled(); !installed {
		if err := certs.NewInstaller().InstallCert(encVault.GetBridgeTLSCert()); err != nil {
			return fmt.Errorf("failed to install certs: %w", err)
		}

		if err := encVault.SetCertsInstalled(true); err != nil {
			return fmt.Errorf("failed to set certs installed: %w", err)
		}

		if err := encVault.SetCertsInstalled(true); err != nil {
			return fmt.Errorf("could not set certs installed: %w", err)
		}
	}

	// GODT-1950: Add teardown actions (e.g. to close the vault).

	return fn(encVault, insecure, corrupt)
}

func newVault(locations *locations.Locations) (*vault.Vault, bool, bool, error) {
	var insecure bool

	vaultDir, err := locations.ProvideSettingsPath()
	if err != nil {
		return nil, false, false, fmt.Errorf("could not get vault dir: %w", err)
	}

	var vaultKey []byte

	if key, err := getVaultKey(vaultDir); err != nil {
		insecure = true
	} else {
		vaultKey = key
	}

	gluonDir, err := locations.ProvideGluonPath()
	if err != nil {
		return nil, false, false, fmt.Errorf("could not provide gluon path: %w", err)
	}

	vault, corrupt, err := vault.New(vaultDir, gluonDir, vaultKey)
	if err != nil {
		return nil, false, false, fmt.Errorf("could not create vault: %w", err)
	}

	return vault, insecure, corrupt, nil
}

func getVaultKey(vaultDir string) ([]byte, error) {
	helper, err := vault.GetHelper(vaultDir)
	if err != nil {
		return nil, fmt.Errorf("could not get keychain helper: %w", err)
	}

	keychain, err := keychain.NewKeychain(helper, constants.KeyChainName)
	if err != nil {
		return nil, fmt.Errorf("could not create keychain: %w", err)
	}

	secrets, err := keychain.List()
	if err != nil {
		return nil, fmt.Errorf("could not list keychain: %w", err)
	}

	if !slices.Contains(secrets, vaultSecretName) {
		tok, err := crypto.RandomToken(32)
		if err != nil {
			return nil, fmt.Errorf("could not generate random token: %w", err)
		}

		if err := keychain.Put(vaultSecretName, base64.StdEncoding.EncodeToString(tok)); err != nil {
			return nil, fmt.Errorf("could not put keychain item: %w", err)
		}
	}

	_, keyEnc, err := keychain.Get(vaultSecretName)
	if err != nil {
		return nil, fmt.Errorf("could not get keychain item: %w", err)
	}

	keyDec, err := base64.StdEncoding.DecodeString(keyEnc)
	if err != nil {
		return nil, fmt.Errorf("could not decode keychain item: %w", err)
	}

	return keyDec, nil
}
