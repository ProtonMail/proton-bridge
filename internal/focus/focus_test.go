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
// along with Proton Mail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package focus

import (
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/stretchr/testify/require"
)

func TestFocus_Raise(t *testing.T) {
	// Start the focus service.
	service, err := NewService(semver.MustParse("1.2.3"))
	require.NoError(t, err)

	// Try to dial it, it should succeed.
	require.True(t, TryRaise())

	// The service should report a raise call.
	<-service.GetRaiseCh()

	// Stop the service.
	service.Close()

	// Try to dial it, it should fail.
	require.False(t, TryRaise())
}

func TestFocus_Version(t *testing.T) {
	// Start the focus service.
	_, err := NewService(semver.MustParse("1.2.3"))
	require.NoError(t, err)

	// Try to dial it, it should succeed.
	version, ok := TryVersion()
	require.True(t, ok)
	require.Equal(t, "1.2.3", version.String())
}
