// Copyright (c) 2023 Proton AG
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

package bridge

import (
	"context"
	"encoding/json"
	"time"

	"github.com/ProtonMail/gluon/reporter"
	"github.com/ProtonMail/proton-bridge/v3/internal/safe"
	"github.com/ProtonMail/proton-bridge/v3/internal/telemetry"
	"github.com/ProtonMail/proton-bridge/v3/internal/vault"
	"github.com/sirupsen/logrus"
)

func (bridge *Bridge) IsTelemetryAvailable() bool {
	var flag = true

	safe.RLock(func() {
		for _, user := range bridge.users {
			flag = flag && user.IsTelemetryEnabled(context.Background())
		}
	}, bridge.usersLock)

	return flag
}

func (bridge *Bridge) SendHeartbeat(heartbeat *telemetry.HeartbeatData) bool {
	data, err := json.Marshal(heartbeat)
	if err != nil {
		if err := bridge.reporter.ReportMessageWithContext("Cannot parse heartbeat data.", reporter.Context{
			"error": err,
		}); err != nil {
			logrus.WithError(err).Error("Failed to parse heartbeat data.")
		}
		return false
	}

	var sent = false

	safe.RLock(func() {
		if len(bridge.users) > 0 {
			for _, user := range bridge.users {
				if err := user.SendTelemetry(context.Background(), data); err == nil {
					sent = true
					break
				}
			}
		}
	}, bridge.usersLock)

	return sent
}

func (bridge *Bridge) GetLastHeartbeatSent() time.Time {
	return bridge.vault.GetLastHeartbeatSent()
}

func (bridge *Bridge) SetLastHeartbeatSent(timestamp time.Time) error {
	return bridge.vault.SetLastHeartbeatSent(timestamp)
}

func (bridge *Bridge) initHeartbeat() {
	safe.RLock(func() {
		var splitMode = false
		for _, user := range bridge.users {
			if user.GetAddressMode() == vault.SplitMode {
				splitMode = true
				break
			}
		}
		bridge.heartbeat.SetNbAccount(len(bridge.users))
		bridge.heartbeat.SetSplitMode(splitMode)
	}, bridge.usersLock)

	bridge.heartbeat.SetRollout(bridge.GetUpdateRollout())
	bridge.heartbeat.SetAutoStart(bridge.GetAutostart())
	bridge.heartbeat.SetAutoUpdate(bridge.GetAutoUpdate())
	bridge.heartbeat.SetBeta(bridge.GetUpdateChannel())
	bridge.heartbeat.SetDoh(bridge.GetProxyAllowed())
	bridge.heartbeat.SetShowAllMail(bridge.GetShowAllMail())
	bridge.heartbeat.SetIMAPConnectionMode(bridge.GetIMAPSSL())
	bridge.heartbeat.SetSMTPConnectionMode(bridge.GetSMTPSSL())
	bridge.heartbeat.SetIMAPPort(bridge.GetIMAPPort())
	bridge.heartbeat.SetSMTPPort(bridge.GetSMTPPort())
	bridge.heartbeat.SetCacheLocation(bridge.GetGluonCacheDir())
	if val, err := bridge.GetKeychainApp(); err != nil {
		bridge.heartbeat.SetKeyChainPref(val)
	}
	bridge.heartbeat.SetPrevVersion(bridge.GetLastVersion().String())

	bridge.heartbeat.StartSending()
}
