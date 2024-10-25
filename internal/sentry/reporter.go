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

package sentry

import (
	"errors"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/ProtonMail/gluon/reporter"
	"github.com/ProtonMail/proton-bridge/v3/internal/constants"
	"github.com/ProtonMail/proton-bridge/v3/pkg/algo"
	"github.com/ProtonMail/proton-bridge/v3/pkg/restarter"
	"github.com/elastic/go-sysinfo"
	"github.com/getsentry/sentry-go"
	"github.com/sirupsen/logrus"
)

const hostNotDetectedField = "not-detected"

var skippedFunctions = []string{} //nolint:gochecknoglobals

func init() { //nolint:gochecknoinits
	sentrySyncTransport := sentry.NewHTTPSyncTransport()
	sentrySyncTransport.Timeout = time.Second * 3
	appVersion := constants.Version
	version, _ := semver.NewVersion(appVersion)
	if version != nil {
		appVersion = version.Original()
	}

	options := sentry.ClientOptions{
		Dsn:            constants.DSNSentry,
		Release:        constants.AppVersion(appVersion),
		BeforeSend:     EnhanceSentryEvent,
		Transport:      sentrySyncTransport,
		ServerName:     GetProtectedHostname(),
		Environment:    constants.BuildEnv,
		MaxBreadcrumbs: 50,
	}

	if err := sentry.Init(options); err != nil {
		logrus.WithError(err).Error("Failed to initialize sentry options")
	}

	sentry.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetFingerprint([]string{"{{ default }}"})
		scope.SetUser(sentry.User{ID: GetProtectedHostname()})
	})

	sentry.Logger = log.New(
		logrus.WithField("pkg", "sentry-go").WriterLevel(logrus.WarnLevel),
		"", 0,
	)
}

type hostInfoData struct {
	hostArch    string
	hostName    string
	hostVersion string
	hostBuild   string
}

func newHostInfoData() hostInfoData {
	return hostInfoData{
		hostArch:    hostNotDetectedField,
		hostName:    hostNotDetectedField,
		hostVersion: hostNotDetectedField,
		hostBuild:   hostNotDetectedField,
	}
}

type Reporter struct {
	appName    string
	appVersion string
	identifier Identifier
	hostInfo   hostInfoData
}

type Identifier interface {
	GetUserAgent() string
}

func getHostInfo() hostInfoData {
	data := newHostInfoData()

	host, err := sysinfo.Host()
	if err != nil {
		return data
	}

	data.hostArch = getHostArch(host)
	osInfo := host.Info().OS
	data.hostName = osInfo.Name
	data.hostVersion = osInfo.Version
	data.hostBuild = osInfo.Build

	return data
}

func GetProtectedHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "Unknown"
	}
	return algo.HashBase64SHA256(hostname)
}

func GetTimeZone() string {
	zone, offset := time.Now().Zone()
	return fmt.Sprintf("%s%+d", zone, offset/3600)
}

// NewReporter creates new sentry reporter with appName and appVersion to report.
func NewReporter(appName string, identifier Identifier) *Reporter {
	return &Reporter{
		appName:    appName,
		appVersion: constants.Revision,
		identifier: identifier,
		hostInfo:   getHostInfo(),
	}
}

func (r *Reporter) ReportException(i interface{}) error {
	SkipDuringUnwind()
	return r.ReportExceptionWithContext(i, map[string]interface{}{
		"build": constants.BuildTime,
		"crash": os.Getenv(restarter.BridgeCrashCount),
	})
}

func (r *Reporter) ReportMessage(msg string) error {
	SkipDuringUnwind()
	return r.ReportMessageWithContext(msg, make(map[string]interface{}))
}

func (r *Reporter) ReportExceptionWithContext(i interface{}, context map[string]interface{}) error {
	SkipDuringUnwind()

	err := fmt.Errorf("recover: %v", i)
	return r.scopedReport(context, func() {
		SkipDuringUnwind()
		if eventID := sentry.CaptureException(err); eventID != nil {
			logrus.WithError(err).
				WithField("reportID", *eventID).
				Warn("Captured exception")
		}
	})
}

func (r *Reporter) ReportMessageWithContext(msg string, context map[string]interface{}) error {
	SkipDuringUnwind()
	return r.scopedReport(context, func() {
		SkipDuringUnwind()
		if eventID := sentry.CaptureMessage(msg); eventID != nil {
			logrus.WithField("message", msg).
				WithField("reportID", *eventID).
				Warn("Captured message")
		}
	})
}

// Report reports a sentry crash with stacktrace from all goroutines.
func (r *Reporter) scopedReport(context map[string]interface{}, doReport func()) error {
	SkipDuringUnwind()

	if os.Getenv("PROTONMAIL_ENV") == "dev" {
		return nil
	}

	tags := map[string]string{
		"OS":          runtime.GOOS,
		"Client":      r.appName,
		"Version":     r.appVersion,
		"UserAgent":   r.identifier.GetUserAgent(),
		"HostArch":    r.hostInfo.hostArch,
		"HostName":    r.hostInfo.hostName,
		"HostVersion": r.hostInfo.hostVersion,
		"HostBuild":   r.hostInfo.hostBuild,
	}

	sentry.WithScope(func(scope *sentry.Scope) {
		SkipDuringUnwind()
		scope.SetTags(tags)
		if len(context) != 0 {
			scope.SetContexts(
				map[string]sentry.Context{"bridge": contextToString(context)},
			)
		}
		doReport()
	})

	if !sentry.Flush(time.Second * 10) {
		return errors.New("failed to report sentry error")
	}

	return nil
}

func ReportError(r reporter.Reporter, msg string, err error) {
	if rerr := r.ReportMessageWithContext(msg, reporter.Context{
		"error": err.Error(),
	}); rerr != nil {
		logrus.WithError(rerr).WithField("msg", msg).Error("Failed to send report")
	}
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
func EnhanceSentryEvent(event *sentry.Event, _ *sentry.EventHint) *sentry.Event {
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

func Flush(maxWaiTime time.Duration) {
	sentry.Flush(maxWaiTime)
}

func contextToString(context sentry.Context) sentry.Context {
	res := make(sentry.Context)

	for k, v := range context {
		res[k] = fmt.Sprintf("%v", v)
	}

	return res
}
