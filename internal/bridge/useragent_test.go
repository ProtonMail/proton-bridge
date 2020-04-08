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

package bridge

import (
	"runtime"
	"testing"

	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
	"github.com/stretchr/testify/assert"
)

func TestUpdateCurrentUserAgentGOOS(t *testing.T) {
	UpdateCurrentUserAgent("ver", "", "", "")
	assert.Equal(t, "Bridge/ver ("+runtime.GOOS+"; unknown client)", pmapi.CurrentUserAgent)
}

func TestUpdateCurrentUserAgentOS(t *testing.T) {
	UpdateCurrentUserAgent("ver", "os", "", "")
	assert.Equal(t, "Bridge/ver (os; unknown client)", pmapi.CurrentUserAgent)
}

func TestUpdateCurrentUserAgentClientVer(t *testing.T) {
	UpdateCurrentUserAgent("ver", "os", "", "cver")
	assert.Equal(t, "Bridge/ver (os; unknown client)", pmapi.CurrentUserAgent)
}

func TestUpdateCurrentUserAgentClientName(t *testing.T) {
	UpdateCurrentUserAgent("ver", "os", "mail", "")
	assert.Equal(t, "Bridge/ver (os; mail)", pmapi.CurrentUserAgent)
}

func TestUpdateCurrentUserAgentClientNameAndVersion(t *testing.T) {
	UpdateCurrentUserAgent("ver", "os", "mail", "cver")
	assert.Equal(t, "Bridge/ver (os; mail/cver)", pmapi.CurrentUserAgent)
}
