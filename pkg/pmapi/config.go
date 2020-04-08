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

package pmapi

import (
	"net/http"
	"runtime"
)

// RootURL is the API root URL.
//
// This can be changed using build flags: pmapi_local for "http://localhost/api",
// pmapi_dev or pmapi_prod. Default is pmapi_prod.
var RootURL = "https://api.protonmail.ch" //nolint[gochecknoglobals]

// CurrentUserAgent is the default User-Agent for go-pmapi lib. This can be changed to program
// version and email client.
// e.g. Bridge/1.0.4 (Windows) MicrosoftOutlook/16.0.9330.2087
var CurrentUserAgent = "GoPMAPI/1.0.14 (" + runtime.GOOS + "; no client)" //nolint[gochecknoglobals]

// The HTTP transport to use by default.
var defaultTransport = &http.Transport{ //nolint[gochecknoglobals]
	Proxy: http.ProxyFromEnvironment,
}

// checkTLSCerts controls whether TLS certs are checked against known fingerprints.
// The default is for this to always be done.
var checkTLSCerts = true //nolint[gochecknoglobals]
