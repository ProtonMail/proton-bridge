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

// NewInMemoryCache creates a new in memory cache which stores up to the given
// number of bytes of cached data.
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

func (c *inMemoryCache) Lock(userID string) {
	c.lock.Lock()
	defer c.lock.Unlock()

	for _, message := range c.data[userID] {
		c.size -= len(message)
	}

	delete(c.data, userID)
}

func (c *inMemoryCache) Delete(userID string) error {
	c.Lock(userID)
	return nil
}

// Has returns whether the given message exists in the cache.
func (c *inMemoryCache) Has(userID, messageID string) bool {
	c.lock.RLock()
	defer c.lock.RUnlock()

	if !c.isUserUnlocked(userID) {
		// This might look counter intuitive but in order to be able to test
		// "re-unlocking" mechanism we need to return true here.
		//
		// The situation is the same as it would happen for onDiskCache with
		// locked user. Later during `Get` cache would return proper error
		// `ErrCacheNeedsUnlock`. It is expected that store would then try to
		// re-unlock.
		//
		// In order to do proper behaviour we should implement
		// encryption for inMemoryCache.
		return true
	}

	_, ok := c.data[userID][messageID]
	return ok
}

func (c *inMemoryCache) Get(userID, messageID string) ([]byte, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	if !c.isUserUnlocked(userID) {
		return nil, ErrCacheNeedsUnlock
	}

	literal, ok := c.data[userID][messageID]
	if !ok {
		return nil, errors.New("no such message in cache")
	}

	return literal, nil
}

func (c *inMemoryCache) isUserUnlocked(userID string) bool {
	_, ok := c.data[userID]
	return ok
}

// Set saves the message literal to memory for further usage.
//
// NOTE(GODT-1158, GODT-1488): Once memory limit is reached we should do proper
// rotation based on usage frequency.
func (c *inMemoryCache) Set(userID, messageID string, literal []byte) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	if !c.isUserUnlocked(userID) {
		return ErrCacheNeedsUnlock
	}

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

	if !c.isUserUnlocked(userID) {
		return nil
	}

	c.size -= len(c.data[userID][messageID])

	delete(c.data[userID], messageID)

	return nil
}
