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

package crash

import (
	"fmt"

	"github.com/0xAX/notificator"
)

// ShowErrorNotification shows a system notification that the app with the given appName has crashed.
// NOTE: Icons shouldn't be hardcoded.
func ShowErrorNotification(appName string) RecoveryAction {
	return func(_ interface{}) error {
		notify := notificator.New(notificator.Options{
			DefaultIcon: "../frontend/ui/icon/icon.png",
			AppName:     appName,
		})

		return notify.Push(
			"Fatal Error",
			fmt.Sprintf("%v has encountered a fatal error.", appName),
			"/frontend/icon/icon.png",
			notificator.UR_CRITICAL,
		)
	}
}
