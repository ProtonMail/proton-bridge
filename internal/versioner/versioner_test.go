// Copyright (c) 2022 Proton AG
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

package versioner

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListVersions(t *testing.T) {
	updates, err := ioutil.TempDir("", "updates")
	require.NoError(t, err)

	v := newTestVersioner(t, "myCoolApp", updates, "2.3.4-beta", "2.3.4", "2.3.5", "2.4.0")

	versions, err := v.ListVersions()
	require.NoError(t, err)

	assert.Equal(t, semver.MustParse("2.4.0"), versions[0].version)
	assert.Equal(t, filepath.Join(updates, "2.4.0"), versions[0].path)

	assert.Equal(t, semver.MustParse("2.3.5"), versions[1].version)
	assert.Equal(t, filepath.Join(updates, "2.3.5"), versions[1].path)

	assert.Equal(t, semver.MustParse("2.3.4"), versions[2].version)
	assert.Equal(t, filepath.Join(updates, "2.3.4"), versions[2].path)

	assert.Equal(t, semver.MustParse("2.3.4-beta"), versions[3].version)
	assert.Equal(t, filepath.Join(updates, "2.3.4-beta"), versions[3].path)
}

func newTestVersioner(t *testing.T, exeName, updates string, versions ...string) *Versioner {
	for _, version := range versions {
		makeDummyVersionDirectory(t, exeName, updates, version)
	}

	return New(updates)
}

func makeDummyVersionDirectory(t *testing.T, exeName, updates, version string) string {
	target := filepath.Join(updates, version)
	require.NoError(t, os.Mkdir(target, 0o700))

	exe, err := os.Create(filepath.Join(target, getExeName(exeName)))
	require.NoError(t, err)
	require.NotNil(t, exe)
	require.NoError(t, exe.Close())
	require.NoError(t, os.Chmod(exe.Name(), 0o700))

	sig, err := os.Create(filepath.Join(target, getExeName(exeName)+".sig"))
	require.NoError(t, err)
	require.NotNil(t, sig)
	require.NoError(t, sig.Close())

	return target
}
