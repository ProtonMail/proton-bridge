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

// This test implies human interactions to enter password and is disabled by default.
func _TestTrustedCertsDarwin(t *testing.T) { //nolint:unused
	template, err := NewTLSTemplate()
	require.NoError(t, err)

	certPEM, _, err := GenerateCert(template)
	require.NoError(t, err)

	require.Error(t, installCert([]byte{0}))   // Cannot install an invalid cert.
	require.Error(t, uninstallCert(certPEM))   // Cannot uninstall a cert that is not installed.
	require.NoError(t, installCert(certPEM))   // Can install a valid cert.
	require.NoError(t, installCert(certPEM))   // Can install an already installed cert.
	require.NoError(t, uninstallCert(certPEM)) // Can uninstall an installed cert.
	require.Error(t, uninstallCert(certPEM))   // Cannot uninstall an already uninstalled cert.
	require.NoError(t, installCert(certPEM))   // Can reinstall an uninstalled cert.
	require.NoError(t, uninstallCert(certPEM)) // Can uninstall a reinstalled cert.
}
