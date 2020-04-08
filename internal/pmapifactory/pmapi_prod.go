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

	"github.com/ProtonMail/proton-bridge/internal/bridge"
	"github.com/ProtonMail/proton-bridge/internal/events"
	"github.com/ProtonMail/proton-bridge/pkg/listener"
	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
)

func New(config bridge.Configer, listener listener.Listener) bridge.PMAPIProviderFactory {
	cfg := config.GetAPIConfig()

	pin := pmapi.NewPMAPIPinning(cfg.AppVersion)
	pin.ReportCertIssueLocal = func() {
		listener.Emit(events.TLSCertIssue, "")
	}

	// This transport already has timeouts set governing the roundtrip:
	// - IdleConnTimeout:       5 * time.Minute,
	// - ExpectContinueTimeout: 500 * time.Millisecond,
	// - ResponseHeaderTimeout: 30 * time.Second,
	cfg.Transport = pin.TransportWithPinning()

	// We set additional timeouts/thresholds for the request as a whole:
	cfg.Timeout = 10 * time.Minute          // Overall request timeout (~25MB / 10 mins => ~40kB/s, should be reasonable).
	cfg.FirstReadTimeout = 30 * time.Second // 30s to match 30s response header timeout.
	cfg.MinSpeed = 1 << 13                  // Enforce minimum download speed of 8kB/s.

	return func(userID string) bridge.PMAPIProvider {
		return pmapi.NewClient(cfg, userID)
	}
}
