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

package bridge

import (
	"fmt"
	"path/filepath"

	"github.com/ProtonMail/proton-bridge/internal/store"
	"github.com/ProtonMail/proton-bridge/internal/users"

	"github.com/ProtonMail/proton-bridge/pkg/listener"
)

type storeFactory struct {
	config        StoreFactoryConfiger
	panicHandler  users.PanicHandler
	clientManager users.ClientManager
	eventListener listener.Listener
	storeCache    *store.Cache
}

func newStoreFactory(
	config StoreFactoryConfiger,
	panicHandler users.PanicHandler,
	clientManager users.ClientManager,
	eventListener listener.Listener,
) *storeFactory {
	return &storeFactory{
		config:        config,
		panicHandler:  panicHandler,
		clientManager: clientManager,
		eventListener: eventListener,
		storeCache:    store.NewCache(config.GetIMAPCachePath()),
	}
}

// New creates new store for given user.
func (f *storeFactory) New(user store.BridgeUser) (*store.Store, error) {
	storePath := getUserStorePath(f.config.GetDBDir(), user.ID())
	return store.New(f.panicHandler, user, f.clientManager, f.eventListener, storePath, f.storeCache)
}

// Remove removes all store files for given user.
func (f *storeFactory) Remove(userID string) error {
	storePath := getUserStorePath(f.config.GetDBDir(), userID)
	return store.RemoveStore(f.storeCache, storePath, userID)
}

// getUserStorePath returns the file path of the store database for the given userID.
func getUserStorePath(storeDir string, userID string) (path string) {
	fileName := fmt.Sprintf("mailbox-%v.db", userID)
	return filepath.Join(storeDir, fileName)
}
