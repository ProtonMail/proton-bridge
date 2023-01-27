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

package vault_test

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVault_Cookies(t *testing.T) {
	// create a new test vault.
	s := newVault(t)

	// Check the default cookies are empty.
	cookies, err := s.GetCookies()
	require.NoError(t, err)
	require.Empty(t, cookies)

	// Set some cookies.
	require.NoError(t, s.SetCookies([]byte("something")))

	// Check the cookies are as set.
	newCookies, err := s.GetCookies()
	require.NoError(t, err)
	require.Equal(t, []byte("something"), newCookies)
}
