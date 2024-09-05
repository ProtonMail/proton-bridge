// Copyright (c) 2024 Proton AG
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
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	"github.com/ProtonMail/gluon/async"
	"github.com/bradenaw/juniper/parallel"
	"github.com/bradenaw/juniper/xslices"
	"github.com/sirupsen/logrus"
)

// Vault is an encrypted data vault that stores bridge and user data.
type Vault struct {
	path string
	gcm  cipher.AEAD

	enc []byte

	ref map[string]int

	lock sync.RWMutex

	panicHandler async.PanicHandler
}

var ErrDecryptFailed = errors.New("failed to decrypt vault")
var ErrUnmarshal = errors.New("vault contents are corrupt")

// New constructs a new encrypted data vault at the given filepath using the given encryption key.
// The first error is a corruption error for an existing vault, the second errors refrain to all other errors.
func New(vaultDir, gluonCacheDir string, key []byte, panicHandler async.PanicHandler) (*Vault, error, error) {
	if err := os.MkdirAll(vaultDir, 0o700); err != nil {
		return nil, nil, err
	}

	hash256 := sha256.Sum256(key)

	aes, err := aes.NewCipher(hash256[:])
	if err != nil {
		return nil, nil, err
	}

	gcm, err := cipher.NewGCM(aes)
	if err != nil {
		return nil, nil, err
	}

	vault, corrupt, err := newVault(filepath.Join(vaultDir, "vault.enc"), gluonCacheDir, gcm)
	if err != nil {
		return nil, corrupt, err
	}

	vault.panicHandler = panicHandler

	return vault, corrupt, nil
}

// GetUserIDs returns the user IDs and usernames of all users in the vault.
func (vault *Vault) GetUserIDs() []string {
	vault.lock.RLock()
	defer vault.lock.RUnlock()

	return xslices.Map(vault.getUnsafe().Users, func(user UserData) string {
		return user.UserID
	})
}

func (vault *Vault) getUsers() ([]*User, error) {
	vault.lock.Lock()
	defer vault.lock.Unlock()

	users := vault.getUnsafe().Users

	result := make([]*User, 0, len(users))

	for _, user := range users {
		u, err := vault.newUserUnsafe(user.UserID)
		if err != nil {
			for _, v := range result {
				if err := v.Close(); err != nil {
					logrus.WithError(err).Error("Fait to close user after failed get")
				}
			}

			return nil, err
		}

		result = append(result, u)
	}

	return result, nil
}

// HasUser returns true if the vault contains a user with the given ID.
func (vault *Vault) HasUser(userID string) bool {
	vault.lock.RLock()
	defer vault.lock.RUnlock()

	return xslices.IndexFunc(vault.getUnsafe().Users, func(user UserData) bool {
		return user.UserID == userID
	}) >= 0
}

// GetUser provides access to a vault user. It returns an error if the user does not exist.
func (vault *Vault) GetUser(userID string, fn func(*User)) error {
	user, err := vault.NewUser(userID)
	if err != nil {
		return err
	}
	defer func() { _ = user.Close() }()

	fn(user)

	return nil
}

// NewUser returns a new vault user. It must be closed before it can be deleted.
func (vault *Vault) NewUser(userID string) (*User, error) {
	vault.lock.Lock()
	defer vault.lock.Unlock()

	return vault.newUserUnsafe(userID)
}

func (vault *Vault) newUserUnsafe(userID string) (*User, error) {
	if idx := xslices.IndexFunc(vault.getUnsafe().Users, func(user UserData) bool {
		return user.UserID == userID
	}); idx < 0 {
		return nil, errors.New("no such user")
	}

	return vault.attachUserUnsafe(userID), nil
}

// ForUser executes a callback for each user in the vault.
func (vault *Vault) ForUser(parallelism int, fn func(*User) error) error {
	users, err := vault.getUsers()
	if err != nil {
		return err
	}

	r := parallel.DoContext(context.Background(), parallelism, len(users), func(_ context.Context, idx int) error {
		defer async.HandlePanic(vault.panicHandler)

		user := users[idx]
		return fn(user)
	})

	for _, u := range users {
		if err := u.Close(); err != nil {
			logrus.WithError(err).Error("Failed to close user after ForUser")
		}
	}

	return r
}

