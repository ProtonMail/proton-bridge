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

// +build !pmapi_qa

package pmapi

import (
	"net/http"

	"github.com/ProtonMail/proton-bridge/internal/events"
	"github.com/ProtonMail/proton-bridge/pkg/listener"
)

func GetRoundTripper(cm *ClientManager, listener listener.Listener) http.RoundTripper {
	// We use a TLS dialer.
	basicDialer := NewBasicTLSDialer()

	// We wrap the TLS dialer in a layer which enforces connections to trusted servers.
	pinningDialer := NewPinningTLSDialer(basicDialer)

	// We want any pin mismatches to be communicated back to bridge GUI and reported.
	pinningDialer.SetTLSIssueNotifier(func() { listener.Emit(events.TLSCertIssue, "") })
	pinningDialer.EnableRemoteTLSIssueReporting(cm)

	// We wrap the pinning dialer in a layer which adds "alternative routing" feature.
	proxyDialer := NewProxyTLSDialer(pinningDialer, cm)

	return CreateTransportWithDialer(proxyDialer)
}
