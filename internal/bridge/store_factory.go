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

package bridge

import (
	"fmt"
	"path/filepath"

	"github.com/ProtonMail/proton-bridge/v2/internal/sentry"
	"github.com/ProtonMail/proton-bridge/v2/internal/store"
	"github.com/ProtonMail/proton-bridge/v2/internal/store/cache"
	"github.com/ProtonMail/proton-bridge/v2/internal/users"
	"github.com/ProtonMail/proton-bridge/v2/pkg/listener"
	"github.com/ProtonMail/proton-bridge/v2/pkg/message"
)

type storeFactory struct {
	cacheProvider  CacheProvider
	sentryReporter *sentry.Reporter
	panicHandler   users.PanicHandler
	eventListener  listener.Listener
	events         *store.Events
	cache          cache.Cache
	builder        *message.Builder
}

func newStoreFactory(
	cacheProvider CacheProvider,
	sentryReporter *sentry.Reporter,
	panicHandler users.PanicHandler,
	eventListener listener.Listener,
	cache cache.Cache,
	builder *message.Builder,
) *storeFactory {
	return &storeFactory{
		cacheProvider:  cacheProvider,
		sentryReporter: sentryReporter,
		panicHandler:   panicHandler,
		eventListener:  eventListener,
		events:         store.NewEvents(cacheProvider.GetIMAPCachePath()),
		cache:          cache,
		builder:        builder,
	}
}

// New creates new store for given user.
func (f *storeFactory) New(user store.BridgeUser) (*store.Store, error) {
	return store.New(
		f.sentryReporter,
		f.panicHandler,
		user,
		f.eventListener,
		f.cache,
		f.builder,
		getUserStorePath(f.cacheProvider.GetDBDir(), user.ID()),
		f.events,
	)
}

// Remove removes all store files for given user.
func (f *storeFactory) Remove(userID string) error {
	return store.RemoveStore(
		f.events,
		getUserStorePath(f.cacheProvider.GetDBDir(), userID),
		userID,
	)
}

// getUserStorePath returns the file path of the store database for the given userID.
func getUserStorePath(storeDir string, userID string) (path string) {
	return filepath.Join(storeDir, fmt.Sprintf("mailbox-%v.db", userID))
}
