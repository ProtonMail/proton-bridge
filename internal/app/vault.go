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

package app

import (
	"fmt"
	"path"

	"github.com/ProtonMail/gluon/async"
	"github.com/ProtonMail/proton-bridge/v3/internal/certs"
	"github.com/ProtonMail/proton-bridge/v3/internal/constants"
	"github.com/ProtonMail/proton-bridge/v3/internal/locations"
	"github.com/ProtonMail/proton-bridge/v3/internal/vault"
	"github.com/ProtonMail/proton-bridge/v3/pkg/keychain"
	"github.com/sirupsen/logrus"
)

func WithVault(locations *locations.Locations, panicHandler async.PanicHandler, fn func(*vault.Vault, bool, bool) error) error {
	logrus.Debug("Creating vault")
	defer logrus.Debug("Vault stopped")

	// Create the encVault.
	encVault, insecure, corrupt, err := newVault(locations, panicHandler)
	if err != nil {
		return fmt.Errorf("could not create vault: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"insecure": insecure,
		"corrupt":  corrupt,
	}).Debug("Vault created")

	// Install the certificates if needed.
	if installed := encVault.GetCertsInstalled(); !installed {
		logrus.Debug("Installing certificates")

		certPEM, _ := encVault.GetBridgeTLSCert()

		if err := certs.NewInstaller().InstallCert(certPEM); err != nil {
			return fmt.Errorf("failed to install certs: %w", err)
		}

		if err := encVault.SetCertsInstalled(true); err != nil {
			return fmt.Errorf("failed to set certs installed: %w", err)
		}

		logrus.Debug("Certificates successfully installed")
	}

	// GODT-1950: Add teardown actions (e.g. to close the vault).

	return fn(encVault, insecure, corrupt)
}

func newVault(locations *locations.Locations, panicHandler async.PanicHandler) (*vault.Vault, bool, bool, error) {
	vaultDir, err := locations.ProvideSettingsPath()
	if err != nil {
		return nil, false, false, fmt.Errorf("could not get vault dir: %w", err)
	}

	logrus.WithField("vaultDir", vaultDir).Debug("Loading vault from directory")

	var (
		vaultKey []byte
		insecure bool
	)

	if key, err := loadVaultKey(vaultDir); err != nil {
		insecure = true

		// We store the insecure vault in a separate directory
		vaultDir = path.Join(vaultDir, "insecure")
	} else {
		vaultKey = key
	}

	gluonCacheDir, err := locations.ProvideGluonCachePath()
	if err != nil {
		return nil, false, false, fmt.Errorf("could not provide gluon path: %w", err)
	}

	vault, corrupt, err := vault.New(vaultDir, gluonCacheDir, vaultKey, panicHandler)
	if err != nil {
		return nil, false, false, fmt.Errorf("could not create vault: %w", err)
	}

	return vault, insecure, corrupt, nil
}

func loadVaultKey(vaultDir string) ([]byte, error) {
	helper, err := vault.GetHelper(vaultDir)
	if err != nil {
		return nil, fmt.Errorf("could not get keychain helper: %w", err)
	}

	kc, err := keychain.NewKeychain(helper, constants.KeyChainName)
	if err != nil {
		return nil, fmt.Errorf("could not create keychain: %w", err)
	}

	has, err := vault.HasVaultKey(kc)
	if err != nil {
		return nil, fmt.Errorf("could not check for vault key: %w", err)
	}

	if has {
		return vault.GetVaultKey(kc)
	}

	return vault.NewVaultKey(kc)
}
