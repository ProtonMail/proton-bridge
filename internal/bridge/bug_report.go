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
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.

package bridge

import (
	"context"
	"io"

	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/proton-bridge/v3/internal/constants"
	"github.com/ProtonMail/proton-bridge/v3/internal/logging"
	"github.com/ProtonMail/proton-bridge/v3/internal/safe"
	"github.com/ProtonMail/proton-bridge/v3/internal/vault"
)

const (
	DefaultMaxBugReportZipSize         = 7 * 1024 * 1024
	DefaultMaxSessionCountForBugReport = 10
)

func (bridge *Bridge) ReportBug(ctx context.Context, osType, osVersion, title, description, username, email, client string, attachLogs bool) error {
	var account = username

	if info, err := bridge.QueryUserInfo(username); err == nil {
		account = info.Username
	} else if userIDs := bridge.GetUserIDs(); len(userIDs) > 0 {
		if err := bridge.vault.GetUser(userIDs[0], func(user *vault.User) {
			account = user.Username()
		}); err != nil {
			return err
		}
	}

	var attachment []proton.ReportBugAttachment

	if attachLogs {
		logsPath, err := bridge.locator.ProvideLogsPath()
		if err != nil {
			return err
		}

		buffer, err := logging.ZipLogsForBugReport(logsPath, DefaultMaxSessionCountForBugReport, DefaultMaxBugReportZipSize)
		if err != nil {
			return err
		}

		body, err := io.ReadAll(buffer)
		if err != nil {
			return err
		}

		attachment = append(attachment, proton.ReportBugAttachment{
			Name:     "logs.zip",
			Filename: "logs.zip",
			MIMEType: "application/zip",
			Body:     body,
		})
	}

	safe.Lock(func() {
		for _, user := range bridge.users {
			user.ReportBugSent()
		}
	}, bridge.usersLock)

	_, err := bridge.api.ReportBug(ctx, proton.ReportBugReq{
		OS:        osType,
		OSVersion: osVersion,

		Title:       "[Bridge] Bug - " + title,
		Description: description,

		Client:        client,
		ClientType:    proton.ClientTypeEmail,
		ClientVersion: constants.AppVersion(bridge.curVersion.Original()),

		Username: account,
		Email:    email,
	}, attachment...)

	return err
}
