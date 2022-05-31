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
	"strings"

	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
	"github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"
)

// AddressInfo is the format of the data held in the addresses bucket in the store.
// It allows us to easily keep an address and its ID together and serialisation/deserialisation to []byte.
type AddressInfo struct {
	Address, AddressID string
}

// GetAddressID returns the ID of the given address.
func (store *Store) GetAddressID(addr string) (id string, err error) {
	addrs, err := store.GetAddressInfo()
	if err != nil {
		return
	}

	for _, addrInfo := range addrs {
		if strings.EqualFold(addrInfo.Address, addr) {
			id = addrInfo.AddressID
			return
		}
	}

	err = errors.New("no such address")

	return
}

// GetAddressInfo returns information about all addresses owned by the user.
// The first element is the user's primary address and the rest (if present) are aliases.
// It tries to source the information from the store but if the store doesn't yet have it, it
// fetches it from the API and caches it for later.
func (store *Store) GetAddressInfo() (addrs []AddressInfo, err error) {
	if addrs, err = store.getAddressInfoFromStore(); err == nil && len(addrs) > 0 {
		return
	}

	// Store does not have address info yet, need to build it first from API.
	addressList := store.client().Addresses()
	if addressList == nil {
		err = errors.New("addresses unavailable")
		store.log.WithError(err).Error("Could not get user addresses from API")
		return
	}

	if err = store.createOrUpdateAddressInfo(addressList); err != nil {
		store.log.WithError(err).Warn("Could not update address IDs in store")
		return
	}

	return store.getAddressInfoFromStore()
}

// getAddressIDsByAddressFromStore returns a map from address to addressID for each address owned by the user.
func (store *Store) getAddressInfoFromStore() (addrs []AddressInfo, err error) {
	store.log.Debug("Retrieving address info from store")

	tx := func(tx *bolt.Tx) (err error) {
		c := tx.Bucket(addressInfoBucket).Cursor()
		for index, addrInfoBytes := c.First(); index != nil; index, addrInfoBytes = c.Next() {
			var addrInfo AddressInfo

			if err = json.Unmarshal(addrInfoBytes, &addrInfo); err != nil {
				store.log.WithError(err).Error("Could not unmarshal address and addressID")
				return
			}

			addrs = append(addrs, addrInfo)
		}

		return
	}

	err = store.db.View(tx)

	return
}

// createOrUpdateAddressInfo updates the store address/addressID bucket to match the given address list.
// The address list supplied is assumed to contain active emails in any order.
// It firstly (and stupidly) deletes the bucket of addresses and then fills it with up to date info.
// This is because a user might delete an address and we don't want old addresses lying around (and finding the
// specific ones to delete is likely not much more efficient than just rebuilding from scratch).
func (store *Store) createOrUpdateAddressInfo(addressList pmapi.AddressList) (err error) {
	tx := func(tx *bolt.Tx) error {
		if err := tx.DeleteBucket(addressInfoBucket); err != nil {
			store.log.WithError(err).Error("Could not delete addressIDs bucket")
			return err
		}

		if _, err := tx.CreateBucketIfNotExists(addressInfoBucket); err != nil {
			store.log.WithError(err).Error("Could not recreate addressIDs bucket")
			return err
		}

		addrsBucket := tx.Bucket(addressInfoBucket)

		for index, address := range filterAddresses(addressList) {
			ib := itob(uint32(index))

			info, err := json.Marshal(AddressInfo{
				Address:   address.Email,
				AddressID: address.ID,
			})
			if err != nil {
				store.log.WithError(err).Error("Could not marshal address and addressID")
				return err
			}

			if err := addrsBucket.Put(ib, info); err != nil {
				store.log.WithError(err).Error("Could not put address and addressID into store")
				return err
			}
		}

		return nil
	}

	return store.db.Update(tx)
}

// filterAddresses filters out inactive addresses and ensures the original address is listed first.
func filterAddresses(addressList pmapi.AddressList) (filteredList pmapi.AddressList) {
	for _, address := range addressList {
		if !address.Receive {
			continue
		}

		filteredList = append(filteredList, address)
	}

	return
}
