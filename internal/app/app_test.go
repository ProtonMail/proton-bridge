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

package app

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolveMailServerAddresses(t *testing.T) {
	bindAddress, advertiseAddress, err := resolveMailServerAddresses("127.0.0.1", "")
	require.NoError(t, err)
	require.Equal(t, "127.0.0.1", bindAddress)
	require.Equal(t, "127.0.0.1", advertiseAddress)

	bindAddress, advertiseAddress, err = resolveMailServerAddresses("0.0.0.0", "mail.example.local")
	require.NoError(t, err)
	require.Equal(t, "0.0.0.0", bindAddress)
	require.Equal(t, "mail.example.local", advertiseAddress)

	bindAddress, advertiseAddress, err = resolveMailServerAddresses("::", "2001:db8::1")
	require.NoError(t, err)
	require.Equal(t, "::", bindAddress)
	require.Equal(t, "2001:db8::1", advertiseAddress)

	bindAddress, advertiseAddress, err = resolveMailServerAddresses("127.0.0.1", "localhost")
	require.NoError(t, err)
	require.Equal(t, "127.0.0.1", bindAddress)
	require.Equal(t, "localhost", advertiseAddress)
}

func TestResolveMailServerAddressesRejectsInvalidAddresses(t *testing.T) {
	_, _, err := resolveMailServerAddresses("", "")
	require.Error(t, err)

	_, _, err = resolveMailServerAddresses("localhost", "")
	require.Error(t, err)

	_, _, err = resolveMailServerAddresses("127.0.0.1:1143", "")
	require.Error(t, err)

	_, _, err = resolveMailServerAddresses("0.0.0.0", "")
	require.Error(t, err)

	_, _, err = resolveMailServerAddresses("127.0.0.1", "mail.example.local:1143")
	require.Error(t, err)

	_, _, err = resolveMailServerAddresses("127.0.0.1", "0.0.0.0")
	require.Error(t, err)

	_, _, err = resolveMailServerAddresses("127.0.0.1", "::")
	require.Error(t, err)

	_, _, err = resolveMailServerAddresses("127.0.0.1", "[::1]:1143")
	require.Error(t, err)

	_, _, err = resolveMailServerAddresses("127.0.0.1", "bad_hostname")
	require.Error(t, err)
}

func TestIsValidHostname(t *testing.T) {
	require.True(t, isValidHostname("localhost"))
	require.True(t, isValidHostname("mail.example.local"))
	require.True(t, isValidHostname("a-b.example"))

	require.False(t, isValidHostname(""))
	require.False(t, isValidHostname("mail..example"))
	require.False(t, isValidHostname("-mail.example"))
	require.False(t, isValidHostname("mail-.example"))
	require.False(t, isValidHostname("bad_hostname"))
	require.False(t, isValidHostname("mail.example.local:1143"))
}
