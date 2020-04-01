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

// Package pmapifactory creates pmapi client instances.
package pmapifactory

import (
	"time"

	"github.com/ProtonMail/proton-bridge/internal/events"
	"github.com/ProtonMail/proton-bridge/pkg/listener"
	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
	"github.com/sirupsen/logrus"
)

func GetClientConfig(clientConfig *pmapi.ClientConfig) *pmapi.ClientConfig {
	// We set additional timeouts/thresholds for the request as a whole:
	clientConfig.Timeout = 10 * time.Minute          // Overall request timeout (~25MB / 10 mins => ~40kB/s, should be reasonable).
	clientConfig.FirstReadTimeout = 30 * time.Second // 30s to match 30s response header timeout.
	clientConfig.MinSpeed = 1 << 13                  // Enforce minimum download speed of 8kB/s.

	return clientConfig
}

func SetClientRoundTripper(cm *pmapi.ClientManager, cfg *pmapi.ClientConfig, listener listener.Listener) {
	logrus.Info("Setting dialer with pinning")

	pin := pmapi.NewDialerWithPinning(cm, cfg.AppVersion)

	pin.ReportCertIssueLocal = func() {
		listener.Emit(events.TLSCertIssue, "")
	}

	cm.SetClientRoundTripper(pin.TransportWithPinning())
}
