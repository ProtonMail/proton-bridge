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

// Package events provides names of events used by the event listener in bridge.
package events

import (
	"time"

	"github.com/ProtonMail/proton-bridge/pkg/listener"
)

// Constants of events used by the event listener in bridge.
const (
	ErrorEvent                   = "error"
	CloseConnectionEvent         = "closeConnection"
	LogoutEvent                  = "logout"
	AddressChangedEvent          = "addressChanged"
	AddressChangedLogoutEvent    = "addressChangedLogout"
	UserRefreshEvent             = "userRefresh"
	RestartBridgeEvent           = "restartBridge"
	InternetOffEvent             = "internetOff"
	InternetOnEvent              = "internetOn"
	SecondInstanceEvent          = "secondInstance"
	OutgoingNoEncEvent           = "outgoingNoEncryption"
	NoActiveKeyForRecipientEvent = "noActiveKeyForRecipient"
	UpgradeApplicationEvent      = "upgradeApplication"
	TLSCertIssue                 = "tlsCertPinningIssue"
	IMAPTLSBadCert               = "imapTLSBadCert"

	// LogoutEventTimeout is the minimum time to permit between logout events being sent.
	LogoutEventTimeout = 3 * time.Minute
)

// SetupEvents specific to event type and data.
func SetupEvents(listener listener.Listener) {
	listener.SetLimit(LogoutEvent, LogoutEventTimeout)
	listener.SetBuffer(TLSCertIssue)
	listener.SetBuffer(ErrorEvent)
}
