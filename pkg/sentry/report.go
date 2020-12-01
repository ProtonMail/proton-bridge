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
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	log "github.com/sirupsen/logrus"
)

var (
	isPanicHandlerRegexp = regexp.MustCompile(`^ReportSentryCrash|^(\(\*PanicHandler\)\.)?HandlePanic`) //nolint[gochecknoglobals]
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

	var reportID string
	sentry.WithScope(func(scope *sentry.Scope) {
		scope.SetTags(tags)
		eventID := sentry.CaptureException(reportErr)
		if eventID != nil {
			reportID = string(*eventID)
		}
	})

	if !sentry.Flush(time.Second * 10) {
		log.WithField("error", reportErr).Error("Failed to report sentry error")
		return errors.New("failed to report sentry error")
	}

	log.WithField("error", reportErr).WithField("id", reportID).Warn("Sentry error reported")
	return
}

// EnhanceSentryEvent swaps type with value and removes panic handlers from the stacktrace.
func EnhanceSentryEvent(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
	for idx, exception := range event.Exception {
		exception.Type, exception.Value = exception.Value, exception.Type
		if exception.Stacktrace != nil {
			exception.Stacktrace.Frames = filterOutPanicHandlers(exception.Stacktrace.Frames)
		}
		event.Exception[idx] = exception
	}
	return event
}

func filterOutPanicHandlers(frames []sentry.Frame) []sentry.Frame {
	idx := 0
	for _, frame := range frames {
		if strings.HasPrefix(frame.Module, "github.com/ProtonMail/proton-bridge") &&
			isPanicHandlerRegexp.MatchString(frame.Function) {
			break
		}
		idx++
	}
	return frames[:idx]
}
