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
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKeychainSettingsIO(t *testing.T) {
	dir := t.TempDir()

	// test loading non existing file. no error but loads defaults.
	settings, err := LoadKeychainSettings(dir)
	require.NoError(t, err)
	require.Equal(t, settings, KeychainSettings{})

	// test file creation
	settings.Helper = "dummy1"
	settings.DisableTest = true
	require.NoError(t, settings.Save(dir))

	// test reading existing file
	readBack, err := LoadKeychainSettings(dir)
	require.NoError(t, err)
	require.Equal(t, settings, readBack)

	// test file overwrite and read back
	settings.Helper = "dummy2"
	require.NoError(t, settings.Save(dir))
	readBack, err = LoadKeychainSettings(dir)
	require.NoError(t, err)
	require.Equal(t, settings, readBack)

	// test error on invalid content
	settingsFilePath := filepath.Join(dir, keychainSettingsFileName)
	require.NoError(t, os.WriteFile(settingsFilePath, []byte("][INVALID"), 0o600))
	_, err = LoadKeychainSettings(dir)
	require.Error(t, err)
}
