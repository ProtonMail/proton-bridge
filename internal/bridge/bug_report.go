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
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.

package bridge

import (
	"context"
	"errors"
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

type ReportBugReq struct {
	OSType      string
	OSVersion   string
	Title       string
	Description string
	Username    string
	Email       string
	EmailClient string
	IncludeLogs bool
}

func (bridge *Bridge) ReportBug(ctx context.Context, report *ReportBugReq) error {
	if info, err := bridge.QueryUserInfo(report.Username); err == nil {
		report.Username = info.Username
	} else if userIDs := bridge.GetUserIDs(); len(userIDs) > 0 {
		if err := bridge.vault.GetUser(userIDs[0], func(user *vault.User) {
			report.Username = user.Username()
		}); err != nil {
			return err
		}
	}

	var attachments []proton.ReportBugAttachment
	if report.IncludeLogs {
		logs, err := bridge.CollectLogs()
		if err != nil {
			return err
		}
		attachments = append(attachments, logs)
	}

	var firstAtt proton.ReportBugAttachment
	if len(attachments) > 0 && report.IncludeLogs {
		firstAtt = attachments[0]
	}

	attachmentType := proton.AttachmentTypeSync
	if len(attachments) > 1 {
		attachmentType = proton.AttachmentTypeAsync
	}

	token, err := bridge.createTicket(ctx, report, attachmentType, firstAtt)
	if err != nil || token == "" {
		return err
	}

	safe.RLock(func() {
		for _, user := range bridge.users {
			user.ReportBugSent()
		}
	}, bridge.usersLock)

	// if we have a token we can append more attachment to the bugReport
	for i, att := range attachments {
		if i == 0 && report.IncludeLogs {
			continue
		}
		err := bridge.appendComment(ctx, token, att)
		if err != nil {
			return err
		}
	}
	return err
}

func (bridge *Bridge) CollectLogs() (proton.ReportBugAttachment, error) {
	logsPath, err := bridge.locator.ProvideLogsPath()
	if err != nil {
		return proton.ReportBugAttachment{}, err
	}

	buffer, err := logging.ZipLogsForBugReport(logsPath, DefaultMaxSessionCountForBugReport, DefaultMaxBugReportZipSize)
	if err != nil {
		return proton.ReportBugAttachment{}, err
	}

	body, err := io.ReadAll(buffer)
	if err != nil {
		return proton.ReportBugAttachment{}, err
	}

	return proton.ReportBugAttachment{
		Name:     "logs.zip",
		Filename: "logs.zip",
		MIMEType: "application/zip",
		Body:     body,
	}, nil
}

func (bridge *Bridge) createTicket(ctx context.Context, report *ReportBugReq,
	asyncAttach proton.AttachmentType, att proton.ReportBugAttachment) (string, error) {
	var attachments []proton.ReportBugAttachment
	attachments = append(attachments, att)
	res, err := bridge.api.ReportBug(ctx, proton.ReportBugReq{
		OS:        report.OSType,
		OSVersion: report.OSVersion,

		Title:       "[Bridge] Bug - " + report.Title,
		Description: report.Description,

		Client:        report.EmailClient,
		ClientType:    proton.ClientTypeEmail,
		ClientVersion: constants.AppVersion(bridge.curVersion.Original()),

		Username: report.Username,
		Email:    report.Email,

		AsyncAttachments: asyncAttach,
	}, attachments...)

	if err != nil || asyncAttach != proton.AttachmentTypeAsync {
		return "", err
	}

	if asyncAttach == proton.AttachmentTypeAsync && res.Token == nil {
		return "", errors.New("no token returns for AsyncAttachments")
	}

	return *res.Token, nil
}

func (bridge *Bridge) appendComment(ctx context.Context, token string, att proton.ReportBugAttachment) error {
	var attachments []proton.ReportBugAttachment
	attachments = append(attachments, att)
	return bridge.api.ReportBugAttachement(ctx, proton.ReportBugAttachmentReq{
		Product: proton.ClientTypeEmail,
		Body:    "Comment adding attachment: " + att.Filename,
		Token:   token,
	}, attachments...)
}
