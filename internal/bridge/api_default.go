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
// along with Proton Mail Bridge.  If not, see <https://www.gnu.org/licenses/>.

//go:build !build_qa && !test_integration

package bridge

import (
	"net/http"

	"github.com/Masterminds/semver/v3"
	"github.com/ProtonMail/gluon/async"
	"github.com/ProtonMail/go-proton-api"
)

// newAPIOptions returns a set of API options for the given parameters.
func newAPIOptions(
	apiURL string,
	version *semver.Version,
	cookieJar http.CookieJar,
	transport http.RoundTripper,
	panicHandler async.PanicHandler,
) []proton.Option {
	return defaultAPIOptions(apiURL, version, cookieJar, transport, panicHandler)
}
