// Copyright (c) 2022 Proton AG
//
// This file is part of Proton Mail Bridge.Bridge.
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

package context

import (
	"math/rand"
	"os"

	"github.com/ProtonMail/proton-bridge/v2/internal/config/settings"
)

type fakeSettings struct {
	*settings.Settings
	dir string
}

// newFakeSettings creates a temporary folder for files.
// It's expected the test calls `ClearData` before finish to remove it from the file system.
func newFakeSettings() *fakeSettings {
	dir, err := os.MkdirTemp("", "test-settings")
	if err != nil {
		panic(err)
	}

	s := &fakeSettings{
		Settings: settings.New(dir),
		dir:      dir,
	}

	// We should use nonstandard ports to not conflict with bridge.
	s.SetInt(settings.IMAPPortKey, 21100+rand.Intn(100)) //nolint:gosec // G404 It is OK to use weak random number generator here
	s.SetInt(settings.SMTPPortKey, 21200+rand.Intn(100)) //nolint:gosec // G404 It is OK to use weak random number generator here

	return s
}
