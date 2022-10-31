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
	"encoding/json"
	"fmt"

	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
	"github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"
)

// GetAddress returns the store address by given ID.
func (store *Store) GetAddress(addressID string) (*Address, error) {
	store.lock.RLock()
	defer store.lock.RUnlock()

	storeAddress, ok := store.addresses[addressID]
	if !ok {
		return nil, fmt.Errorf("addressID %v does not exist", addressID)
	}

	return storeAddress, nil
}

// RebuildMailboxes truncates all mailbox buckets and recreates them from the metadata bucket again.
func (store *Store) RebuildMailboxes() (err error) {
	store.lock.Lock()
	defer store.lock.Unlock()

	log.WithField("user", store.UserID()).Trace("Truncating mailboxes")

	store.addresses = nil

	if err = store.truncateMailboxesBucket(); err != nil {
		log.WithError(err).Error("Could not truncate mailboxes bucket")
		return
	}

	if err = store.truncateAddressInfoBucket(); err != nil {
		log.WithError(err).Error("Could not truncate address info bucket")
		return
	}

	if err = store.init(false); err != nil {
		log.WithError(err).Error("Could not init store")
		return
	}

	if err := store.increaseMailboxesVersion(); err != nil {
		log.WithError(err).Error("Could not increase structure version")
		// Do not return here. The truncation was already done and mode
		// was changed in DB so we need to sync so that users start to see
		// messages and not block other operations.
	}

	log.WithField("user", store.UserID()).Trace("Rebuilding mailboxes")
	return store.initMailboxesBucket()
}

// createOrDeleteAddressesEvent creates address objects in the store for each necessary address
// and deletes any address objects that shouldn't be there.
// It doesn't do anything to addresses that are rightfully there.
// It should only be called from the event loop.
func (store *Store) createOrDeleteAddressesEvent() (err error) {
	store.lock.Lock()
	defer store.lock.Unlock()

	labels, err := store.initCounts()
	if err != nil {
		return errors.Wrap(err, "failed to initialise label counts")
	}

	addrInfo, err := store.GetAddressInfo()
	if err != nil {
		return errors.Wrap(err, "failed to get addresses and address IDs")
	}

	// We need at least one address to continue.
	if len(addrInfo) < 1 {
		return errors.New("no addresses to initialise")
	}

	// If in combined mode, we only need the user's primary address.
	if store.addressMode == combinedMode {
		addrInfo = addrInfo[:1]
	}

	// Go through all addresses that *should* be there.
	for _, addr := range addrInfo {
		if _, ok := store.addresses[addr.AddressID]; ok {
			continue
		}

		// This address is missing so we create it.
		if err = store.addAddress(addr.Address, addr.AddressID, labels); err != nil {
			return errors.Wrap(err, "failed to add address to store")
		}
	}

	// Go through all addresses that *should not* be there.
	for _, addr := range store.addresses {
		belongs := false

		for _, a := range addrInfo {
			if addr.addressID == a.AddressID {
				belongs = true
				break
			}
		}

		if belongs {
			continue
		}

		delete(store.addresses, addr.addressID)
	}

	if err = store.truncateMailboxesBucket(); err != nil {
		log.WithError(err).Error("Could not truncate mailboxes bucket")
		return
	}

	return store.initMailboxesBucket()
}

// truncateAddressInfoBucket removes the address info bucket.
func (store *Store) truncateAddressInfoBucket() (err error) {
	log.Trace("Truncating address info bucket")

	tx := func(tx *bolt.Tx) (err error) {
		if err = tx.DeleteBucket(addressInfoBucket); err != nil {
			return
		}

		if _, err = tx.CreateBucketIfNotExists(addressInfoBucket); err != nil {
			return
		}

		return
	}

	return store.db.Update(tx)
}

// truncateMailboxesBucket removes the mailboxes bucket.
func (store *Store) truncateMailboxesBucket() (err error) {
	log.Trace("Truncating mailboxes bucket")

	tx := func(tx *bolt.Tx) (err error) {
		mbs := tx.Bucket(mailboxesBucket)

		return mbs.ForEach(func(addrIDMailbox, _ []byte) (err error) {
			addr := mbs.Bucket(addrIDMailbox)

			if err = addr.DeleteBucket(imapIDsBucket); err != nil {
				return
			}

			if _, err = addr.CreateBucketIfNotExists(imapIDsBucket); err != nil {
				return
			}

			if err = addr.DeleteBucket(apiIDsBucket); err != nil {
				return
			}

			if _, err = addr.CreateBucketIfNotExists(apiIDsBucket); err != nil {
				return
			}

			return
		})
	}

	return store.db.Update(tx)
}

// initMailboxesBucket recreates the mailboxes bucket from the metadata bucket.
func (store *Store) initMailboxesBucket() error {
	return store.db.Update(func(tx *bolt.Tx) error {
		i := 0
		msgs := []*pmapi.Message{}

		err := tx.Bucket(metadataBucket).ForEach(func(k, v []byte) error {
			msg := &pmapi.Message{}

			if err := json.Unmarshal(v, msg); err != nil {
				return err
			}
			msgs = append(msgs, msg)

			// Calling txCreateOrUpdateMessages does some overhead by iterating
			// all mailboxes, accessing buckets and so on. It's better to do in
			// batches instead of one by one (seconds vs hours for huge accounts).
			// Average size of metadata is 1k bytes, sometimes up to 2k bytes.
			// 10k messages will take about 20 MB of memory.
			i++
			if i%10000 == 0 {
				store.log.WithField("i", i).Debug("Init mboxes heartbeat")

				for _, a := range store.addresses {
					if err := a.txCreateOrUpdateMessages(tx, msgs); err != nil {
						return err
					}
				}
				msgs = []*pmapi.Message{}
			}

			return nil
		})
		if err != nil {
			return err
		}

		for _, a := range store.addresses {
			if err := a.txCreateOrUpdateMessages(tx, msgs); err != nil {
				return err
			}
		}

		return nil
	})
}
