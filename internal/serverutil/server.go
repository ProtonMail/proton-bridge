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
	"github.com/ProtonMail/proton-bridge/pkg/ports"
	"github.com/sirupsen/logrus"
)

// Server which can handle disconnected users and lost internet connection.
type Server interface {
	HandlePanic()
	DisconnectUser(string)
	ListenRetryAndServe(int, time.Duration)
	Close()
	Port() int
	IsRunning() bool
}

func monitorDisconnectedUsers(s Server, l listener.Listener) {
	ch := make(chan string)
	l.Add(events.CloseConnectionEvent, ch)
	for address := range ch {
		s.DisconnectUser(address)
	}
}

func redirectInternetEventsToOneChannel(l listener.Listener) (isInternetOn chan bool) {
	on := make(chan string)
	l.Add(events.InternetOnEvent, on)
	off := make(chan string)
	l.Add(events.InternetOffEvent, off)

	// Redirect two channels into one. When select was used the algorithm
	// first read all on channels and then read all off channels.
	isInternetOn = make(chan bool, 20)
	go func() {
		for {
			logrus.WithField("try", <-on).Trace("Internet ON")
			isInternetOn <- true
		}
	}()

	go func() {
		for {
			logrus.WithField("try", <-off).Trace("Internet OFF")
			isInternetOn <- false
		}
	}()
	return
}

const (
	recheckPortAfter    = 50 * time.Millisecond
	stopPortChecksAfter = 15 * time.Second
	retryListenerAfter  = 5 * time.Second
)

func monitorInternetConnection(s Server, l listener.Listener) {
	isInternetOn := redirectInternetEventsToOneChannel(l)
	for {
		var expectedIsPortFree bool
		if <-isInternetOn {
			if s.IsRunning() {
				continue
			}
			go func() {
				defer s.HandlePanic()
				// We had issues on Mac that from time to time something
				// blocked our port for a bit after we closed IMAP server
				// due to connection issues.
				// Restart always helped, so we do retry to not bother user.
				s.ListenRetryAndServe(10, retryListenerAfter)
			}()
			expectedIsPortFree = false
		} else {
			if !s.IsRunning() {
				continue
			}
			s.Close()
			expectedIsPortFree = true
		}
		start := time.Now()
		for {
			isPortFree := ports.IsPortFree(s.Port())
			logrus.
				WithField("port", s.Port()).
				WithField("isFree", isPortFree).
				WithField("wantToBeFree", expectedIsPortFree).
				Trace("Check port")
			if isPortFree == expectedIsPortFree {
				break
			}
			// Safety stop if something went wrong.
			if time.Since(start) > stopPortChecksAfter {
				logrus.WithField("expectedIsPortFree", expectedIsPortFree).Warn("Server start/stop check timeouted")
				break
			}
			time.Sleep(recheckPortAfter)
		}
	}
}

// ListenAndServe starts the server and keeps it on based on internet
// availability. It also monitors and disconnect users if requested.
func ListenAndServe(s Server, l listener.Listener) {
	go monitorDisconnectedUsers(s, l)
	go monitorInternetConnection(s, l)

	// When starting the Bridge, we don't want to retry to notify user
	// quickly about the issue. Very probably retry will not help anyway.
	s.ListenRetryAndServe(0, 0)
}
