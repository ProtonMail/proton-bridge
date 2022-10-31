// Copyright (c) 2022 Proton AG
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
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime/pprof"
	"time"

	"github.com/ProtonMail/proton-bridge/v2/internal/constants"
	"github.com/ProtonMail/proton-bridge/v2/internal/crash"
	"github.com/sirupsen/logrus"
)

func DumpStackTrace(logsPath string) crash.RecoveryAction {
	return func(r interface{}) error {
		file := filepath.Join(logsPath, getStackTraceName(constants.Version, constants.Revision))

		f, err := os.OpenFile(filepath.Clean(file), os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0o600)
		if err != nil {
			return err
		}

		if _, err := f.WriteString(fmt.Sprintf("Recover: %v", r)); err != nil {
			return err
		}

		if err := pprof.Lookup("goroutine").WriteTo(f, 2); err != nil {
			return err
		}

		logrus.WithField("file", file).Warn("Saved crash report")

		return nil
	}
}

func getStackTraceName(version, revision string) string {
	return fmt.Sprintf("v%v_%v_crash_%v.log", version, revision, time.Now().Unix())
}

func MatchStackTraceName(name string) bool {
	return regexp.MustCompile(`^v.*_crash_.*\.log$`).MatchString(name)
}
