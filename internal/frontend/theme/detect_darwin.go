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

//go:build darwin
// +build darwin

package theme

import (
	"os"
	"path/filepath"

	"howett.net/plist"
)

func detectSystemTheme() Theme {
	home, err := os.UserHomeDir()
	if err != nil {
		return Light
	}

	path := filepath.Join(home, "/Library/Preferences/.GlobalPreferences.plist")
	prefFile, err := os.Open(path)
	if err != nil {
		return Light
	}
	defer prefFile.Close()

	var data struct {
		AppleInterfaceStyle string `plist:AppleInterfaceStyle`
	}

	dec := plist.NewDecoder(prefFile)
	err = dec.Decode(&data)
	if err == nil && data.AppleInterfaceStyle == "Dark" {
		return Dark
	}

	return Light
}
