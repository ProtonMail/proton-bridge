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

package logging

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/ProtonMail/proton-bridge/internal/constants"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/writer"
)

const (
	// MaxLogSize defines the maximum log size we should permit.
	// Zendesk has a file size limit of 20MB. When the last N log files are zipped,
	// it should fit under 20MB. So here we permit up to 10MB (most files are a few hundred kB).
	MaxLogSize = 10 * 2 << 20

	// MaxLogs defines how many old log files should be kept.
	MaxLogs = 3
)

func Init(logsPath string) error {
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: time.StampMilli,
	})

	rotator, err := NewRotator(MaxLogSize, func() (io.WriteCloser, error) {
		if err := clearLogs(logsPath, MaxLogs); err != nil {
			return nil, err
		}

		return os.Create(filepath.Join(logsPath, getLogName(constants.Version, constants.Revision)))
	})
	if err != nil {
		return err
	}

	logrus.SetOutput(rotator)

	logrus.AddHook(&writer.Hook{
		Writer: os.Stderr,
		LogLevels: []logrus.Level{
			logrus.PanicLevel,
			logrus.FatalLevel,
			logrus.ErrorLevel,
		},
	})

	return nil
}

func SetLevel(level string) {
	if lvl, err := logrus.ParseLevel(level); err == nil {
		logrus.SetLevel(lvl)
	}

	if logrus.GetLevel() == logrus.DebugLevel || logrus.GetLevel() == logrus.TraceLevel {
		logrus.SetOutput(os.Stderr)
	}
}

func getLogName(version, revision string) string {
	return fmt.Sprintf("v%v_%v_%v.log", version, revision, time.Now().Unix())
}

func matchLogName(name string) bool {
	return regexp.MustCompile(`^v.*\.log$`).MatchString(name)
}
