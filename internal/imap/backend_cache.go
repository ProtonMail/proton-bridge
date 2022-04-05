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

package imap

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
)

// Cache keys.
const (
	SubscriptionException = "subscription_exceptions"
)

// addToCache adds item to existing item list.
// Starting from following structure:
//   {
//		"username": {"label": "item1;item2"}
//   }
//
// After calling addToCache("username", "label", "newItem") we get:
//   {
//		"username": {"label": "item1;item2;newItem"}
//   }
//
func (ib *imapBackend) addToCache(userID, label, toAdd string) {
	list := ib.getCacheList(userID, label)

	if list != "" {
		list = list + ";" + toAdd
	} else {
		list = toAdd
	}

	ib.imapCacheLock.Lock()
	ib.imapCache[userID][label] = list
	ib.imapCacheLock.Unlock()

	if err := ib.saveIMAPCache(); err != nil {
		log.Info("Backend/userinfo: could not save cache: ", err)
	}
}

func (ib *imapBackend) removeFromCache(userID, label, toRemove string) {
	list := ib.getCacheList(userID, label)

	split := strings.Split(list, ";")

	for i, item := range split {
		if item == toRemove {
			split = append(split[:i], split[i+1:]...)
		}
	}

	ib.imapCacheLock.Lock()
	ib.imapCache[userID][label] = strings.Join(split, ";")
	ib.imapCacheLock.Unlock()

	if err := ib.saveIMAPCache(); err != nil {
		log.Info("Backend/userinfo: could not save cache: ", err)
	}
}

func (ib *imapBackend) getCacheList(userID, label string) (list string) {
	if err := ib.loadIMAPCache(); err != nil {
		log.WithError(err).Warn("Could not load cache")
	}

	ib.imapCacheLock.Lock()
	if ib.imapCache == nil {
		ib.imapCache = map[string]map[string]string{}
	}

	if ib.imapCache[userID] == nil {
		ib.imapCache[userID] = map[string]string{}
		ib.imapCache[userID][SubscriptionException] = ""
	}

	list = ib.imapCache[userID][label]

	ib.imapCacheLock.Unlock()

	if err := ib.saveIMAPCache(); err != nil {
		log.WithError(err).Warn("Could not save cache")
	}
	return
}

func (ib *imapBackend) loadIMAPCache() error {
	if ib.imapCache != nil {
		return nil
	}

	ib.imapCacheLock.Lock()
	defer ib.imapCacheLock.Unlock()

	f, err := os.Open(ib.imapCachePath)
	if err != nil {
		return err
	}
	defer f.Close() //nolint:errcheck,gosec

	return json.NewDecoder(f).Decode(&ib.imapCache)
}

func (ib *imapBackend) saveIMAPCache() error {
	if ib.imapCache == nil {
		return errors.New("cannot save cache: cache is nil")
	}

	ib.imapCacheLock.Lock()
	defer ib.imapCacheLock.Unlock()

	f, err := os.Create(ib.imapCachePath)
	if err != nil {
		return err
	}
	defer f.Close() //nolint:errcheck,gosec

	return json.NewEncoder(f).Encode(ib.imapCache)
}
