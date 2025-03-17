// Copyright (c) 2025 Proton AG
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
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path"

	"github.com/ProtonMail/gluon/async"
	"github.com/ProtonMail/proton-bridge/v3/internal/certs"
	"github.com/ProtonMail/proton-bridge/v3/internal/constants"
	"github.com/ProtonMail/proton-bridge/v3/internal/locations"
	"github.com/ProtonMail/proton-bridge/v3/internal/sentry"
	"github.com/ProtonMail/proton-bridge/v3/internal/vault"
	"github.com/ProtonMail/proton-bridge/v3/pkg/keychain"
	"github.com/sirupsen/logrus"
)

func WithVault(reporter *sentry.Reporter, locations *locations.Locations, keychains *keychain.List, panicHandler async.PanicHandler, fn func(*vault.Vault, bool, bool) error) error {
	logrus.Debug("Creating vault")
	defer logrus.Debug("Vault stopped")

	// Create the encVault.
	encVault, insecure, corrupt, err := newVault(reporter, locations, keychains, panicHandler)
	if err != nil {
		return fmt.Errorf("could not create vault: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"insecure": insecure,
		"corrupt":  corrupt != nil,
	}).Debug("Vault created")

	if corrupt != nil {
		logrus.WithError(corrupt).Warn("Failed to load existing vault, vault has been reset")
	}

	cert, _ := encVault.GetBridgeTLSCert()
	certs.NewInstaller().LogCertInstallStatus(cert)

	// GODT-1950: Add teardown actions (e.g. to close the vault).

	return fn(encVault, insecure, corrupt != nil)
}

func newVault(reporter *sentry.Reporter, locations *locations.Locations, keychains *keychain.List, panicHandler async.PanicHandler) (*vault.Vault, bool, error, error) {
	vaultDir, err := locations.ProvideSettingsPath()
	if err != nil {
		return nil, false, nil, fmt.Errorf("could not get vault dir: %w", err)
	}

	logrus.WithField("vaultDir", vaultDir).Debug("Loading vault from directory")

	var (
		vaultKey       []byte
		insecure       bool
		lastUsedHelper string
	)

	if key, helper, err := loadVaultKey(vaultDir, keychains); err != nil {
		if reporter != nil {
			if rerr := reporter.ReportMessageWithContext("Could not load/create vault key", map[string]any{
				"keychainDefaultHelper":       keychains.GetDefaultHelper(),
				"keychainUsableHelpersLength": len(keychains.GetHelpers()),
				"error":                       err.Error(),
			}); rerr != nil {
				logrus.WithError(err).Info("Failed to report keychain issue to Sentry")
			}
		}

		logrus.WithError(err).Error("Could not load/create vault key")
		insecure = true

		// We store the insecure vault in a separate directory
		vaultDir = path.Join(vaultDir, "insecure")
	} else {
		vaultKey = key
		lastUsedHelper = helper
		logHashedVaultKey(vaultKey) // Log a hash of the vault key.
	}

	gluonCacheDir, err := locations.ProvideGluonCachePath()
	if err != nil {
		return nil, false, nil, fmt.Errorf("could not provide gluon path: %w", err)
	}

	userVault, corrupt, err := vault.New(vaultDir, gluonCacheDir, vaultKey, panicHandler)
	if err != nil {
		return nil, false, corrupt, fmt.Errorf("could not create vault: %w", err)
	}

	// Remember the last successfully used keychain and store that as the user preference.
	if err := vault.SetHelper(vaultDir, lastUsedHelper); err != nil {
		logrus.WithError(err).Error("Could not store last used keychain helper")
	}

	return userVault, insecure, corrupt, nil
}

// loadVaultKey - loads the key used to encrypt the vault alongside the keychain helper used to access it.
func loadVaultKey(vaultDir string, keychains *keychain.List) (key []byte, keychainHelper string, err error) {
	keychainHelper, err = vault.GetHelper(vaultDir)
	if err != nil {
		return nil, keychainHelper, fmt.Errorf("could not get keychain helper: %w", err)
	}

	kc, keychainHelper, err := keychain.NewKeychain(keychainHelper, constants.KeyChainName, keychains.GetHelpers(), keychains.GetDefaultHelper())
	if err != nil {
		return nil, keychainHelper, fmt.Errorf("could not create keychain: %w", err)
	}

	key, err = vault.GetVaultKey(kc)
	if err != nil {
		if keychain.IsErrKeychainNoItem(err) {
			logrus.WithError(err).Warn("no vault key found, generating new")
			key, err := vault.NewVaultKey(kc)
			return key, keychainHelper, err
		}

		return nil, keychainHelper, fmt.Errorf("could not check for vault key: %w", err)
	}

	return key, keychainHelper, nil
}

// logHashedVaultKey - computes a sha256 hash and encodes it to base 64. The resulting string is logged.
func logHashedVaultKey(vaultKey []byte) {
	hashedKey := sha256.Sum256(vaultKey)
	logrus.WithField("hashedKey", hex.EncodeToString(hashedKey[:])).Info("Found vault key")
}
