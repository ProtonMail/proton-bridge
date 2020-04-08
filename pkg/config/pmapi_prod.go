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

// +build pmapi_prod

package config

import (
	"net/http"
	"strings"
	"time"

	"github.com/ProtonMail/proton-bridge/internal/events"
	"github.com/ProtonMail/proton-bridge/pkg/listener"
	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
)

func (c *Config) GetAPIConfig() *pmapi.ClientConfig {
	return &pmapi.ClientConfig{
		AppVersion:       strings.Title(c.appName) + "_" + c.version,
		ClientID:         c.appName,
		Timeout:          10 * time.Minute, // Overall request timeout (~25MB / 10 mins => ~40kB/s, should be reasonable).
		FirstReadTimeout: 30 * time.Second, // 30s to match 30s response header timeout.
		MinSpeed:         1 << 13,          // Enforce minimum download speed of 8kB/s.
	}
}

func (c *Config) GetRoundTripper(cm *pmapi.ClientManager, listener listener.Listener) http.RoundTripper {
	pin := pmapi.NewDialerWithPinning(cm, c.GetAPIConfig().AppVersion)

	pin.ReportCertIssueLocal = func() {
		listener.Emit(events.TLSCertIssue, "")
	}

	return pin.TransportWithPinning()
}