// AddUser creates a new user in the vault with the given ID, username and password.
// A gluon key is generated using the package's token generator. If a password is found in the password archive for this user,
// it is restored, otherwise a new bridge password is generated using the package's token generator.
func (vault *Vault) AddUser(userID, username, primaryEmail, authUID, authRef string, keyPass []byte) (*User, error) {
	vault.lock.Lock()
	defer vault.lock.Unlock()

	return vault.addUserUnsafe(userID, username, primaryEmail, authUID, authRef, keyPass)
}

func (vault *Vault) addUserUnsafe(userID, username, primaryEmail, authUID, authRef string, keyPass []byte) (*User, error) {
	logrus.WithField("userID", userID).Info("Adding vault user")

	var exists bool

	if err := vault.modUnsafe(func(data *Data) {
		if idx := xslices.IndexFunc(data.Users, func(user UserData) bool {
			return user.UserID == userID
		}); idx >= 0 {
			exists = true
		} else {
			bridgePass := data.Settings.PasswordArchive.get(primaryEmail)
			if len(bridgePass) == 0 {
				bridgePass = newRandomToken(16)
			}

			data.Users = append(data.Users, newDefaultUser(userID, username, primaryEmail, authUID, authRef, keyPass, bridgePass))
		}
	}); err != nil {
		return nil, err
	}

	if exists {
		return nil, errors.New("user already exists")
	}

	return vault.attachUserUnsafe(userID), nil
}

// GetOrAddUser retrieves an existing user and updates the authRef and keyPass or creates a new user. Returns
// the user and whether the user did not exist before.
func (vault *Vault) GetOrAddUser(userID, username, primaryEmail, authUID, authRef string, keyPass []byte) (*User, bool, error) {
	vault.lock.Lock()
	defer vault.lock.Unlock()

	{
		users := vault.getUnsafe().Users

		idx := xslices.IndexFunc(users, func(user UserData) bool {
			return user.UserID == userID
		})

		if idx >= 0 {
			user := vault.attachUserUnsafe(userID)

			if err := user.setAuthAndKeyPassUnsafe(authUID, authRef, keyPass); err != nil {
				return nil, false, err
			}

			if err := user.updateUsernameUnsafe(username); err != nil {
				return nil, false, err
			}

			return user, false, nil
		}
	}

	u, err := vault.addUserUnsafe(userID, username, primaryEmail, authUID, authRef, keyPass)

	return u, true, err
}

// DeleteUser removes the given user from the vault.
func (vault *Vault) DeleteUser(userID string) error {
	vault.lock.Lock()
	defer vault.lock.Unlock()

	logrus.WithField("userID", userID).Info("Deleting vault user")

	if _, ok := vault.ref[userID]; ok {
		return fmt.Errorf("user %s is currently in use", userID)
	}

	return vault.modUnsafe(func(data *Data) {
		idx := xslices.IndexFunc(data.Users, func(user UserData) bool {
			return user.UserID == userID
		})

		if idx < 0 {
			return
		}
		data.Settings.PasswordArchive.set(data.Users[idx].PrimaryEmail, data.Users[idx].BridgePass)
		data.Users = append(data.Users[:idx], data.Users[idx+1:]...)
	})
}

func (vault *Vault) Migrated() bool {
	vault.lock.RLock()
	defer vault.lock.RUnlock()

	return vault.getUnsafe().Migrated
}

func (vault *Vault) SetMigrated() error {
	vault.lock.Lock()
	defer vault.lock.Unlock()

	return vault.modUnsafe(func(data *Data) {
		data.Migrated = true
	})
}

func (vault *Vault) Reset(gluonDir string) error {
	vault.lock.Lock()
	defer vault.lock.Unlock()

	return vault.modUnsafe(func(data *Data) {
		*data = newDefaultData(gluonDir)
	})
}

func (vault *Vault) Path() string {
	return vault.path
}

func (vault *Vault) Close() error {
	vault.lock.Lock()
	defer vault.lock.Unlock()

	if len(vault.ref) > 0 {
		return errors.New("vault is still in use")
	}

	vault.gcm = nil

	return nil
}

