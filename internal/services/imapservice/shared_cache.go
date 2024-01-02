// Copyright (c) 2024 Proton AG
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
// along with Proton Mail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package imapservice

import (
	"context"
	"errors"
	"sync"

	"github.com/ProtonMail/gluon/connector"
)

type CacheAccessor interface {
	connector.IMAPState
	Close()
}

// SharedCache is meant to protect access to the database and guarantee it's always valid. There may be some corner
// cases where the Gluon connector can get closed while we are processing events in parallel. If for some reason
// Gluon closes the database, the instance is invalidated and any attempts to access this state will return
// `ErrCacheNotAvailable`.
type SharedCache struct {
	cache connector.IMAPState
	lock  sync.RWMutex
}

func NewSharedCached() *SharedCache {
	return &SharedCache{}
}

var ErrCacheNotAvailable = errors.New("cache no longer available")

func (s *SharedCache) Set(cache connector.IMAPState) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.cache = cache
}

func (s *SharedCache) Close() {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.cache = nil
}

func (s *SharedCache) Acquire() (CacheAccessor, error) {
	s.lock.RLock()

	if s.cache == nil {
		s.lock.RUnlock()
		return nil, ErrCacheNotAvailable
	}

	return &cacheAccessor{sharedCache: s}, nil
}

type cacheAccessor struct {
	sharedCache *SharedCache
}

func (c cacheAccessor) Read(ctx context.Context, f func(context.Context, connector.IMAPStateRead) error) error {
	return c.sharedCache.cache.Read(ctx, f)
}

func (c cacheAccessor) Write(ctx context.Context, f func(context.Context, connector.IMAPStateWrite) error) error {
	return c.sharedCache.cache.Write(ctx, f)
}

func (c cacheAccessor) Close() {
	c.sharedCache.lock.RUnlock()
}
