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
	"context"
	"time"

	"github.com/ProtonMail/proton-bridge/v2/internal/store/cache"
)

func (store *Store) StartWatcher() {
	if !cache.IsOnDiskCache(store.cache) {
		return
	}

	store.done = make(chan struct{})

	ctx, cancel := context.WithCancel(context.Background())
	store.msgCachePool.ctx = ctx

	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()
		defer cancel()

		for {
			// NOTE(GODT-1158): Race condition here? What if DB was already closed?
			messageIDs, err := store.getAllMessageIDs()
			if err != nil {
				return
			}

			for _, messageID := range messageIDs {
				if !store.IsCached(messageID) {
					store.msgCachePool.newJob(messageID)
				}
			}

			select {
			case <-store.done:
				return
			case <-ticker.C:
				continue
			}
		}
	}()
}

func (store *Store) stopWatcher() {
	if store.done == nil {
		return
	}

	select {
	default:
		close(store.done)

	case <-store.done:
		return
	}
}
