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

package user

import (
	"context"
	"encoding/json"

	"github.com/ProtonMail/proton-bridge/v3/internal/telemetry"
)

func (user *User) SendRepairTrigger(ctx context.Context) {
	if !user.IsTelemetryEnabled(ctx) {
		return
	}

	triggerData := telemetry.NewRepairTriggerData()
	data, err := json.Marshal(triggerData)
	if err != nil {
		user.log.WithError(err).Error("Failed to parse repair trigger data.")
		return
	}

	if err := user.SendTelemetry(ctx, data); err != nil {
		user.log.WithError(err).Error("Failed to send repair trigger event.")
		return
	}

	user.log.Info("Repair trigger event successfully sent.")
}

func (user *User) SendRepairDeferredTrigger(ctx context.Context) {
	if !user.IsTelemetryEnabled(ctx) {
		return
	}

	deferredTriggerData := telemetry.NewRepairDeferredTriggerData()
	data, err := json.Marshal(deferredTriggerData)
	if err != nil {
		user.log.WithError(err).Error("Failed to parse deferred repair trigger data.")
		return
	}

	if err := user.SendTelemetry(ctx, data); err != nil {
		user.log.WithError(err).Error("Failed to send deferred repair trigger event.")
		return
	}

	user.log.Info("Deferred repair trigger event successfully sent.")
}
