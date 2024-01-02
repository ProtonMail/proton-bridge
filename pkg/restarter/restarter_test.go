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

package restarter

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRemoveFlagWithValue(t *testing.T) {
	tests := []struct {
		argList  []string
		flag     string
		expected []string
	}{
		{[]string{}, "b", nil},
		{[]string{"-a", "-b=value", "-c"}, "b", []string{"-a", "-c"}},
		{[]string{"-a", "--b=value", "-c"}, "b", []string{"-a", "-c"}},
		{[]string{"-a", "-b", "value", "-c"}, "b", []string{"-a", "-c"}},
		{[]string{"-a", "--b", "value", "-c"}, "b", []string{"-a", "-c"}},
		{[]string{"-a", "-B=value", "-c"}, "b", []string{"-a", "-B=value", "-c"}},
	}

	for _, tt := range tests {
		require.Equal(t, removeFlagWithValue(tt.argList, tt.flag), tt.expected)
	}
}

func TestRemoveFlagsWithValue(t *testing.T) {
	tests := []struct {
		argList  []string
		flags    []string
		expected []string
	}{
		{[]string{}, []string{"a", "b"}, nil},
		{[]string{"-a", "-b=value", "-c"}, []string{"b"}, []string{"-a", "-c"}},
		{[]string{"-a", "--b=value", "-c"}, []string{"b", "c"}, []string{"-a"}},
		{[]string{"-a", "-b", "value", "-c"}, []string{"c", "b"}, []string{"-a"}},
	}

	for _, tt := range tests {
		require.Equal(t, removeFlagsWithValue(tt.argList, tt.flags...), tt.expected)
	}
}

func TestRemoveFlag(t *testing.T) {
	tests := []struct {
		argList  []string
		flag     string
		expected []string
	}{
		{[]string{}, "b", []string{}},
		{[]string{"-a", "-b", "-c"}, "b", []string{"-a", "-c"}},
		{[]string{"-a", "--b", "-b"}, "b", []string{"-a"}},
		{[]string{"-a", "-c"}, "b", []string{"-a", "-c"}},
	}

	for _, tt := range tests {
		require.Equal(t, removeFlag(tt.argList, tt.flag), tt.expected)
	}
}
