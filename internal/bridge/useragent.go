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
	"fmt"
	"runtime"

	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
)

// UpdateCurrentUserAgent updates user agent on pmapi so each request has this
// information in headers for statistic purposes.
func UpdateCurrentUserAgent(bridgeVersion, os, clientName, clientVersion string) {
	if os == "" {
		os = runtime.GOOS
	}
	mailClient := "unknown client"
	if clientName != "" {
		mailClient = clientName
		if clientVersion != "" {
			mailClient += "/" + clientVersion
		}
	}
	pmapi.CurrentUserAgent = fmt.Sprintf("Bridge/%s (%s; %s)", bridgeVersion, os, mailClient)
}
