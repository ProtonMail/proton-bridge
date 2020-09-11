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

package transfer

import (
	"sync"
	"time"
)

type timeIt struct {
	lock    sync.Locker
	name    string
	groups  map[string]int64
	ongoing map[string]time.Time
}

func newTimeIt(name string) *timeIt {
	return &timeIt{
		lock:    &sync.Mutex{},
		name:    name,
		groups:  map[string]int64{},
		ongoing: map[string]time.Time{},
	}
}

func (t *timeIt) clear() {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.groups = map[string]int64{}
	t.ongoing = map[string]time.Time{}
}

func (t *timeIt) start(group, id string) {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.ongoing[group+"/"+id] = time.Now()
}

func (t *timeIt) stop(group, id string) {
	endTime := time.Now()

	t.lock.Lock()
	defer t.lock.Unlock()

	startTime, ok := t.ongoing[group+"/"+id]
	if !ok {
		log.WithField("group", group).WithField("id", id).Error("Stop called before start")
		return
	}
	delete(t.ongoing, group+"/"+id)

	diff := endTime.Sub(startTime).Milliseconds()
	t.groups[group] += diff
}

func (t *timeIt) logResults() {
	t.lock.Lock()
	defer t.lock.Unlock()

	// Print also ongoing to be sure that nothing was left out.
	// Basically ongoing should be empty.
	log.WithField("name", t.name).WithField("result", t.groups).WithField("ongoing", t.ongoing).Debug("Time measurement")
}
