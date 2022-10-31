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

package store

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/pkg/errors"
)

// Events caches the last event IDs for all accounts (there should be only one instance).
type Events struct {
	// eventMap is map from userID => key (such as last event) => value (such as event ID).
	eventMap map[string]map[string]string
	path     string
	lock     *sync.RWMutex
}

// NewEvents constructs a new event cache at the given path.
func NewEvents(path string) *Events {
	return &Events{
		path: path,
		lock: &sync.RWMutex{},
	}
}

func (c *Events) getEventID(userID string) string {
	c.lock.Lock()
	defer c.lock.Unlock()

	if err := c.loadEvents(); err != nil {
		log.WithError(err).Warn("Problem to load store events")
	}

	if c.eventMap == nil {
		c.eventMap = map[string]map[string]string{}
	}
	if c.eventMap[userID] == nil {
		c.eventMap[userID] = map[string]string{}
	}

	return c.eventMap[userID]["events"]
}

func (c *Events) setEventID(userID, eventID string) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.eventMap[userID] == nil {
		c.eventMap[userID] = map[string]string{}
	}
	c.eventMap[userID]["events"] = eventID

	return c.saveEvents()
}

func (c *Events) loadEvents() error {
	if c.eventMap != nil {
		return nil
	}

	f, err := os.Open(c.path)
	if err != nil {
		return err
	}
	defer f.Close() //nolint:errcheck,gosec

	return json.NewDecoder(f).Decode(&c.eventMap)
}

func (c *Events) saveEvents() error {
	if c.eventMap == nil {
		return errors.New("events: cannot save events: events map is nil")
	}

	f, err := os.Create(c.path)
	if err != nil {
		return err
	}
	defer f.Close() //nolint:errcheck,gosec

	return json.NewEncoder(f).Encode(c.eventMap)
}

func (c *Events) clearUserEvents(userID string) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.eventMap == nil {
		log.WithField("user", userID).Warning("Cannot clear user events: event map is nil")
		return nil
	}

	log.WithField("user", userID).Trace("Removing user events from event loop")

	delete(c.eventMap, userID)

	return c.saveEvents()
}
