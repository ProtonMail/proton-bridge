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

//go:build darwin

package clientconfig

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEscapeXMLString(t *testing.T) {
	require.Equal(t, escapeXMLString(`abc&&''""<<>>def`), `abc&amp;&amp;&apos;&apos;&quot;&quot;&lt;&lt;&gt;&gt;def`)
}

// This test requires human interaction (user configuration profile installation prompt). It is for debugging purpose and is disabled by default.
func _TestInstallCert(t *testing.T) { //nolint:unused
	require.NoError(
		t,
		(&AppleMail{}).Configure(`127.0.0.1`, 1143, 1025, true, false, `user&>>`, `<<abc&&'"def>>`, `user&a`, []byte(`ir8R9vhdNXyB7isWzhyEkQ`)),
	)
}
