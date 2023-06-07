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

package logging

import (
	"context"
	"errors"
	"os"
	"regexp"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	// DefaultMaxLogFileSize defines the maximum log size we should permit: 5 MB
	//
	// The Zendesk limit for an attachment is 50MB and this is what will
	// be allowed via the API. However, if that fails for some reason, the
	// fallback is sending the report via email, which has a limit of 10mb
	// total or 7MB per file. Since we can produce up to 6 logs, and we
	// compress all the files (average compression - 80%), we need to have
	// a limit of 30MB total before compression, hence 5MB per log file.
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
func Init(logsPath string, sessionID SessionID, appName AppName, rotationSize, pruningSize int64, level string) error {
	logrus.SetFormatter(&logrus.TextFormatter{
		DisableColors:   true,
		FullTimestamp:   true,
		TimestampFormat: time.StampMilli,
	})

	logrus.AddHook(newColoredStdOutHook())

	rotator, err := NewDefaultRotator(logsPath, sessionID, appName, rotationSize, pruningSize)
	if err != nil {
		return err
	}

	logrus.SetOutput(rotator)

	return setLevel(level)
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
