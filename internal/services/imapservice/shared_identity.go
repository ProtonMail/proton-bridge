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
	"sync"

	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/v3/internal/services/useridentity"
	"golang.org/x/exp/slices"
)

type sharedIdentity interface {
	UserID() string
	GetAddress(id string) (proton.Address, bool)
	GetPrimaryAddress() (proton.Address, error)
	GetAddresses() []proton.Address
	WithAddrKR(addrID string, fn func(userKR, addrKR *crypto.KeyRing) error) error
	CheckAuth(email string, password []byte) (string, error)
}

type rwIdentity struct {
	lock               sync.RWMutex
	identity           *useridentity.State
	bridgePassProvider useridentity.BridgePassProvider
	keyPassProvider    useridentity.KeyPassProvider
}

func (r *rwIdentity) GetPrimaryAddress() (proton.Address, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	return r.identity.GetPrimaryAddr()
}

func (r *rwIdentity) GetAddresses() []proton.Address {
	r.lock.RLock()
	defer r.lock.RUnlock()

	return slices.Clone(r.identity.AddressesSorted)
}

func (r *rwIdentity) Clone() *useridentity.State {
	r.lock.RLock()
	defer r.lock.RUnlock()

	return r.identity.Clone()
}

func newRWIdentity(identity *useridentity.State,
	bridgePassProvider useridentity.BridgePassProvider,
	keyPassProvider useridentity.KeyPassProvider,
) *rwIdentity {
	return &rwIdentity{
		identity:           identity,
		bridgePassProvider: bridgePassProvider,
		keyPassProvider:    keyPassProvider,
	}
}

func (r *rwIdentity) UserID() string {
	r.lock.RLock()
	defer r.lock.RUnlock()

	return r.identity.User.ID
}

func (r *rwIdentity) GetAddress(id string) (proton.Address, bool) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	return r.identity.GetAddrByID(id)
}

func (r *rwIdentity) WithAddrKR(addrID string, fn func(userKR *crypto.KeyRing, addrKR *crypto.KeyRing) error) error {
	r.lock.RLock()
	defer r.lock.RUnlock()

	return r.identity.WithAddrKR(addrID, r.keyPassProvider.KeyPass(), fn)
}

func (r *rwIdentity) WithAddrKRs(fn func(*crypto.KeyRing, map[string]*crypto.KeyRing) error) error {
	r.lock.RLock()
	defer r.lock.RUnlock()

	return r.identity.WithAddrKRs(r.keyPassProvider.KeyPass(), fn)
}

func (r *rwIdentity) CheckAuth(email string, password []byte) (string, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	return r.identity.CheckAuth(email, password, r.bridgePassProvider)
}

func (r *rwIdentity) Write(f func(identity *useridentity.State) error) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	return f(r.identity)
}
