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
	"encoding/json"
	"os"
	"sync"

	"github.com/pkg/errors"
)

// Cache caches the last event IDs for all accounts (there should be only one instance).
type Cache struct {
	// cache is map from userID => key (such as last event) => value (such as event ID).
	cache map[string]map[string]string
	path  string
	lock  *sync.RWMutex
}

// NewCache constructs a new cache at the given path.
func NewCache(path string) *Cache {
	return &Cache{
		path: path,
		lock: &sync.RWMutex{},
	}
}

func (c *Cache) getEventID(userID string) string {
	c.lock.Lock()
	defer c.lock.Unlock()

	if err := c.loadCache(); err != nil {
		log.WithError(err).Warn("Problem to load store cache")
	}

	if c.cache == nil {
		c.cache = map[string]map[string]string{}
	}
	if c.cache[userID] == nil {
		c.cache[userID] = map[string]string{}
	}

	return c.cache[userID]["events"]
}

func (c *Cache) setEventID(userID, eventID string) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.cache[userID] == nil {
		c.cache[userID] = map[string]string{}
	}
	c.cache[userID]["events"] = eventID

	return c.saveCache()
}

func (c *Cache) loadCache() error {
	if c.cache != nil {
		return nil
	}

	f, err := os.Open(c.path)
	if err != nil {
		return err
	}
	defer f.Close() //nolint[errcheck]

	return json.NewDecoder(f).Decode(&c.cache)
}

func (c *Cache) saveCache() error {
	if c.cache == nil {
		return errors.New("events: cannot save cache: cache is nil")
	}

	f, err := os.Create(c.path)
	if err != nil {
		return err
	}
	defer f.Close() //nolint[errcheck]

	return json.NewEncoder(f).Encode(c.cache)
}

func (c *Cache) clearCacheUser(userID string) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.cache == nil {
		log.WithField("user", userID).Warning("Cannot clear user from cache: cache is nil")
		return nil
	}

	log.WithField("user", userID).Trace("Removing user from event loop cache")

	delete(c.cache, userID)

	return c.saveCache()
}
