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
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"github.com/ProtonMail/proton-bridge/v3/internal/constants"
	"github.com/sirupsen/logrus"
)

const (
	// MaxLogSize defines the maximum log size we should permit: 5 MB
	//
	// The Zendesk limit for an attachement is 50MB and this is what will
	// be allowed via the API. However, if that fails for some reason, the
	// fallback is sending the report via email, which has a limit of 10mb
	// total or 7MB per file. Since we can produce up to 6 logs, and we
	// compress all the files (avarage compression - 80%), we need to have
	// a limit of 30MB total before compression, hence 5MB per log file.
	MaxLogSize = 5 * 1024 * 1024

	// MaxLogs defines how many log files should be kept.
	MaxLogs = 10
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

func Init(logsPath, level string) error {
	logrus.SetFormatter(&logrus.TextFormatter{
		DisableColors:   true,
		FullTimestamp:   true,
		TimestampFormat: time.StampMilli,
	})

	logrus.AddHook(newColoredStdOutHook())

	rotator, err := NewRotator(MaxLogSize, func() (io.WriteCloser, error) {
		if err := clearLogs(logsPath, MaxLogs, MaxLogs); err != nil {
			return nil, err
		}

		return os.Create(filepath.Join(logsPath, getLogName(constants.Version, constants.Revision))) //nolint:gosec // G304
	})
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

func getLogName(version, revision string) string {
	return fmt.Sprintf("v%v_%v_%v.log", version, revision, time.Now().Unix())
}

func getLogTime(name string) int {
	re := regexp.MustCompile(`^v.*_.*_(?P<timestamp>\d+).log$`)

	match := re.FindStringSubmatch(name)

	if len(match) == 0 {
		logrus.Warn("Could not parse log name: ", name)
		return 0
	}

	timestamp, err := strconv.Atoi(match[re.SubexpIndex("timestamp")])
	if err != nil {
		return 0
	}

	return timestamp
}

func MatchLogName(name string) bool {
	return regexp.MustCompile(`^v.*\.log$`).MatchString(name)
}

func MatchGUILogName(name string) bool {
	return regexp.MustCompile(`^gui_v.*\.log$`).MatchString(name)
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

func LogFromContext(ctx context.Context) *logrus.Entry {
	fields, ok := ctx.Value(logrusFields).(logrus.Fields)
	if !ok || fields == nil {
		return logrus.WithField("ctx", "empty")
	}

	return logrus.WithFields(fields)
}
