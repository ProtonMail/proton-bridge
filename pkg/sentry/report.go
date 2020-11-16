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

package sentry

import (
	"errors"
	"runtime"
	"time"

	"github.com/getsentry/sentry-go"
	log "github.com/sirupsen/logrus"
)

// ReportSentryCrash reports a sentry crash.
func ReportSentryCrash(clientID, appVersion, userAgent string, reportErr error) (err error) {
	if reportErr == nil {
		return
	}

	tags := map[string]string{
		"OS":        runtime.GOOS,
		"Client":    clientID,
		"Version":   appVersion,
		"UserAgent": userAgent,
		"UserID":    "",
	}

	sentry.WithScope(func(scope *sentry.Scope) {
		scope.SetTags(tags)
		sentry.CaptureException(reportErr)
	})

	if !sentry.Flush(time.Second * 10) {
		log.WithField("error", reportErr).Error("failed to report sentry error")
		return errors.New("failed to report sentry error")
	}

	log.WithField("error", reportErr).Warn("reported sentry error")
	return
}
