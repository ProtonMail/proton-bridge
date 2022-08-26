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

package useragent

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsVersionCatalinaOrNewer(t *testing.T) {
	testData := map[struct{ version string }]bool{
		{""}:       false,
		{"18.0.0"}: false,
		{"19.0.0"}: true,
		{"20.0.0"}: true,
		{"21.0.0"}: true,
	}

	for args, exp := range testData {
		got := isVersionEqualOrNewer(getMinCatalina(), args.version)
		assert.Equal(t, exp, got, "version %v", args.version)
	}
}

func TestIsVersionBigSurOrNewer(t *testing.T) {
	testData := map[struct{ version string }]bool{
		{""}:       false,
		{"18.0.0"}: false,
		{"19.0.0"}: false,
		{"20.0.0"}: true,
		{"21.0.0"}: true,
	}

	for args, exp := range testData {
		got := isVersionEqualOrNewer(getMinBigSur(), args.version)
		assert.Equal(t, exp, got, "version %v", args.version)
	}
}
