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
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.

//go:build darwin

package certs

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCertInKeychain(t *testing.T) {
	// no trust settings change is performed, so this test will not trigger an OS security prompt.
	certPEM := generatePEMCertificate(t)
	require.True(t, osSupportCertInstall())
	require.False(t, isCertInKeychain(certPEM))
	require.NoError(t, addCertToKeychain(certPEM))
	require.True(t, isCertInKeychain(certPEM))
	require.Error(t, addCertToKeychain(certPEM))
	require.True(t, isCertInKeychain(certPEM))
	require.NoError(t, removeCertFromKeychain(certPEM))
	require.False(t, isCertInKeychain(certPEM))
	require.Error(t, removeCertFromKeychain(certPEM))
	require.False(t, isCertInKeychain(certPEM))
}

// This test require human interaction (macOS security prompts), and is disabled by default.
func _TestCertificateTrust(t *testing.T) { //nolint:unused
	certPEM := generatePEMCertificate(t)
	require.False(t, isCertTrusted(certPEM))
	require.NoError(t, addCertToKeychain(certPEM))
	require.NoError(t, setCertTrusted(certPEM))
	require.True(t, isCertTrusted(certPEM))
	require.NoError(t, removeCertTrust(certPEM))
	require.False(t, isCertTrusted(certPEM))
	require.NoError(t, removeCertFromKeychain(certPEM))
}

// This test require human interaction (macOS security prompts), and is disabled by default.
func _TestInstallAndRemove(t *testing.T) { //nolint:unused
	certPEM := generatePEMCertificate(t)

	// fresh install
	require.False(t, isCertInstalled(certPEM))
	require.NoError(t, installCert(certPEM))
	require.True(t, isCertInKeychain(certPEM))
	require.True(t, isCertTrusted(certPEM))
	require.True(t, isCertInstalled(certPEM))
	require.NoError(t, uninstallCert(certPEM))
	require.False(t, isCertInKeychain(certPEM))
	require.False(t, isCertTrusted(certPEM))
	require.False(t, isCertInstalled(certPEM))

	// Install where certificate is already in Keychain, but not trusted.
	require.NoError(t, addCertToKeychain(certPEM))
	require.False(t, isCertInstalled(certPEM))
	require.NoError(t, installCert(certPEM))
	require.True(t, isCertInstalled(certPEM))

	// Install where certificate is already installed
	require.NoError(t, installCert(certPEM))

	// Remove when certificate is not trusted.
	require.NoError(t, removeCertTrust(certPEM))
	require.NoError(t, uninstallCert(certPEM))
	require.False(t, isCertInstalled(certPEM))

	// Remove when certificate has already been removed.
	require.NoError(t, uninstallCert(certPEM))
	require.False(t, isCertTrusted(certPEM))
	require.False(t, isCertInKeychain(certPEM))
}

func generatePEMCertificate(t *testing.T) []byte {
	template, err := NewTLSTemplate()
	require.NoError(t, err)

	certPEM, _, err := GenerateCert(template)
	require.NoError(t, err)

	return certPEM
}
