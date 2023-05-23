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
	"crypto/tls"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

// GetBridgeTLSCert returns the PEM-encoded certificate for the bridge.
// If CertPEMPath is set, it will attempt to read the certificate from the file.
// Otherwise, or on read/validation failure, it will return the certificate from the vault.
func (vault *Vault) GetBridgeTLSCert() ([]byte, []byte) {
	vault.lock.RLock()
	defer vault.lock.RUnlock()

	certs := vault.getUnsafe().Certs

	if certPath, keyPath := certs.CustomCertPath, certs.CustomKeyPath; certPath != "" && keyPath != "" {
		if certPEM, keyPEM, err := readPEMCert(certPath, keyPath); err == nil {
			return certPEM, keyPEM
		}

		logrus.Error("Failed to read certificate from file, using default")
	}

	return certs.Bridge.Cert, certs.Bridge.Key
}

// SetBridgeTLSCertPath sets the path to PEM-encoded certificates for the bridge.
func (vault *Vault) SetBridgeTLSCertPath(certPath, keyPath string) error {
	if _, _, err := readPEMCert(certPath, keyPath); err != nil {
		return fmt.Errorf("invalid certificate: %w", err)
	}

	return vault.modSafe(func(data *Data) {
		data.Certs.CustomCertPath = certPath
		data.Certs.CustomKeyPath = keyPath
	})
}

// SetBridgeTLSCertKey sets the path to PEM-encoded certificates for the bridge.
func (vault *Vault) SetBridgeTLSCertKey(cert, key []byte) error {
	return vault.modSafe(func(data *Data) {
		data.Certs.Bridge.Cert = cert
		data.Certs.Bridge.Key = key
	})
}

func (vault *Vault) GetCertsInstalled() bool {
	return vault.getSafe().Certs.Installed
}

func (vault *Vault) SetCertsInstalled(installed bool) error {
	return vault.modSafe(func(data *Data) {
		data.Certs.Installed = installed
	})
}

func readPEMCert(certPEMPath, keyPEMPath string) ([]byte, []byte, error) {
	certPEM, err := os.ReadFile(filepath.Clean(certPEMPath))
	if err != nil {
		return nil, nil, err
	}

	keyPEM, err := os.ReadFile(filepath.Clean(keyPEMPath))
	if err != nil {
		return nil, nil, err
	}

	if _, err := tls.X509KeyPair(certPEM, keyPEM); err != nil {
		return nil, nil, err
	}

	return certPEM, keyPEM, nil
}
