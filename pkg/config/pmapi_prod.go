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
		AppVersion:        c.getAPIOS() + strings.Title(c.appName) + "_" + c.version,
		ClientID:          c.appName,
		Timeout:           25 * time.Minute, // Overall request timeout (~25MB / 25 mins => ~16kB/s, should be reasonable).
		FirstReadTimeout:  30 * time.Second, // 30s to match 30s response header timeout.
		MinBytesPerSecond: 1 << 10,          // Enforce minimum download speed of 1kB/s.
	}
}

func (c *Config) GetRoundTripper(cm *pmapi.ClientManager, listener listener.Listener) http.RoundTripper {
	// We use a TLS dialer.
	basicDialer := pmapi.NewBasicTLSDialer()

	// We wrap the TLS dialer in a layer which enforces connections to trusted servers.
	pinningDialer := pmapi.NewPinningTLSDialer(basicDialer)

	// We want any pin mismatches to be communicated back to bridge GUI and reported.
	pinningDialer.SetTLSIssueNotifier(func() { listener.Emit(events.TLSCertIssue, "") })
	pinningDialer.EnableRemoteTLSIssueReporting(c.GetAPIConfig().AppVersion, c.GetAPIConfig().UserAgent)

	// We wrap the pinning dialer in a layer which adds "alternative routing" feature.
	proxyDialer := pmapi.NewProxyTLSDialer(pinningDialer, cm)

	return pmapi.CreateTransportWithDialer(proxyDialer)
}
