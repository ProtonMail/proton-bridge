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

// Package events provides names of events used by the event listener in bridge.
package events

import (
	"time"

	"github.com/ProtonMail/proton-bridge/v2/pkg/listener"
)

// Constants of events used by the event listener in bridge.
const (
	ErrorEvent                   = "error"
	CredentialsErrorEvent        = "credentialsError"
	CloseConnectionEvent         = "closeConnection"
	LogoutEvent                  = "logout"
	AddressChangedEvent          = "addressChanged"
	AddressChangedLogoutEvent    = "addressChangedLogout"
	UserRefreshEvent             = "userRefresh"
	RestartBridgeEvent           = "restartBridge"
	InternetConnChangedEvent     = "internetChanged"
	InternetOff                  = "internetOff"
	InternetOn                   = "internetOn"
	SecondInstanceEvent          = "secondInstance"
	OutgoingNoEncEvent           = "outgoingNoEncryption"
	NoActiveKeyForRecipientEvent = "noActiveKeyForRecipient"
	UpgradeApplicationEvent      = "upgradeApplication"
	TLSCertIssue                 = "tlsCertPinningIssue"
	UserChangeDone               = "QMLUserChangedDone"

	// LogoutEventTimeout is the minimum time to permit between logout events being sent.
	LogoutEventTimeout = 3 * time.Minute
)

// SetupEvents specific to event type and data.
func SetupEvents(listener listener.Listener) {
	listener.SetLimit(LogoutEvent, LogoutEventTimeout)
	listener.SetBuffer(ErrorEvent)
	listener.SetBuffer(CredentialsErrorEvent)
	listener.SetBuffer(InternetConnChangedEvent)
	listener.SetBuffer(UpgradeApplicationEvent)
	listener.SetBuffer(TLSCertIssue)
	listener.SetBuffer(UserRefreshEvent)
	listener.Book(UserChangeDone)
}
