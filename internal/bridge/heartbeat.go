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

package bridge

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/ProtonMail/gluon/async"
	"github.com/ProtonMail/gluon/reporter"
	"github.com/ProtonMail/proton-bridge/v3/internal/safe"
	"github.com/ProtonMail/proton-bridge/v3/internal/telemetry"
	"github.com/ProtonMail/proton-bridge/v3/internal/vault"
	"github.com/sirupsen/logrus"
)

const HeartbeatCheckInterval = time.Hour

type heartBeatState struct {
	task *async.Group
	telemetry.Heartbeat
	taskLock     sync.Mutex
	taskStarted  bool
	taskInterval time.Duration
}

func newHeartBeatState(ctx context.Context, panicHandler async.PanicHandler) *heartBeatState {
	return &heartBeatState{
		task: async.NewGroup(ctx, panicHandler),
	}
}

func (h *heartBeatState) init(bridge *Bridge, manager telemetry.HeartbeatManager) {
	h.Heartbeat = telemetry.NewHeartbeat(manager, 1143, 1025, bridge.GetGluonCacheDir(), bridge.keychains.GetDefaultHelper())
	h.taskInterval = manager.GetHeartbeatPeriodicInterval()
	h.SetRollout(bridge.GetUpdateRollout())
	h.SetAutoStart(bridge.GetAutostart())
	h.SetAutoUpdate(bridge.GetAutoUpdate())
	h.SetBeta(bridge.GetUpdateChannel())
	h.SetDoh(bridge.GetProxyAllowed())
	h.SetShowAllMail(bridge.GetShowAllMail())
	h.SetIMAPConnectionMode(bridge.GetIMAPSSL())
	h.SetSMTPConnectionMode(bridge.GetSMTPSSL())
	h.SetIMAPPort(bridge.GetIMAPPort())
	h.SetSMTPPort(bridge.GetSMTPPort())
	h.SetCacheLocation(bridge.GetGluonCacheDir())
	if val, err := bridge.GetKeychainApp(); err != nil {
		h.SetKeyChainPref(val)
	} else {
		h.SetKeyChainPref(bridge.keychains.GetDefaultHelper())
	}
	h.SetPrevVersion(bridge.GetLastVersion().String())

	safe.RLock(func() {
		var splitMode = false
		for _, user := range bridge.users {
			if user.GetAddressMode() == vault.SplitMode {
				splitMode = true
				break
			}
		}
		var nbAccount = len(bridge.users)
		h.SetNbAccount(nbAccount)
		h.SetSplitMode(splitMode)

		// Do not try to send if there is no user yet.
		if nbAccount > 0 {
			defer h.start()
		}
	}, bridge.usersLock)
}

func (h *heartBeatState) start() {
	h.taskLock.Lock()
	defer h.taskLock.Unlock()
	if h.taskStarted {
		return
	}

	h.taskStarted = true

	h.task.PeriodicOrTrigger(h.taskInterval, 0, func(ctx context.Context) {
		logrus.WithField("pkg", "bridge/heartbeat").Debug("Checking for heartbeat")

		h.TrySending(ctx)
	})
}

func (h *heartBeatState) stop() {
	h.taskLock.Lock()
	defer h.taskLock.Unlock()
	if !h.taskStarted {
		return
	}

	h.task.CancelAndWait()
	h.taskStarted = false
}

func (bridge *Bridge) IsTelemetryAvailable(ctx context.Context) bool {
	var flag = true
	if bridge.GetTelemetryDisabled() {
		return false
	}

	safe.RLock(func() {
		for _, user := range bridge.users {
			flag = flag && user.IsTelemetryEnabled(ctx)
		}
	}, bridge.usersLock)

	return flag
}

func (bridge *Bridge) SendHeartbeat(ctx context.Context, heartbeat *telemetry.HeartbeatData) bool {
	data, err := json.Marshal(heartbeat)
	if err != nil {
		if err := bridge.reporter.ReportMessageWithContext("Cannot parse heartbeat data.", reporter.Context{
			"error": err,
		}); err != nil {
			logrus.WithField("pkg", "bridge/heartbeat").WithError(err).Error("Failed to parse heartbeat data.")
		}
		return false
	}

	var sent = false

	safe.RLock(func() {
		for _, user := range bridge.users {
			if err := user.SendTelemetry(ctx, data); err == nil {
				sent = true
				break
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

func (bridge *Bridge) GetHeartbeatPeriodicInterval() time.Duration {
	return HeartbeatCheckInterval
}
