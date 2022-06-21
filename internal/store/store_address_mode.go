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
	"github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"
)

type addressMode string

const (
	splitMode    addressMode = "split"
	combinedMode addressMode = "combined"
	modeKey                  = "mode"
)

// getAddressMode returns the current address mode (split or combined) of the store.
// It first looks in the local cache but if that is not yet set, it loads it from the database.
func (store *Store) getAddressMode() (mode addressMode, err error) {
	if store.addressMode != "" {
		mode = store.addressMode
		return
	}

	tx := func(tx *bolt.Tx) (err error) {
		b := tx.Bucket(addressModeBucket)

		dbMode := b.Get([]byte(modeKey))
		if dbMode == nil {
			return errors.New("address mode not set")
		}

		mode = addressMode(dbMode)

		return
	}

	err = store.db.View(tx)

	return
}

// IsCombinedMode returns whether the store is set to combined mode.
func (store *Store) IsCombinedMode() bool {
	return store.addressMode == combinedMode
}

// UseCombinedMode sets whether the store should be set to combined mode.
func (store *Store) UseCombinedMode(useCombined bool) (err error) {
	if useCombined {
		err = store.switchAddressMode(combinedMode)
	} else {
		err = store.switchAddressMode(splitMode)
	}

	return
}

// switchAddressMode sets the address mode to the given value and rebuilds the mailboxes.
func (store *Store) switchAddressMode(mode addressMode) (err error) {
	if store.addressMode == mode {
		log.Debug("The store is using the correct address mode")
		return
	}

	if err = store.setAddressMode(mode); err != nil {
		log.WithError(err).Error("Could not set store address mode")
		return
	}

	if err = store.RebuildMailboxes(); err != nil {
		log.WithError(err).Error("Could not rebuild mailboxes after switching address mode")
		return
	}

	return
}

// setAddressMode sets the current address mode (split or combined) of the store.
// It writes to database and updates the local value in the store object.
func (store *Store) setAddressMode(mode addressMode) (err error) {
	store.log.WithField("mode", string(mode)).Info("Setting store address mode")

	tx := func(tx *bolt.Tx) (err error) {
		b := tx.Bucket(addressModeBucket)
		return b.Put([]byte(modeKey), []byte(mode))
	}

	if err = store.db.Update(tx); err != nil {
		return
	}

	store.addressMode = mode

	return
}
