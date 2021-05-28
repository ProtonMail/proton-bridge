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

package serverutil

import (
	"time"

	"github.com/ProtonMail/proton-bridge/internal/events"
	"github.com/ProtonMail/proton-bridge/pkg/listener"
)

// Server which can handle disconnected users and lost internet connection.
type Server interface {
	DisconnectUser(string)
	ListenRetryAndServe(int, time.Duration)
}

func monitorDisconnectedUsers(s Server, l listener.Listener) {
	ch := make(chan string)
	l.Add(events.CloseConnectionEvent, ch)
	for address := range ch {
		s.DisconnectUser(address)
	}
}

// ListenAndServe starts the server and keeps it on based on internet
// availability. It also monitors and disconnect users if requested.
func ListenAndServe(s Server, l listener.Listener) {
	go monitorDisconnectedUsers(s, l)

	// When starting the Bridge, we don't want to retry to notify user
	// quickly about the issue. Very probably retry will not help anyway.
	s.ListenRetryAndServe(0, 0)
}
