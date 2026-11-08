// Copyright (c) 2026 Proton AG
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

package vault_test

import (
	"testing"

	"github.com/ProtonMail/proton-bridge/v3/internal/certs"
	"github.com/ProtonMail/proton-bridge/v3/internal/vault"
	"github.com/stretchr/testify/require"
)

func TestVault_TLSCerts(t *testing.T) {
	// create a new test vault.
	s := newVault(t)

	// Check the default bridge TLS certs.
	cert, key := s.GetBridgeTLSCert()
	require.NotEmpty(t, cert)
	require.NotEmpty(t, key)
}

func TestVault_SetBridgeTLSCertForAddress(t *testing.T) {
	s := newVault(t)

	require.NoError(t, s.SetBridgeTLSCertForAddress("192.168.1.10"))

	cert, key := s.GetBridgeTLSCert()
	require.NotEmpty(t, key)
	require.True(t, certs.CertMatchesAddress(cert, "192.168.1.10"))
	require.False(t, certs.CertMatchesAddress(cert, "127.0.0.1"))
}

func TestBridgeTLSCertAddressesWithIPAddress(t *testing.T) {
	require.Equal(t, []string{"192.168.1.10"}, vault.BridgeTLSCertAddresses("192.168.1.10"))
}
