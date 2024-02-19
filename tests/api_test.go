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

package tests

import (
	"net/http"
	"net/url"
	"os"

	"github.com/Masterminds/semver/v3"
	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/go-proton-api/server"
	"github.com/ProtonMail/proton-bridge/v3/internal/dialer"
)

type API interface {
	SetMinAppVersion(*semver.Version)
	AddCallWatcher(func(server.Call), ...string)

	GetHostURL() string
	GetDomain() string
	GetAppVersion() string

	Close()
}

func newTestAPI() API {
	if hostURL := os.Getenv("FEATURE_TEST_HOST_URL"); hostURL != "" {
		return newLiveAPI(hostURL)
	}

	return newFakeAPI()
}

type fakeAPI struct {
	*server.Server
}

func newFakeAPI() API {
	return &fakeAPI{
		Server: server.New(),
	}
}

func (api *fakeAPI) GetAppVersion() string {
	return proton.DefaultAppVersion
}

type liveAPI struct {
	*server.Server

	domain string
}

func newLiveAPI(hostURL string) API {
	url, err := url.Parse(hostURL)
	if err != nil {
		panic(err)
	}

	tr := proton.InsecureTransport()
	dialer.SetBasicTransportTimeouts(tr)
	tr.Proxy = http.ProxyFromEnvironment

	return &liveAPI{
		Server: server.New(
			server.WithProxyOrigin(hostURL),
			server.WithProxyTransport(tr),
		),
		domain: url.Hostname(),
	}
}

func (api *liveAPI) GetHostURL() string {
	return api.Server.GetProxyURL()
}

func (api *liveAPI) GetDomain() string {
	return api.domain
}

func (api *liveAPI) GetAppVersion() string {
	return os.Getenv("FEATURE_TEST_APP_VERSION")
}
