// Copyright (c) 2020 Proton Technologies AG
//
// This file is part of ProtonMail Bridge.
//
// ProtonMail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ProtonMail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ProtonMail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package useragent

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetMacVersion(t *testing.T) {
	testData := map[string]struct{ major, minor, tiny int }{
		"10.14.4":     {10, 14, 4},
		"10.14.4\r\n": {10, 14, 4},
		"10.14.0":     {10, 14, 0},
		"10.14":       {10, 14, 0},
		"10":          {10, 0, 0},
	}

	for arg, exp := range testData {
		gotMajor, gotMinor, gotTiny := parseMacVersion(arg)
		assert.Equal(t, exp.major, gotMajor, "arg %q", arg)
		assert.Equal(t, exp.minor, gotMinor, "arg %q", arg)
		assert.Equal(t, exp.tiny, gotTiny, "arg %q", arg)
	}
}

func TestIsVersionCatalinaOrNewer(t *testing.T) {
	testData := map[struct{ major, minor int }]bool{
		{9, 0}:   false,
		{9, 15}:  false,
		{10, 13}: false,
		{10, 14}: false,
		{10, 15}: true,
		{10, 16}: true,
	}

	for args, exp := range testData {
		got := isVersionCatalinaOrNewer(args.major, args.minor)
		assert.Equal(t, exp, got, "version %q.%q", args.major, args.minor)
	}
}
