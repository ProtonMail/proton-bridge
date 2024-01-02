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

package userevents

import (
	"context"
	"sync"

	"github.com/ProtonMail/proton-bridge/v3/internal/vault"
)

// EventIDStore exposes behavior expected of a type which allows us to store and retrieve event Ids.
// Note: this may be accessed from multiple go-routines.
type EventIDStore interface {
	// Load the last stored event, return "" for empty.
	Load(ctx context.Context) (string, error)
	// Store the new id.
	Store(ctx context.Context, id string) error
}

type InMemoryEventIDStore struct {
	lock sync.Mutex
	id   string
}

func NewInMemoryEventIDStore() *InMemoryEventIDStore {
	return &InMemoryEventIDStore{}
}

func (i *InMemoryEventIDStore) Load(_ context.Context) (string, error) {
	i.lock.Lock()
	defer i.lock.Unlock()

	return i.id, nil
}

func (i *InMemoryEventIDStore) Store(_ context.Context, id string) error {
	i.lock.Lock()
	defer i.lock.Unlock()

	i.id = id

	return nil
}

type VaultEventIDStore struct {
	vault *vault.User
}

func NewVaultEventIDStore(vault *vault.User) *VaultEventIDStore {
	return &VaultEventIDStore{vault: vault}
}

func (v VaultEventIDStore) Load(_ context.Context) (string, error) {
	return v.vault.EventID(), nil
}

func (v VaultEventIDStore) Store(_ context.Context, id string) error {
	return v.vault.SetEventID(id)
}