func (vault *Vault) attachUserUnsafe(userID string) *User {
	logrus.WithField("userID", userID).Trace("Attaching vault user")

	vault.ref[userID]++

	return &User{
		vault:  vault,
		userID: userID,
	}
}

func (vault *Vault) detachUser(userID string) error {
	vault.lock.Lock()
	defer vault.lock.Unlock()

	logrus.WithField("userID", userID).Trace("Detaching vault user")

	if _, ok := vault.ref[userID]; !ok {
		return fmt.Errorf("user %s is not attached", userID)
	}

	vault.ref[userID]--

	if vault.ref[userID] == 0 {
		delete(vault.ref, userID)
	}

	return nil
}

func newVault(path, gluonDir string, gcm cipher.AEAD) (*Vault, error, error) {
	if _, err := os.Stat(path); errors.Is(err, fs.ErrNotExist) {
		if _, err := initVault(path, gluonDir, gcm); err != nil {
			return nil, nil, err
		}
	}

	enc, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, nil, err
	}

	var corrupt error

	if err := unmarshalFile(gcm, enc, new(Data)); err != nil {
		corrupt = err
	}

	if corrupt != nil {
		newEnc, err := initVault(path, gluonDir, gcm)
		if err != nil {
			return nil, corrupt, err
		}

		enc = newEnc
	}

	return &Vault{
		path: path,
		enc:  enc,
		gcm:  gcm,
		ref:  make(map[string]int),
	}, corrupt, nil
}

func (vault *Vault) getSafe() Data {
	vault.lock.RLock()
	defer vault.lock.RUnlock()

	return vault.getUnsafe()
}

func (vault *Vault) getUnsafe() Data {
	var data Data

	if err := unmarshalFile(vault.gcm, vault.enc, &data); err != nil {
		panic(err)
	}

	return data
}

func (vault *Vault) modSafe(fn func(data *Data)) error {
	vault.lock.Lock()
	defer vault.lock.Unlock()

	return vault.modUnsafe(fn)
}

func (vault *Vault) modUnsafe(fn func(data *Data)) error {
	var data Data

	if err := unmarshalFile(vault.gcm, vault.enc, &data); err != nil {
		return err
	}

	fn(&data)

	enc, err := marshalFile(vault.gcm, data)
	if err != nil {
		return err
	}

	vault.enc = enc

	tmpFile := vault.path + ".tmp"

	if err := os.WriteFile(tmpFile, vault.enc, 0o600); err != nil {
		return fmt.Errorf("failed write new vault to disk: %w", err)
	}

	if err := os.Rename(tmpFile, vault.path); err != nil {
		return fmt.Errorf("failed to overwrite old vault data: %w", err)
	}

	return nil
}

func (vault *Vault) getUser(userID string) UserData {
	vault.lock.RLock()
	defer vault.lock.RUnlock()

	users := vault.getUnsafe().Users

	idx := xslices.IndexFunc(users, func(user UserData) bool {
		return user.UserID == userID
	})

	if idx < 0 {
		panic("Unknown user")
	}

	return users[idx]
}

// getUserUnsafe - fetches the relevant UserData.
// Should only be called from contexts in which the vault mutex has been read locked.
func (vault *Vault) getUserUnsafe(userID string) UserData {
	users := vault.getUnsafe().Users

	idx := xslices.IndexFunc(users, func(user UserData) bool {
		return user.UserID == userID
	})

	if idx < 0 {
		panic("Unknown user")
	}

	return users[idx]
}

func (vault *Vault) modUser(userID string, fn func(userData *UserData)) error {
	vault.lock.Lock()
	defer vault.lock.Unlock()

	return vault.modUserUnsafe(userID, fn)
}

func (vault *Vault) modUserUnsafe(userID string, fn func(userData *UserData)) error {
	return vault.modUnsafe(func(data *Data) {
		idx := xslices.IndexFunc(data.Users, func(user UserData) bool {
			return user.UserID == userID
		})

		fn(&data.Users[idx])
	})
}

func initVault(path, gluonDir string, gcm cipher.AEAD) ([]byte, error) {
	enc, err := marshalFile(gcm, newDefaultData(gluonDir))
	if err != nil {
		return nil, err
	}

	if err := os.WriteFile(path, enc, 0o600); err != nil {
		return nil, err
	}

	return enc, nil
}
