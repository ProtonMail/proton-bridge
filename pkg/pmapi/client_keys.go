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

package pmapi

import (
	"context"
)

// Unlock unlocks all the user and address keys using the given passphrase, creating user and address keyrings.
// If the keyrings are already present, they are not recreated.
func (c *client) Unlock(ctx context.Context, passphrase []byte) (err error) {
	c.keyRingLock.Lock()
	defer c.keyRingLock.Unlock()

	return c.unlock(ctx, passphrase)
}

// unlock unlocks the user's keys but without locking the keyring lock first.
// Should only be used internally by methods that first lock the lock.
func (c *client) unlock(ctx context.Context, passphrase []byte) error {
	if _, err := c.CurrentUser(ctx); err != nil {
		return err
	}

	if c.userKeyRing == nil {
		if err := c.unlockUser(passphrase); err != nil {
			return ErrUnlockFailed{err}
		}
	}

	for _, address := range c.addresses {
		if c.addrKeyRing[address.ID] == nil {
			if err := c.unlockAddress(passphrase, address); err != nil {
				return ErrUnlockFailed{err}
			}
		}
	}

	return nil
}

func (c *client) ReloadKeys(ctx context.Context, passphrase []byte) (err error) {
	c.keyRingLock.Lock()
	defer c.keyRingLock.Unlock()

	c.clearKeys()

	return c.unlock(ctx, passphrase)
}

func (c *client) clearKeys() {
	if c.userKeyRing != nil {
		c.userKeyRing.ClearPrivateParams()
		c.userKeyRing = nil
	}

	for id, kr := range c.addrKeyRing {
		if kr != nil {
			kr.ClearPrivateParams()
		}
		delete(c.addrKeyRing, id)
	}
}

func (c *client) IsUnlocked() bool {
	if c.userKeyRing == nil {
		return false
	}

	for _, address := range c.addresses {
		if address.HasKeys != MissingKeys && c.addrKeyRing[address.ID] == nil {
			return false
		}
	}

	return true
}
