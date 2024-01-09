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

package logging

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/bradenaw/juniper/xslices"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

const (
	// DefaultMaxLogFileSize defines the maximum log size we should permit: 5 MB
	//
	// The Zendesk limit for an attachment is 50MB and this is what will
	// be allowed via the API. However, if that fails for some reason, the
	// fallback is sending the report via email, which has a limit of 10mb
	// total or 7MB per file.
	DefaultMaxLogFileSize = 5 * 1024 * 1024
)

type AppName string

const (
	BridgeShortAppName   AppName = "bri"
	LauncherShortAppName AppName = "lau"
	GUIShortAppName      AppName = "gui"
)

type coloredStdOutHook struct {
	formatter logrus.Formatter
}

func newColoredStdOutHook() *coloredStdOutHook {
	return &coloredStdOutHook{
		formatter: &logrus.TextFormatter{
			ForceColors:     true,
			FullTimestamp:   true,
			TimestampFormat: time.StampMilli,
		},
	}
}

func (cs *coloredStdOutHook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,
	}
}

func (cs *coloredStdOutHook) Fire(entry *logrus.Entry) error {
	bytes, err := cs.formatter.Format(entry)
	if err != nil {
		return err
	}

	if _, err := os.Stdout.Write(bytes); err != nil {
		return err
	}

	return nil
}

// Init Initialize logging. Log files are rotated when their size exceeds rotationSize. if pruningSize >= 0, pruning occurs using
// the default pruning algorithm.
func Init(logsPath string, sessionID SessionID, appName AppName, rotationSize, pruningSize int64, level string) (io.Closer, error) {
	logrus.SetFormatter(&logrus.TextFormatter{
		DisableColors:    true,
		ForceQuote:       true,
		FullTimestamp:    true,
		QuoteEmptyFields: true,
		TimestampFormat:  "2006-01-02 15:04:05.000",
	})

	logrus.AddHook(newColoredStdOutHook())

	rotator, err := NewDefaultRotator(logsPath, sessionID, appName, rotationSize, pruningSize)
	if err != nil {
		return nil, err
	}

	logrus.SetOutput(rotator)

	return rotator, setLevel(level)
}

// Close closes the log file. if closer is nil, no error is reported.
func Close(closer io.Closer) error {
	if closer == nil {
		return nil
	}

	logrus.SetOutput(os.Stdout)
	return closer.Close()
}

// ZipLogsForBugReport returns an archive containing the logs for bug report.
func ZipLogsForBugReport(logsPath string, maxSessionCount int, maxZipSize int64) (*bytes.Buffer, error) {
	paths, err := getOrderedLogFileListForBugReport(logsPath, maxSessionCount)
	if err != nil {
		return nil, err
	}

	buffer, _, err := zipFilesWithMaxSize(paths, maxZipSize)
	return buffer, err
}

// getOrderedLogFileListForBugReport returns the ordered list of log file paths to include in the user triggered bug reports. Only the last
// maxSessionCount sessions are included. Priorities:
// - session in chronologically descending order.
// - for each session: last 2 bridge logs, first bridge log, gui logs, launcher logs, all other bridge logs.
func getOrderedLogFileListForBugReport(logsPath string, maxSessionCount int) ([]string, error) {
	sessionInfoList, err := buildSessionInfoList(logsPath)
	if err != nil {
		return nil, err
	}

	sortedSessions := maps.Values(sessionInfoList)
	slices.SortFunc(sortedSessions, func(lhs, rhs *sessionInfo) bool { return lhs.sessionID > rhs.sessionID })
	count := len(sortedSessions)
	if count > maxSessionCount {
		sortedSessions = sortedSessions[:maxSessionCount]
	}

	filePathFunc := func(logFileInfo logFileInfo) string { return filepath.Join(logsPath, logFileInfo.filename) }

	var result []string
	for _, session := range sortedSessions {
		bridgeLogCount := len(session.bridgeLogs)
		if bridgeLogCount > 0 {
			result = append(result, filepath.Join(logsPath, session.bridgeLogs[bridgeLogCount-1].filename))
		}
		if bridgeLogCount > 1 {
			result = append(result, filepath.Join(logsPath, session.bridgeLogs[bridgeLogCount-2].filename))
		}
		if bridgeLogCount > 2 {
			result = append(result, filepath.Join(logsPath, session.bridgeLogs[0].filename))
		}
		if len(session.guiLogs) > 0 {
			result = append(result, xslices.Map(session.guiLogs, filePathFunc)...)
		}
		if len(session.launcherLogs) > 0 {
			result = append(result, xslices.Map(session.launcherLogs, filePathFunc)...)
		}
		if bridgeLogCount > 3 {
			result = append(result, xslices.Map(session.bridgeLogs[1:bridgeLogCount-2], filePathFunc)...)
		}
	}

	return result, nil
}

// setLevel will change the level of logging and in case of Debug or Trace
// level it will also prevent from writing to file. Setting level to Info or
// higher will not set writing to file again if it was previously cancelled by
// Debug or Trace.
func setLevel(level string) error {
	if level == "" {
		level = "debug"
	}

	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		return err
	}

	logrus.SetLevel(logLevel)

	// The hook to print panic, fatal and error to stderr is always
	// added. We want to avoid log duplicates by replacing all hooks.
	if logrus.GetLevel() == logrus.TraceLevel {
		_ = logrus.StandardLogger().ReplaceHooks(logrus.LevelHooks{})
		logrus.SetOutput(os.Stderr)
		logrus.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: time.StampMilli,
		})
	}

	return nil
}

func getLogSessionID(filename string) (SessionID, error) {
	re := regexp.MustCompile(`^(?P<sessionID>\d{8}_\d{9})_.*\.log$`)

	match := re.FindStringSubmatch(filename)

	errInvalidFileName := errors.New("log file name is invalid")
	if len(match) == 0 {
		logrus.WithField("filename", filename).Warn("Could not parse log filename")
		return "", errInvalidFileName
	}

	index := re.SubexpIndex("sessionID")
	if index < 0 {
		logrus.WithField("filename", filename).Warn("Could not parse log filename")
		return "", errInvalidFileName
	}

	return SessionID(match[index]), nil
}

func getLogTime(filename string) time.Time {
	sessionID, err := getLogSessionID(filename)
	if err != nil {
		return time.Time{}
	}
	return sessionID.toTime()
}

// MatchBridgeLogName return true iff filename is a bridge log filename.
func MatchBridgeLogName(filename string) bool {
	return matchLogName(filename, BridgeShortAppName)
}

// MatchGUILogName return true iff filename is a bridge-gui log filename.
func MatchGUILogName(filename string) bool {
	return matchLogName(filename, GUIShortAppName)
}

// MatchLauncherLogName return true iff filename is a launcher log filename.
func MatchLauncherLogName(filename string) bool {
	return matchLogName(filename, LauncherShortAppName)
}

func matchLogName(logName string, appName AppName) bool {
	return regexp.MustCompile(`^\d{8}_\d{9}_\Q` + string(appName) + `\E_\d{3}_.*\.log$`).MatchString(logName)
}

type logKey string

const logrusFields = logKey("logrus")

func WithLogrusField(ctx context.Context, key string, value interface{}) context.Context {
	fields, ok := ctx.Value(logrusFields).(logrus.Fields)
	if !ok || fields == nil {
		fields = logrus.Fields{}
	}

	fields[key] = value
	return context.WithValue(ctx, logrusFields, fields)
}
