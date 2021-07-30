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

package cache

import (
	"errors"
	"sync"
)

type inMemoryCache struct {
	lock        sync.RWMutex
	data        map[string]map[string][]byte
	size, limit int
}

// NewInMemoryCache creates a new in memory cache which stores up to the given number of bytes of cached data.
// NOTE(GODT-1158): Make this threadsafe.
func NewInMemoryCache(limit int) Cache {
	return &inMemoryCache{
		data:  make(map[string]map[string][]byte),
		limit: limit,
	}
}

func (c *inMemoryCache) Unlock(userID string, passphrase []byte) error {
	c.data[userID] = make(map[string][]byte)
	return nil
}

func (c *inMemoryCache) Delete(userID string) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	for _, message := range c.data[userID] {
		c.size -= len(message)
	}

	delete(c.data, userID)

	return nil
}

// Has returns whether the given message exists in the cache.
func (c *inMemoryCache) Has(userID, messageID string) bool {
	if _, err := c.Get(userID, messageID); err != nil {
		return false
	}

	return true
}

func (c *inMemoryCache) Get(userID, messageID string) ([]byte, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	literal, ok := c.data[userID][messageID]
	if !ok {
		return nil, errors.New("no such message in cache")
	}

	return literal, nil
}

// NOTE(GODT-1158): What to actually do when memory limit is reached? Replace something existing? Return error? Drop silently?
// NOTE(GODT-1158): Pull in cache-rotating feature from old IMAP cache.
func (c *inMemoryCache) Set(userID, messageID string, literal []byte) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.size+len(literal) > c.limit {
		return nil
	}

	c.size += len(literal)
	c.data[userID][messageID] = literal

	return nil
}

func (c *inMemoryCache) Rem(userID, messageID string) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.size -= len(c.data[userID][messageID])

	delete(c.data[userID], messageID)

	return nil
}
