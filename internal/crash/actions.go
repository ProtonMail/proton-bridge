// Copyright (c) 2026 Proton AG
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

package crash

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/0xAX/notificator"
	"github.com/sirupsen/logrus"
)

// ShowErrorNotification shows a system notification that the app with the given appName has crashed.
// NOTE: Icons shouldn't be hardcoded.
func ShowErrorNotification(appName string) RecoveryAction {
	return func(_ interface{}) error {
		// Skip silently on platforms where the underlying notifier binary is
		// not available. Headless Linux installs (Debian 13 minimal, server
		// installs, CI containers) frequently lack notify-send and the
		// underlying notificator library does not guard against this, which
		// previously caused a nil dereference during crash recovery.
		if !notifierAvailable() {
			return nil
		}

		notify := notificator.New(notificator.Options{
			DefaultIcon: "../frontend/ui/icon/icon.png",
			AppName:     appName,
		})
		if notify == nil {
			return nil
		}

		return notify.Push(
			"Fatal Error",
			fmt.Sprintf("%v has encountered a fatal error.", appName),
			"/frontend/icon/icon.png",
			notificator.UR_CRITICAL,
		)
	}
}

// notifierAvailable reports whether the platform notifier binary used by
// github.com/0xAX/notificator is present on PATH. The library calls these
// binaries unconditionally and will surface a nil pointer or exec error if
// they are missing.
func notifierAvailable() bool {
	switch runtime.GOOS {
	case "linux":
		if _, err := exec.LookPath("notify-send"); err != nil {
			logrus.WithError(err).Debug("notify-send not found; skipping desktop notification")
			return false
		}
		return true
	case "darwin", "windows":
		return true
	default:
		return false
	}
}
