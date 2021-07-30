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

package store

import (
	"sync"

	"github.com/sirupsen/logrus"
)

type Cacher struct {
	storer  Storer
	jobs    chan string
	done    chan struct{}
	started bool
	wg      *sync.WaitGroup
}

type Storer interface {
	IsCached(messageID string) bool
	BuildAndCacheMessage(messageID string) error
}

func newCacher(storer Storer) *Cacher {
	return &Cacher{
		storer: storer,
		jobs:   make(chan string),
		done:   make(chan struct{}),
		wg:     &sync.WaitGroup{},
	}
}

// newJob sends a new job to the cacher if it's running.
func (cacher *Cacher) newJob(messageID string) {
	if !cacher.started {
		return
	}

	select {
	case <-cacher.done:
		return

	default:
		if !cacher.storer.IsCached(messageID) {
			cacher.wg.Add(1)
			go func() { cacher.jobs <- messageID }()
		}
	}
}

func (cacher *Cacher) start() {
	cacher.started = true

	go func() {
		for {
			select {
			case messageID := <-cacher.jobs:
				go cacher.handleJob(messageID)

			case <-cacher.done:
				return
			}
		}
	}()
}

func (cacher *Cacher) handleJob(messageID string) {
	defer cacher.wg.Done()

	if err := cacher.storer.BuildAndCacheMessage(messageID); err != nil {
		logrus.WithError(err).Error("Failed to build and cache message")
	} else {
		logrus.WithField("messageID", messageID).Trace("Message cached")
	}
}

func (cacher *Cacher) stop() {
	cacher.started = false

	cacher.wg.Wait()

	select {
	case <-cacher.done:
		return

	default:
		close(cacher.done)
	}
}
