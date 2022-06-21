// Copyright (c) 2022 Proton AG
//
// This file is part of Proton Mail Bridge.Bridge.
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

package liveapi

import (
	"context"
	"time"
)

func (ctl *Controller) LockEvents(username string) {
	ctl.recordEvent(username)
	persistentClients.eventsPaused.Add(1)
}

func (ctl *Controller) UnlockEvents(username string) {
	persistentClients.eventsPaused.Done()
	ctl.waitForEventChange(username)
}

func (ctl *Controller) recordEvent(username string) {
	ctl.lastEventByUsername[username] = ""
	client, err := getPersistentClient(username)
	if err != nil {
		ctl.log.WithError(err).Error("Cannot get persistent client to record event")
		return
	}

	event, err := client.GetEvent(context.Background(), "")
	if err != nil {
		ctl.log.WithError(err).Error("Cannot record event")
		return
	}

	ctl.lastEventByUsername[username] = event.EventID
	ctl.log.WithField("last", event.EventID).Debug("Event recorded")
}

func (ctl *Controller) waitForEventChange(username string) {
	lastEvent := ctl.lastEventByUsername[username]
	ctl.lastEventByUsername[username] = ""

	log := ctl.log.WithField("last", lastEvent)

	if lastEvent == "" {
		log.Debug("Nothing to wait for, event not recoreded")
		return
	}

	client, err := getPersistentClient(username)
	if err != nil {
		log.WithError(err).Error("Cannot get persistent client to check event")
		return
	}

	for i, delay := range []int{1, 5, 10, 30} {
		ilog := log.WithField("try", i)
		event, err := client.GetEvent(context.Background(), "")
		if err != nil {
			ilog.WithError(err).Error("Cannot check event")
			return
		}
		if lastEvent != event.EventID {
			ilog.WithField("current", event.EventID).Debug("Event changed")
			return
		}

		ilog.WithField("delay", delay).Warn("Retrying event change again")
		time.Sleep(time.Duration(delay) * time.Second)
	}

	ctl.log.WithError(err).Warn("Event check timed out")
}
