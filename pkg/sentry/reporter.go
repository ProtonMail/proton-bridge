// Copyright (c) 2021 Proton Technologies AG
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
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/ProtonMail/proton-bridge/internal/constants"
	"github.com/getsentry/sentry-go"
	"github.com/sirupsen/logrus"
)

var skippedFunctions = []string{} //nolint[gochecknoglobals]

func init() { // nolint[noinit]
	if err := sentry.Init(sentry.ClientOptions{
		Dsn:        constants.DSNSentry,
		Release:    constants.Revision,
		BeforeSend: EnhanceSentryEvent,
	}); err != nil {
		logrus.WithError(err).Error("Failed to initialize sentry options")
	}

	sentry.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetFingerprint([]string{"{{ default }}"})
	})
}

type userAgentProvider interface {
	GetUserAgent() string
}

type Reporter struct {
	appName    string
	appVersion string
	uap        userAgentProvider
}

// NewReporter creates new sentry reporter with appName and appVersion to report.
func NewReporter(appName, appVersion string) *Reporter {
	return &Reporter{
		appName:    appName,
		appVersion: appVersion,
	}
}

func (r *Reporter) SetUserAgentProvider(uap userAgentProvider) {
	r.uap = uap
}

// Report reports a sentry crash with stacktrace from all goroutines.
func (r *Reporter) Report(i interface{}) (err error) {
	SkipDuringUnwind()

	if os.Getenv("PROTONMAIL_ENV") == "dev" {
		return nil
	}

	// In case clientManager is not yet created we can get at least OS string.
	var userAgent string
	if r.uap != nil {
		userAgent = r.uap.GetUserAgent()
	} else {
		userAgent = runtime.GOOS
	}

	reportErr := fmt.Errorf("recover: %v", i)

	tags := map[string]string{
		"OS":        runtime.GOOS,
		"Client":    r.appName,
		"Version":   r.appVersion,
		"UserAgent": userAgent,
		"UserID":    "",
	}

	var reportID string
	sentry.WithScope(func(scope *sentry.Scope) {
		SkipDuringUnwind()
		scope.SetTags(tags)
		if eventID := sentry.CaptureException(reportErr); eventID != nil {
			reportID = string(*eventID)
		}
	})

	if !sentry.Flush(time.Second * 10) {
		return errors.New("failed to report sentry error")
	}

	logrus.WithField("error", reportErr).WithField("id", reportID).Warn("Sentry error reported")

	return nil
}

// SkipDuringUnwind removes caller from the traceback.
func SkipDuringUnwind() {
	pcs := make([]uintptr, 2)
	n := runtime.Callers(2, pcs)
	if n == 0 {
		return
	}

	frames := runtime.CallersFrames(pcs)
	frame, _ := frames.Next()
	if isFunctionFilteredOut(frame.Function) {
		return
	}

	skippedFunctions = append(skippedFunctions, frame.Function)
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
	newFrames := []sentry.Frame{}
	for _, frame := range frames {
		// Sentry splits runtime.Frame.Function into Module and Function.
		function := frame.Module + "." + frame.Function
		if !isFunctionFilteredOut(function) {
			newFrames = append(newFrames, frame)
		}
	}
	return newFrames
}

func isFunctionFilteredOut(function string) bool {
	for _, skipFunction := range skippedFunctions {
		if function == skipFunction {
			return true
		}
	}
	return false
}
