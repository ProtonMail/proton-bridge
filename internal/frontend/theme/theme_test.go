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

// Package settings provides access to persistent user settings.
package theme

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsAvailable(t *testing.T) {
	r := require.New(t)

	want := "dark"

	r.True(IsAvailable("dark"))
	r.True(IsAvailable(Dark))
	r.True(IsAvailable(Theme(want)))

	want = "light"
	r.True(IsAvailable("light"))
	r.True(IsAvailable(Light))
	r.True(IsAvailable(Theme(want)))

	want = "molokai"
	r.False(IsAvailable(""))
	r.False(IsAvailable("molokai"))
	r.False(IsAvailable(Theme(want)))
}
