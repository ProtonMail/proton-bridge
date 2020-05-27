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

// +build !nogui

package qtcommon

// Positions of notification bubble
const (
	TabAccount    = 0
	TabSettings   = 1
	TabHelp       = 2
	TabQuit       = 4
	TabUpdates    = 100
	TabAddAccount = -1
)

// Notifier show bubble notification at postion marked by int
type Notifier interface {
	NotifyBubble(int, string)
}

// SendNotification unifies notification in GUI
func SendNotification(qml Notifier, tabIndex int, msg string) {
	qml.NotifyBubble(tabIndex, msg)
}
