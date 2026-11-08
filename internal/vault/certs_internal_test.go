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
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.

package vault

import (
	"testing"

	"github.com/ProtonMail/gluon/async"
	"github.com/ProtonMail/proton-bridge/v3/internal/certs"
	"github.com/stretchr/testify/require"
)

func TestBridgeTLSCertAddressesResolvesSingleLabelHost(t *testing.T) {
	withTestResolver(t, []string{"murasaki.localdomain", "192.168.1.162"}, nil)

	require.Equal(t, []string{
		"murasaki",
		"murasaki.local",
		"murasaki.localdomain",
		"192.168.1.162",
	}, BridgeTLSCertAddresses("murasaki"))
}

func TestBridgeTLSCertAddressesDeduplicatesResolvedAddresses(t *testing.T) {
	withTestResolver(t, []string{"murasaki", "192.168.1.162", "192.168.1.162"}, nil)

	require.Equal(t, []string{
		"murasaki",
		"murasaki.local",
		"192.168.1.162",
	}, BridgeTLSCertAddresses("murasaki"))
}

func TestSetBridgeTLSCertForAddressIncludesResolvedNamesAndIPs(t *testing.T) {
	withTestResolver(t, []string{"murasaki.localdomain", "192.168.1.162"}, nil)
	s := newInternalTestVault(t)

	require.NoError(t, s.SetBridgeTLSCertForAddress("murasaki"))

	cert, key := s.GetBridgeTLSCert()
	require.NotEmpty(t, key)
	require.True(t, certs.CertMatchesAddress(cert, "murasaki"))
	require.True(t, certs.CertMatchesAddress(cert, "murasaki.local"))
	require.True(t, certs.CertMatchesAddress(cert, "murasaki.localdomain"))
	require.True(t, certs.CertMatchesAddress(cert, "192.168.1.162"))
}

func withTestResolver(t *testing.T, dnsAddresses []string, systemAddresses []string) {
	t.Helper()

	originalLookupDNS := lookupDNS
	originalLookupSystem := lookupSystem
	t.Cleanup(func() {
		lookupDNS = originalLookupDNS
		lookupSystem = originalLookupSystem
	})

	lookupDNS = func(string) []string {
		return dnsAddresses
	}

	lookupSystem = func(string) []string {
		return systemAddresses
	}
}

func newInternalTestVault(t *testing.T) *Vault {
	t.Helper()

	s, corrupt, err := New(t.TempDir(), t.TempDir(), []byte("my secret key"), async.NoopPanicHandler{})
	require.NoError(t, err)
	require.NoError(t, corrupt)

	return s
}
