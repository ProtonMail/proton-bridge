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

package pmapi

import (
	"context"
	"net/http"
	"time"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/sirupsen/logrus"
)

type Manager interface {
	NewClient(string, string, string, time.Time) Client
	NewClientWithRefresh(context.Context, string, string) (Client, *AuthRefresh, error)
	NewClientWithLogin(context.Context, string, []byte) (Client, *Auth, error)

	DownloadAndVerify(kr *crypto.KeyRing, url, sig string) ([]byte, error)
	ReportBug(context.Context, ReportBugReq) error
	SendSimpleMetric(context.Context, string, string, string) error

	SetLogging(logger *logrus.Entry, verbose bool)
	SetTransport(http.RoundTripper)
	SetCookieJar(http.CookieJar)
	SetRetryCount(int)
	AddConnectionObserver(ConnectionObserver)

	AllowProxy()
	DisallowProxy()
}
