// Copyright (c) 2021 Proton Technologies AG
//
// This file is part of ProtonMail Bridge.
//
// ProtonMail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ProtonMail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ProtonMail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package store

import (
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/pkg/message"
	"github.com/sirupsen/logrus"
	bolt "go.etcd.io/bbolt"
)

const passphraseKey = "passphrase"

// UnlockCache unlocks the cache for the user with the given keyring.
func (store *Store) UnlockCache(kr *crypto.KeyRing) error {
	passphrase, err := store.getCachePassphrase()
	if err != nil {
		return err
	}

	if passphrase == nil {
		if passphrase, err = crypto.RandomToken(32); err != nil {
			return err
		}

		enc, err := kr.Encrypt(crypto.NewPlainMessage(passphrase), nil)
		if err != nil {
			return err
		}

		if err := store.setCachePassphrase(enc.GetBinary()); err != nil {
			return err
		}
	} else {
		dec, err := kr.Decrypt(crypto.NewPGPMessage(passphrase), nil, crypto.GetUnixTime())
		if err != nil {
			return err
		}

		passphrase = dec.GetBinary()
	}

	if err := store.cache.Unlock(store.user.ID(), passphrase); err != nil {
		return err
	}

	store.cacher.start()

	return nil
}

func (store *Store) getCachePassphrase() ([]byte, error) {
	var passphrase []byte

	if err := store.db.View(func(tx *bolt.Tx) error {
		passphrase = tx.Bucket(cachePassphraseBucket).Get([]byte(passphraseKey))
		return nil
	}); err != nil {
		return nil, err
	}

	return passphrase, nil
}

func (store *Store) setCachePassphrase(passphrase []byte) error {
	return store.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(cachePassphraseBucket).Put([]byte(passphraseKey), passphrase)
	})
}

func (store *Store) clearCachePassphrase() error {
	return store.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(cachePassphraseBucket).Delete([]byte(passphraseKey))
	})
}

func (store *Store) getCachedMessage(messageID string) ([]byte, error) {
	if store.cache.Has(store.user.ID(), messageID) {
		return store.cache.Get(store.user.ID(), messageID)
	}

	job, done := store.newBuildJob(messageID, message.ForegroundPriority)
	defer done()

	literal, err := job.GetResult()
	if err != nil {
		return nil, err
	}

	// NOTE(GODT-1158): No need to block until cache has been set; do this async?
	if err := store.cache.Set(store.user.ID(), messageID, literal); err != nil {
		logrus.WithError(err).Error("Failed to cache message")
	}

	return literal, nil
}

// IsCached returns whether the given message already exists in the cache.
func (store *Store) IsCached(messageID string) bool {
	return store.cache.Has(store.user.ID(), messageID)
}

// BuildAndCacheMessage builds the given message (with background priority) and puts it in the cache.
// It builds with background priority.
func (store *Store) BuildAndCacheMessage(messageID string) error {
	job, done := store.newBuildJob(messageID, message.BackgroundPriority)
	defer done()

	literal, err := job.GetResult()
	if err != nil {
		return err
	}

	return store.cache.Set(store.user.ID(), messageID, literal)
}
