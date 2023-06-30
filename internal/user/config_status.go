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
	if user.configStatus.IsFromFailure() {
		user.SendConfigStatusRecovery()
		return
	}
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
			user.log.WithError(err).Error("Failed to report config_success data parsing error.")
		}
		return
	}

	if err := user.SendTelemetry(context.Background(), data); err == nil {
		user.log.Info("Configuration Status Success event sent.")
		if err := user.configStatus.ApplySuccess(); err != nil {
			user.log.WithError(err).Error("Failed to ApplySuccess on config_status.")
		}
	}
}

func (user *User) SendConfigStatusAbort(withTelemetry bool) {
	if err := user.configStatus.Remove(); err != nil {
		user.log.WithError(err).Error("Failed to remove config_status file.")
	}

	if !user.configStatus.IsPending() {
		return
	}
	if !withTelemetry {
		return
	}
	var builder configstatus.ConfigAbortBuilder
	abort := builder.New(user.configStatus.Data)
	data, err := json.Marshal(abort)
	if err != nil {
		if err := user.reporter.ReportMessageWithContext("Cannot parse config_abort data.", reporter.Context{
			"error": err,
		}); err != nil {
			user.log.WithError(err).Error("Failed to report config_abort data parsing error.")
		}
		return
	}

	if err := user.SendTelemetry(context.Background(), data); err == nil {
		user.log.Info("Configuration Status Abort event sent.")
	}
}

func (user *User) SendConfigStatusRecovery() {
	if !user.configStatus.IsFromFailure() {
		user.SendConfigStatusSuccess()
		return
	}
	if !user.telemetryManager.IsTelemetryAvailable() {
		return
	}
	if !user.configStatus.IsPending() {
		return
	}

	var builder configstatus.ConfigRecoveryBuilder
	success := builder.New(user.configStatus.Data)
	data, err := json.Marshal(success)
	if err != nil {
		if err := user.reporter.ReportMessageWithContext("Cannot parse config_recovery data.", reporter.Context{
			"error": err,
		}); err != nil {
			user.log.WithError(err).Error("Failed to report config_recovery data parsing error.")
		}
		return
	}

	if err := user.SendTelemetry(context.Background(), data); err == nil {
		user.log.Info("Configuration Status Recovery event sent.")
		if err := user.configStatus.ApplySuccess(); err != nil {
			user.log.WithError(err).Error("Failed to ApplySuccess on config_status.")
		}
	}
}

func (user *User) SendConfigStatusProgress() {
	if !user.telemetryManager.IsTelemetryAvailable() {
		return
	}
	if !user.configStatus.IsPending() {
		return
	}

	var builder configstatus.ConfigProgressBuilder
	progress := builder.New(user.configStatus.Data)
	if progress.Values.NbDaySinceLast == 0 || progress.Values.NbDay == 0 {
		return
	}

	data, err := json.Marshal(progress)
	if err != nil {
		if err := user.reporter.ReportMessageWithContext("Cannot parse config_progress data.", reporter.Context{
			"error": err,
		}); err != nil {
			user.log.WithError(err).Error("Failed to report config_progress data parsing error.")
		}
		return
	}

	if err := user.SendTelemetry(context.Background(), data); err == nil {
		user.log.Info("Configuration Status Progress event sent.")
		if err := user.configStatus.ApplyProgress(); err != nil {
			user.log.WithError(err).Error("Failed to ApplyProgress on config_status.")
		}
	}
}

func (user *User) ReportConfigStatusFailure(errDetails string) {
	if user.configStatus.IsPending() {
		return
	}

	if err := user.configStatus.ApplyFailure(errDetails); err != nil {
		user.log.WithError(err).Error("Failed to ApplyFailure on config_status.")
	} else {
		user.log.Info("Configuration Status is back to Pending due to Failure.")
	}
}

func (user *User) ReportBugClicked() {
	if !user.configStatus.IsPending() {
		return
	}

	if err := user.configStatus.ReportClicked(); err != nil {
		user.log.WithError(err).Error("Failed to log ReportClicked in config_status.")
	}
}

func (user *User) ReportBugSent() {
	if !user.configStatus.IsPending() {
		return
	}

	if err := user.configStatus.ReportSent(); err != nil {
		user.log.WithError(err).Error("Failed to log ReportSent in config_status.")
	}
}

func (user *User) AutoconfigUsed(client string) {
	if !user.configStatus.IsPending() {
		return
	}

	if err := user.configStatus.AutoconfigUsed(client); err != nil {
		user.log.WithError(err).Error("Failed to log Autoconf in config_status.")
	}
}

func (user *User) KBArticleOpened(article string) {
	if !user.configStatus.IsPending() {
		return
	}

	var kb_articles_to_track = [...]string{
		"https://proton.me/support/bridge",
		"https://proton.me/support/protonmail-bridge-clients-apple-mail",
		"https://proton.me/support/protonmail-bridge-clients-macos-outlook-2019",
		"https://proton.me/support/protonmail-bridge-clients-windows-outlook-2019",
		"https://proton.me/support/protonmail-bridge-clients-windows-thunderbird",
		"https://proton.me/support/protonmail-bridge-configure-client",
	}

	for id, url := range kb_articles_to_track {
		if url == article {
			if err := user.configStatus.RecordLinkClicked(uint(id)); err != nil {
				user.log.WithError(err).Error("Failed to log LinkClicked in config_status.")
			}
			return
		}
	}
	user.log.WithField("article", article).Error("Failed to find KB article id.")
}
