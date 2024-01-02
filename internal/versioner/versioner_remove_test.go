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

//go:build !darwin
// +build !darwin

package versioner

import (
	"path/filepath"
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// RemoveOldVersions is a noop on darwin; we don't test it there.

func TestRemoveOldVersions(t *testing.T) {
	tempDir := t.TempDir()

	v := newTestVersioner(t, "myCoolApp", tempDir, "2.3.4-beta", "2.3.4", "2.3.5", "2.4.0")

	allVersions, err := v.ListVersions()
	require.NoError(t, err)
	require.Len(t, allVersions, 4)

	assert.NoError(t, v.RemoveOldVersions())

	cleanedVersions, err := v.ListVersions()
	assert.NoError(t, err)
	assert.Len(t, cleanedVersions, 1)

	assert.Equal(t, semver.MustParse("2.4.0"), cleanedVersions[0].version)
	assert.Equal(t, filepath.Join(tempDir, "2.4.0"), cleanedVersions[0].path)
}
