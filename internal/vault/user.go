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
// along with Proton Mail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package vault

import (
	"fmt"

	"github.com/ProtonMail/gluon/imap"
)

type User struct {
	vault  *Vault
	userID string
}

func (user *User) UserID() string {
	return user.vault.getUser(user.userID).UserID
}

func (user *User) Username() string {
	return user.vault.getUser(user.userID).Username
}

// GluonKey returns the key needed to decrypt the user's gluon database.
func (user *User) GluonKey() []byte {
	return user.vault.getUser(user.userID).GluonKey
}

func (user *User) GetGluonIDs() map[string]string {
	return user.vault.getUser(user.userID).GluonIDs
}

func (user *User) SetGluonID(addrID, gluonID string) error {
	return user.vault.modUser(user.userID, func(data *UserData) {
		data.GluonIDs[addrID] = gluonID
	})
}

func (user *User) RemoveGluonID(addrID, gluonID string) error {
	var err error

	if modErr := user.vault.modUser(user.userID, func(data *UserData) {
		if data.GluonIDs[addrID] != gluonID {
			err = fmt.Errorf("gluon ID mismatch: %s != %s", data.GluonIDs[addrID], gluonID)
		} else {
			delete(data.GluonIDs, addrID)
		}
	}); modErr != nil {
		return modErr
	}

	return err
}

func (user *User) GetUIDValidity(addrID string) (imap.UID, bool) {
	validity, ok := user.vault.getUser(user.userID).UIDValidity[addrID]
	if !ok {
		return imap.UID(0), false
	}

	return validity, true
}

func (user *User) SetUIDValidity(addrID string, validity imap.UID) error {
	return user.vault.modUser(user.userID, func(data *UserData) {
		data.UIDValidity[addrID] = validity
	})
}

// AddressMode returns the user's address mode.
func (user *User) AddressMode() AddressMode {
	return user.vault.getUser(user.userID).AddressMode
}

// SetAddressMode sets the address mode for the given user.
func (user *User) SetAddressMode(mode AddressMode) error {
	return user.vault.modUser(user.userID, func(data *UserData) {
		data.AddressMode = mode
	})
}

// BridgePass returns the user's bridge password (unencoded).
func (user *User) BridgePass() []byte {
	return user.vault.getUser(user.userID).BridgePass
}

// AuthUID returns the user's auth UID.
func (user *User) AuthUID() string {
	return user.vault.getUser(user.userID).AuthUID
}

// AuthRef returns the user's auth refresh token.
func (user *User) AuthRef() string {
	return user.vault.getUser(user.userID).AuthRef
}

// SetAuth sets the auth secrets for the given user.
func (user *User) SetAuth(authUID, authRef string) error {
	return user.vault.modUser(user.userID, func(data *UserData) {
		data.AuthUID = authUID
		data.AuthRef = authRef
	})
}

// KeyPass returns the user's (salted) key password.
func (user *User) KeyPass() []byte {
	return user.vault.getUser(user.userID).KeyPass
}

// SetKeyPass sets the user's (salted) key password.
func (user *User) SetKeyPass(keyPass []byte) error {
	return user.vault.modUser(user.userID, func(data *UserData) {
		data.KeyPass = keyPass
	})
}

// SyncStatus return's the user's sync status.
func (user *User) SyncStatus() SyncStatus {
	return user.vault.getUser(user.userID).SyncStatus
}

// SetHasLabels sets whether the user's labels have been synced.
func (user *User) SetHasLabels(hasLabels bool) error {
	return user.vault.modUser(user.userID, func(data *UserData) {
		data.SyncStatus.HasLabels = hasLabels
	})
}

// SetHasMessages sets whether the user's messages have been synced.
func (user *User) SetHasMessages(hasMessages bool) error {
	return user.vault.modUser(user.userID, func(data *UserData) {
		data.SyncStatus.HasMessages = hasMessages
	})
}

// SetLastMessageID sets the last synced message ID for the given user.
func (user *User) SetLastMessageID(messageID string) error {
	return user.vault.modUser(user.userID, func(data *UserData) {
		data.SyncStatus.LastMessageID = messageID
	})
}

// ClearSyncStatus clears the user's sync status.
func (user *User) ClearSyncStatus() error {
	return user.vault.modUser(user.userID, func(data *UserData) {
		data.SyncStatus = SyncStatus{}
		data.EventID = ""
	})
}

// EventID returns the last processed event ID of the user.
func (user *User) EventID() string {
	return user.vault.getUser(user.userID).EventID
}

// SetEventID sets the event ID for the given user.
func (user *User) SetEventID(eventID string) error {
	return user.vault.modUser(user.userID, func(data *UserData) {
		data.EventID = eventID
	})
}

// Clear clears the user's auth secrets.
func (user *User) Clear() error {
	return user.vault.modUser(user.userID, func(data *UserData) {
		data.AuthUID = ""
		data.AuthRef = ""
		data.KeyPass = nil
	})
}

// Close closes the user. This allows it to be removed from the vault.
func (user *User) Close() error {
	return user.vault.detachUser(user.userID)
}
