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
// +build darwin

package keychain

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// hostURL uniquely identifies the app's keychain items within the system keychain.
func hostURL(keychainName string) string {
	// Skip when it was in-app update and not manual
	if path, err := os.Executable(); err == nil && strings.Contains(path, "ProtonMail Bridge") {
		return fmt.Sprintf("ProtonMail%vService", cases.Title(language.Und).String(keychainName))
	}
	return fmt.Sprintf("Proton Mail %v", cases.Title(language.Und).String(keychainName))
}
