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
// along with Proton Mail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package bridge

import (
	"net/http"

	"github.com/Masterminds/semver/v3"
	"github.com/ProtonMail/proton-bridge/v2/internal/constants"
	"github.com/sirupsen/logrus"
	"gitlab.protontech.ch/go/liteapi"
)

// defaultAPIOptions returns a set of default API options for the given parameters.
func defaultAPIOptions(
	apiURL string,
	version *semver.Version,
	cookieJar http.CookieJar,
	transport http.RoundTripper,
	poolSize int,
) []liteapi.Option {
	return []liteapi.Option{
		liteapi.WithHostURL(apiURL),
		liteapi.WithAppVersion(constants.AppVersion(version.Original())),
		liteapi.WithCookieJar(cookieJar),
		liteapi.WithTransport(transport),
		liteapi.WithAttPoolSize(poolSize),
		liteapi.WithLogger(logrus.StandardLogger()),
	}
}
