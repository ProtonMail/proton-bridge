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
	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
	"github.com/sirupsen/logrus"
	bolt "go.etcd.io/bbolt"
)

// Address holds mailboxes for IMAP user (login address). In combined mode
// there is only one address, in split mode there is one object per address.
type Address struct {
	store     *Store
	address   string
	addressID string
	mailboxes map[string]*Mailbox

	log *logrus.Entry
}

func newAddress(
	store *Store,
	address, addressID string,
	labels []*pmapi.Label,
) (addr *Address, err error) {
	l := log.WithField("addressID", addressID)

	storeAddress := &Address{
		store:     store,
		address:   address,
		addressID: addressID,
		log:       l,
	}

	if err = storeAddress.init(labels); err != nil {
		l.WithField("address", address).
			WithError(err).
			Error("Could not initialise store address")

		return
	}

	return storeAddress, nil
}

func (storeAddress *Address) init(foldersAndLabels []*pmapi.Label) (err error) {
	storeAddress.log.WithField("address", storeAddress.address).Debug("Initialising store address")

	storeAddress.mailboxes = make(map[string]*Mailbox)

	err = storeAddress.store.db.Update(func(tx *bolt.Tx) error {
		for _, label := range foldersAndLabels {
			prefix := getLabelPrefix(label)

			var mailbox *Mailbox
			if mailbox, err = txNewMailbox(tx, storeAddress, label.ID, prefix, label.Path, label.Color); err != nil {
				storeAddress.log.
					WithError(err).
					WithField("labelID", label.ID).
					Error("Could not init mailbox for folder or label")
				return err
			}

			storeAddress.mailboxes[label.ID] = mailbox
		}
		return nil
	})

	return
}

// getLabelPrefix returns the correct prefix for a pmapi label according to whether it is exclusive or not.
func getLabelPrefix(l *pmapi.Label) string {
	switch {
	case pmapi.IsSystemLabel(l.ID):
		return ""
	case bool(l.Exclusive):
		return UserFoldersPrefix
	default:
		return UserLabelsPrefix
	}
}

// AddressString returns the address.
func (storeAddress *Address) AddressString() string {
	return storeAddress.address
}

// AddressID returns the address ID.
func (storeAddress *Address) AddressID() string {
	return storeAddress.addressID
}

// APIAddress returns the `pmapi.Address` struct.
func (storeAddress *Address) APIAddress() *pmapi.Address {
	return storeAddress.client().Addresses().ByEmail(storeAddress.address)
}

func (storeAddress *Address) client() pmapi.Client {
	return storeAddress.store.client()
}
