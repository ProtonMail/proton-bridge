// Copyright (c) 2021 Proton Technologies AG
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

package pmapi

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdateCurrentUserAgentGOOS(t *testing.T) {
	userAgent := formatUserAgent("", "", "")
	assert.Equal(t, " ("+runtime.GOOS+")", userAgent)
}

func TestUpdateCurrentUserAgentOS(t *testing.T) {
	userAgent := formatUserAgent("", "", "os")
	assert.Equal(t, " (os)", userAgent)
}

func TestUpdateCurrentUserAgentClientVer(t *testing.T) {
	userAgent := formatUserAgent("", "ver", "os")
	assert.Equal(t, " (os)", userAgent)
}

func TestUpdateCurrentUserAgentClientName(t *testing.T) {
	userAgent := formatUserAgent("mail", "", "os")
	assert.Equal(t, "mail (os)", userAgent)
}

func TestUpdateCurrentUserAgentClientNameAndVersion(t *testing.T) {
	userAgent := formatUserAgent("mail", "ver", "os")
	assert.Equal(t, "mail/ver (os)", userAgent)
}

func TestRemoveBrackets(t *testing.T) {
	userAgent := formatUserAgent("mail (submail)", "ver (subver)", "os (subos)")
	assert.Equal(t, "mail-submail/ver-subver (os subos)", userAgent)
}
