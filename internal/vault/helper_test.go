// Copyright (c) 2024 Proton AG
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

	"github.com/stretchr/testify/require"
)

func TestShouldSkipKeychainTestAccessors(t *testing.T) {
	dir := t.TempDir()
	skip, err := GetShouldSkipKeychainTest(dir)
	require.NoError(t, err)
	require.False(t, skip)
	require.NoError(t, SetShouldSkipKeychainTest(dir, true))
	skip, err = GetShouldSkipKeychainTest(dir)
	require.NoError(t, err)
	require.True(t, skip)
}

func TestHelperAccessors(t *testing.T) {
	dir := t.TempDir()
	helper, err := GetHelper(dir)
	require.NoError(t, err)
	require.Zero(t, len(helper))
	require.NoError(t, SetHelper(dir, "keychain"))
	helper, err = GetHelper(dir)
	require.NoError(t, err)
	require.Equal(t, "keychain", helper)
}
