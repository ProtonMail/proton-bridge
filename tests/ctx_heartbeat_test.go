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

package tests

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/ProtonMail/proton-bridge/v3/internal/bridge"
	"github.com/ProtonMail/proton-bridge/v3/internal/telemetry"
	"github.com/stretchr/testify/assert"
)

type heartbeatRecorder struct {
	lock      sync.Mutex
	heartbeat telemetry.HeartbeatData
	bridge    *bridge.Bridge
	reject    bool
	assert    *assert.Assertions
}

func newHeartbeatRecorder(tb testing.TB) *heartbeatRecorder {
	return &heartbeatRecorder{
		heartbeat: telemetry.HeartbeatData{},
		bridge:    nil,
		reject:    false,
		assert:    assert.New(tb),
	}
}

func (hb *heartbeatRecorder) setBridge(bridge *bridge.Bridge) {
	hb.bridge = bridge
}

func (hb *heartbeatRecorder) GetLastHeartbeatSent() time.Time {
	if hb.bridge == nil {
		return time.Now()
	}
	return hb.bridge.GetLastHeartbeatSent()
}

func (hb *heartbeatRecorder) IsTelemetryAvailable(ctx context.Context) bool {
	if hb.bridge == nil {
		return false
	}
	return hb.bridge.IsTelemetryAvailable(ctx)
}

func (hb *heartbeatRecorder) SendHeartbeat(_ context.Context, metrics *telemetry.HeartbeatData) bool {
	if hb.bridge == nil {
		return false
	}

	if len(hb.bridge.GetUserIDs()) == 0 {
		return false
	}

	if hb.reject {
		return false
	}
	hb.lock.Lock()
	defer hb.lock.Unlock()
	hb.heartbeat = *metrics
	return true
}

func (hb *heartbeatRecorder) GetRecordedHeartbeat() telemetry.HeartbeatData {
	hb.lock.Lock()
	defer hb.lock.Unlock()

	return hb.heartbeat
}

func (hb *heartbeatRecorder) SetLastHeartbeatSent(timestamp time.Time) error {
	if hb.bridge == nil {
		return errors.New("no bridge initialized")
	}
	return hb.bridge.SetLastHeartbeatSent(timestamp)
}

func (hb *heartbeatRecorder) GetHeartbeatPeriodicInterval() time.Duration {
	return 200 * time.Millisecond
}

func (hb *heartbeatRecorder) rejectSend() {
	hb.reject = true
}
