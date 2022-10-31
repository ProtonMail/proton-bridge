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
	bolt "go.etcd.io/bbolt"
)

func (storeAddress *Address) txCreateOrUpdateMessages(tx *bolt.Tx, msgs []*pmapi.Message) error {
	for _, m := range storeAddress.mailboxes {
		if err := m.txCreateOrUpdateMessages(tx, msgs); err != nil {
			return err
		}
	}
	return nil
}

// txDeleteMessage deletes the message from the mailbox buckets for this address.
func (storeAddress *Address) txDeleteMessage(tx *bolt.Tx, apiID string) error {
	for _, m := range storeAddress.mailboxes {
		if err := m.txDeleteMessage(tx, apiID); err != nil {
			return err
		}
	}
	return nil
}
