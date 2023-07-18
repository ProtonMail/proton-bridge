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

package usertypes

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestToType(t *testing.T) {
	type myString string

	// Slices of different types are not equal.
	require.NotEqual(t, []myString{"a", "b", "c"}, []string{"a", "b", "c"})

	// But converting them to the same type makes them equal.
	require.Equal(t, []myString{"a", "b", "c"}, MapTo[string, myString]([]string{"a", "b", "c"}))

	// The conversion can happen in the other direction too.
	require.Equal(t, []string{"a", "b", "c"}, MapTo[myString, string]([]myString{"a", "b", "c"}))
}
