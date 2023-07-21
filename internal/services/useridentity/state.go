// Copyright (c) 2023 Proton AG
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

package useridentity

import (
	"context"
	"fmt"
	"strings"

	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/proton-bridge/v3/internal/usertypes"
	"golang.org/x/exp/maps"
)

// State holds all the required user identity state. The idea of this type is that
// it can be replicated across all services to avoid lock contention. The only
// requirement is that the service with the respective events.
type State struct {
	AddressesSorted []proton.Address
	Addresses       map[string]proton.Address
	User            proton.User

	provider IdentityProvider
}

func NewState(
	user proton.User,
	addresses []proton.Address,
	provider IdentityProvider,
) State {
	addressMap := buildAddressMapFromSlice(addresses)
	return State{
		AddressesSorted: sortAddresses(maps.Values(addressMap)),
		Addresses:       addressMap,
		User:            user,

		provider: provider,
	}
}

func NewStateFromProvider(ctx context.Context, provider IdentityProvider) (State, error) {
	user, err := provider.GetUser(ctx)
	if err != nil {
		return State{}, fmt.Errorf("failed to get user: %w", err)
	}

	addresses, err := provider.GetAddresses(ctx)
	if err != nil {
		return State{}, fmt.Errorf("failed to get user addresses: %w", err)
	}

	return NewState(user, addresses, provider), nil
}

// GetAddr returns the address for the given email address.
func (s *State) GetAddr(email string) (proton.Address, error) {
	for _, addr := range s.AddressesSorted {
		if strings.EqualFold(addr.Email, usertypes.SanitizeEmail(email)) {
			return addr, nil
		}
	}

	return proton.Address{}, fmt.Errorf("address %s not found", email)
}

// GetPrimaryAddr returns the primary address for this user.
func (s *State) GetPrimaryAddr() (proton.Address, error) {
	if len(s.AddressesSorted) == 0 {
		return proton.Address{}, fmt.Errorf("no addresses available")
	}

	return s.AddressesSorted[0], nil
}

func (s *State) OnUserEvent(user proton.User) {
	s.User = user
}

func (s *State) OnRefreshEvent(ctx context.Context) error {
	user, err := s.provider.GetUser(ctx)
	if err != nil {
		return fmt.Errorf("failed to get user:%w", err)
	}

	addresses, err := s.provider.GetAddresses(ctx)
	if err != nil {
		return fmt.Errorf("failed to get addresses:%w", err)
	}

	s.User = user
	s.Addresses = buildAddressMapFromSlice(addresses)
	s.AddressesSorted = sortAddresses(maps.Values(s.Addresses))

	return nil
}

func (s *State) OnUserSpaceChanged(value int) bool {
	if s.User.UsedSpace == value {
		return false
	}

	s.User.UsedSpace = value

	return true
}

type AddressUpdate int

const (
	AddressUpdateNoop AddressUpdate = iota
	AddressUpdateCreated
	AddressUpdateEnabled
	AddressUpdateDisabled
	AddressUpdateUpdated
	AddressUpdateDeleted
)

func (s *State) OnAddressCreated(event proton.AddressEvent) AddressUpdate {
	if _, ok := s.Addresses[event.Address.ID]; ok {
		return AddressUpdateNoop
	}

	s.Addresses[event.Address.ID] = event.Address
	s.AddressesSorted = sortAddresses(maps.Values(s.Addresses))

	if event.Address.Status != proton.AddressStatusEnabled {
		return AddressUpdateNoop
	}

	return AddressUpdateCreated
}

func (s *State) OnAddressUpdated(event proton.AddressEvent) (proton.Address, AddressUpdate) {
	// Address does not exist create it.
	oldAddr, ok := s.Addresses[event.Address.ID]
	if !ok {
		return event.Address, s.OnAddressCreated(event)
	}

	s.Addresses[event.Address.ID] = event.Address
	s.AddressesSorted = sortAddresses(maps.Values(s.Addresses))

	switch {
	// If the address was newly enabled:
	case oldAddr.Status != proton.AddressStatusEnabled && event.Address.Status == proton.AddressStatusEnabled:
		return event.Address, AddressUpdateEnabled

	// If the address was newly disabled:
	case oldAddr.Status == proton.AddressStatusEnabled && event.Address.Status != proton.AddressStatusEnabled:
		return event.Address, AddressUpdateDisabled

	// Otherwise it's just an update:
	default:
		return event.Address, AddressUpdateUpdated
	}
}

func (s *State) OnAddressDeleted(event proton.AddressEvent) (proton.Address, AddressUpdate) {
	addr, ok := s.Addresses[event.ID]
	if !ok {
		return proton.Address{}, AddressUpdateNoop
	}

	delete(s.Addresses, event.ID)
	s.AddressesSorted = sortAddresses(maps.Values(s.Addresses))

	if addr.Status != proton.AddressStatusEnabled {
		return proton.Address{}, AddressUpdateNoop
	}

	return addr, AddressUpdateDeleted
}
