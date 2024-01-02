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

package logging

import (
	"testing"

	"github.com/ProtonMail/proton-bridge/v3/internal/constants"
	"github.com/stretchr/testify/require"
)

func TestLogging_MatchStackTraceName(t *testing.T) {
	filename := getStackTraceName(NewSessionID(), constants.AppName, constants.Version, constants.Tag)
	require.True(t, len(filename) > 0)
	require.True(t, MatchStackTraceName(filename))
	require.False(t, MatchStackTraceName("Invalid.log"))
}
