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

package app

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestCheckSkipKeychainTest(t *testing.T) {
	var expectedResult bool
	dir := t.TempDir()
	app := cli.App{
		Flags: []cli.Flag{
			cliFlagEnableKeychainTest,
			cliFlagDisableKeychainTest,
		},
		Action: func(c *cli.Context) error {
			require.Equal(t, expectedResult, checkSkipKeychainTest(c, dir))
			return nil
		},
	}

	noArgs := []string{"appName"}
	enableArgs := []string{"appName", "-" + flagEnableKeychainTest}
	disableArgs := []string{"appName", "-" + flagDisableKeychainTest}
	bothArgs := []string{"appName", "-" + flagDisableKeychainTest, "-" + flagEnableKeychainTest}

	const trueOnlyOnMac = runtime.GOOS == "darwin"

	expectedResult = false
	require.NoError(t, app.Run(noArgs))

	expectedResult = trueOnlyOnMac
	require.NoError(t, app.Run(disableArgs))
	require.NoError(t, app.Run(noArgs))

	expectedResult = false
	require.NoError(t, app.Run(enableArgs))
	require.NoError(t, app.Run(noArgs))

	expectedResult = trueOnlyOnMac
	require.NoError(t, app.Run(disableArgs))

	expectedResult = false
	require.NoError(t, app.Run(bothArgs))
}
