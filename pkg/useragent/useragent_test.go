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

func TestIsVersionCatalinaOrNewer(t *testing.T) {
	testData := map[struct{ version string }]bool{
		{""}:         false,
		{"9.0.0"}:    false,
		{"9.15.0"}:   false,
		{"10.13.0"}:  false,
		{"10.14.0"}:  false,
		{"10.14.99"}: false,
		{"10.15.0"}:  true,
		{"10.16.0"}:  true,
		{"11.0.0"}:   true,
	}

	for args, exp := range testData {
		got := isVersionCatalinaOrNewer(args.version)
		assert.Equal(t, exp, got, "version %v", args.version)
	}
}
