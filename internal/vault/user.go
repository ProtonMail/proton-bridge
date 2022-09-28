package vault

import (
	"encoding/hex"

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

func (user *User) GetGluonIDs() map[string]string {
	return user.vault.getUser(user.userID).GluonIDs
}

func (user *User) SetGluonID(addrID, gluonID string) error {
	return user.vault.modUser(user.userID, func(data *UserData) {
		data.GluonIDs[addrID] = gluonID
	})
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

func (user *User) GluonKey() []byte {
	return user.vault.getUser(user.userID).GluonKey
}

func (user *User) AddressMode() AddressMode {
	return user.vault.getUser(user.userID).AddressMode
}

func (user *User) BridgePass() string {
	return hex.EncodeToString(user.vault.getUser(user.userID).BridgePass)
}

func (user *User) AuthUID() string {
	return user.vault.getUser(user.userID).AuthUID
}

func (user *User) AuthRef() string {
	return user.vault.getUser(user.userID).AuthRef
}

func (user *User) KeyPass() []byte {
	return user.vault.getUser(user.userID).KeyPass
}

func (user *User) EventID() string {
	return user.vault.getUser(user.userID).EventID
}

func (user *User) HasSync() bool {
	return user.vault.getUser(user.userID).HasSync
}

func (user *User) SetKeyPass(keyPass []byte) error {
	return user.vault.modUser(user.userID, func(data *UserData) {
		data.KeyPass = keyPass
	})
}

// SetAuth sets the auth secrets for the given user.
func (user *User) SetAuth(authUID, authRef string) error {
	return user.vault.modUser(user.userID, func(data *UserData) {
		data.AuthUID = authUID
		data.AuthRef = authRef
	})
}

// SetAddressMode sets the address mode for the given user.
func (user *User) SetAddressMode(mode AddressMode) error {
	return user.vault.modUser(user.userID, func(data *UserData) {
		data.AddressMode = mode
	})
}

// SetEventID sets the event ID for the given user.
func (user *User) SetEventID(eventID string) error {
	return user.vault.modUser(user.userID, func(data *UserData) {
		data.EventID = eventID
	})
}

// SetSync sets the sync state for the given user.
func (user *User) SetSync(hasSync bool) error {
	return user.vault.modUser(user.userID, func(data *UserData) {
		data.HasSync = hasSync
	})
}
