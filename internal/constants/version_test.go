// Copyright (c) 2024 Proton AG
//
// This file is part of Proton Mail Bridge.Bridge.
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

package constants

import (
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/stretchr/testify/require"
)

func TestPrereleaseSemver(t *testing.T) {
	ver, err := semver.MustParse("2.3.0+qa").SetPrerelease("dev")
	require.NoError(t, err)

	require.Equal(t, "2.3.0-dev+qa", ver.String())

	ver, err = semver.MustParse("2.3.0-dev+qa").SetPrerelease("dev")
	require.NoError(t, err)

	require.Equal(t, "2.3.0-dev+qa", ver.String())
}
