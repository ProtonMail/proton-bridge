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

package user

import (
	"context"
	"encoding/json"

	"github.com/ProtonMail/gluon/reporter"
	"github.com/ProtonMail/proton-bridge/v3/internal/configstatus"
)

func (user *User) SendConfigStatusSuccess() {
	if !user.telemetryManager.IsTelemetryAvailable() {
		return
	}

	if !user.configStatus.IsPending() {
		return
	}

	var builder configstatus.ConfigSuccessBuilder
	success := builder.New(user.configStatus.Data)
	data, err := json.Marshal(success)
	if err != nil {
		if err := user.reporter.ReportMessageWithContext("Cannot parse config_success data.", reporter.Context{
			"error": err,
		}); err != nil {
			user.log.WithError(err).Error("Failed to parse config_success data.")
		}
	}

	if err := user.SendTelemetry(context.Background(), data); err == nil {
		if err := user.configStatus.ApplySuccess(); err != nil {
			user.log.WithError(err).Error("Failed to ApplySuccess on config_status.")
		}
	}
}

func (user *User) SendConfigStatusAbort() {
}

func (user *User) SendConfigStatusRecovery() {
}

func (user *User) SendConfigStatusProgress() {
}
