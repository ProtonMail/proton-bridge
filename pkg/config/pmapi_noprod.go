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

// +build !pmapi_prod

package config

import (
	"net/http"
	"strings"

	"github.com/ProtonMail/proton-bridge/pkg/listener"
	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
)

func (c *Config) GetAPIConfig() *pmapi.ClientConfig {
	return &pmapi.ClientConfig{
		AppVersion: c.getAPIOS() + strings.Title(c.appName) + "_" + c.version,
		ClientID:   c.appName,
	}
}

func SetClientRoundTripper(_ *pmapi.ClientManager, _ *pmapi.ClientConfig, _ listener.Listener) {
	// Use the default roundtripper; do nothing.
}

func (c *Config) GetRoundTripper(_ *pmapi.ClientManager, _ listener.Listener) http.RoundTripper {
	return http.DefaultTransport
}
