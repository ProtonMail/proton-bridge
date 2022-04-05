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

import bolt "go.etcd.io/bbolt"

const (
	versionKey = "version"

	// versionOffset makes it possible to force email client to reload all
	// mailboxes. If increased during application update it will trigger
	// the reload on client side without needing to sync DB or re-setup account.
	versionOffset = uint32(3)
)

func (store *Store) getMailboxesVersion() uint32 {
	localVersion := store.readMailboxesVersion()
	// If a read error occurs it returns 0 which is an invalid version value.
	if localVersion == 0 {
		localVersion = 1
		_ = store.writeMailboxesVersion(localVersion)
	}

	// versionOffset will make email clients reload if increased during bridge update.
	return localVersion + versionOffset
}

func (store *Store) increaseMailboxesVersion() error {
	ver := store.readMailboxesVersion()
	// The version is zero if a read error occurred. Operation ++ will make it 1
	// which is default starting value.
	ver++
	return store.writeMailboxesVersion(ver)
}

func (store *Store) readMailboxesVersion() (version uint32) {
	_ = store.db.View(func(tx *bolt.Tx) (err error) {
		b := tx.Bucket(mboxVersionBucket)
		verRaw := b.Get([]byte(versionKey))
		if verRaw != nil {
			version = btoi(verRaw)
		}
		return nil
	})
	return
}

func (store *Store) writeMailboxesVersion(ver uint32) error {
	return store.db.Update(func(tx *bolt.Tx) (err error) {
		b := tx.Bucket(mboxVersionBucket)
		return b.Put([]byte(versionKey), itob(ver))
	})
}
