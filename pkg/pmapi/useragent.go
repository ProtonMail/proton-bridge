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
	"fmt"
	"runtime"
	"strings"
)

// removeBrackets handle unwanted brackets in client identification string and join with given joinBy parameter.
// Mac OS X Mail/13.0 (3601.0.4) -> Mac OS X Mail/13.0-3601.0.4 (joinBy = "-")
func removeBrackets(s string, joinBy string) (r string) {
	r = strings.ReplaceAll(s, " (", joinBy)
	r = strings.ReplaceAll(r, "(", joinBy) // Should be faster than regex.
	r = strings.ReplaceAll(r, ")", "")

	return
}

func formatUserAgent(clientName, clientVersion, os string) string {
	client := ""
	if clientName != "" {
		client = removeBrackets(clientName, "-")
		if clientVersion != "" {
			client += "/" + removeBrackets(clientVersion, "-")
		}
	}

	if os == "" {
		os = runtime.GOOS
	}

	os = removeBrackets(os, " ")

	return fmt.Sprintf("%s (%s)", client, os)
}
