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

package bridge

import (
	"net/http"

	"github.com/Masterminds/semver/v3"
	"github.com/ProtonMail/gluon/async"
	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/proton-bridge/v3/internal/constants"
	"github.com/sirupsen/logrus"
)

// defaultAPIOptions returns a set of default API options for the given parameters.
func defaultAPIOptions(
	apiURL string,
	version *semver.Version,
	cookieJar http.CookieJar,
	transport http.RoundTripper,
	panicHandler async.PanicHandler,
) []proton.Option {
	return []proton.Option{
		proton.WithHostURL(apiURL),
		proton.WithAppVersion(constants.AppVersion(version.Original())),
		proton.WithCookieJar(cookieJar),
		proton.WithTransport(transport),
		proton.WithLogger(logrus.WithField("pkg", "gpa/client")),
		proton.WithPanicHandler(panicHandler),
	}
}
