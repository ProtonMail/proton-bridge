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

package service

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	dummyPort    = 12
	dummyCert    = "A dummy cert"
	dummyToken   = "A dummy token"
	tempFileName = "test.json"
	socketPath   = "/a/socket/file/path"
)

func TestConfig(t *testing.T) {
	conf1 := Config{
		Port:           dummyPort,
		Cert:           dummyCert,
		Token:          dummyToken,
		FileSocketPath: socketPath,
	}

	// Read-back test
	tempDir := t.TempDir()
	tempFilePath := filepath.Join(tempDir, tempFileName)
	require.NoError(t, conf1.save(tempFilePath))

	conf2 := Config{}
	require.NoError(t, conf2.Load(tempFilePath))
	require.Equal(t, conf1, conf2)

	// failure to load
	require.Error(t, conf2.Load(tempFilePath+"_"))

	// failure to save
	require.Error(t, conf2.save(filepath.Join(tempDir, "non/existing/folder", tempFileName)))
}
