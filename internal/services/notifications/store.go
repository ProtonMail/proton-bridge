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

package notifications

import (
	"crypto"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/ProtonMail/go-proton-api"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/maps"
)

// not-const so we can unit test the functionality.
var timeOffset = 24 * time.Hour //nolint:gochecknoglobals
const filename = "notification_cache"

type Store struct {
	displayedMessages     map[string]time.Time
	displayedMessagesLock sync.Mutex

	useCache      bool
	cacheFilepath string
	cacheLock     sync.Mutex

	log *logrus.Entry
}

func NewStore(getCachePath func() (string, error)) *Store {
	log := logrus.WithField("pkg", "notification-store")

	useCacheFile := true
	cachePath, err := getCachePath()
	if err != nil {
		useCacheFile = false
		log.WithError(err).Error("Could not obtain cache directory")
	}

	store := &Store{
		displayedMessages: make(map[string]time.Time),

		useCache:      useCacheFile,
		cacheFilepath: filepath.Clean(filepath.Join(cachePath, filename)),

		log: log,
	}

	store.readCache()

	return store
}

func generateHash(payload proton.NotificationPayload) string {
	hash := crypto.SHA256.New()
	hash.Write([]byte(payload.Body + payload.Subtitle + payload.Title))
	return hex.EncodeToString(hash.Sum(nil))
}

func (s *Store) shouldSendAndStore(notification proton.NotificationEvent) bool {
	s.displayedMessagesLock.Lock()
	defer s.displayedMessagesLock.Unlock()

	// \todo BRIDGE-141 - Add an additional check for the API returned UID
	uid := generateHash(notification.Payload)

	value, ok := s.displayedMessages[uid]
	if !ok {
		s.displayedMessages[uid] = time.Unix(notification.Time, 0).Add(timeOffset)
		s.writeCache()
		return true
	}

	if !time.Now().After(value) {
		return false
	}

	s.displayedMessages[uid] = time.Unix(notification.Time, 0).Add(timeOffset)
	s.writeCache()
	return true
}

func (s *Store) readCache() {
	if !s.useCache {
		return
	}

	s.cacheLock.Lock()
	defer s.cacheLock.Unlock()

	file, err := os.Open(s.cacheFilepath)
	if err != nil {
		s.log.WithError(err).Info("Unable to open cache file")
		return
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			s.log.WithError(err).Error("Unable to close cache file after read")
		}
	}(file)

	s.displayedMessagesLock.Lock()
	defer s.displayedMessagesLock.Unlock()
	if err = json.NewDecoder(file).Decode(&s.displayedMessages); err != nil {
		s.log.WithError(err).Error("Unable to decode cache file")
	}

	// Remove redundant data
	curTime := time.Now()
	maps.DeleteFunc(s.displayedMessages, func(_ string, value time.Time) bool {
		return curTime.After(value)
	})
}

func (s *Store) writeCache() {
	if !s.useCache {
		return
	}

	s.cacheLock.Lock()
	defer s.cacheLock.Unlock()

	file, err := os.Create(s.cacheFilepath)
	if err != nil {
		s.log.WithError(err).Info("Unable to create cache file.")
		return
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			s.log.WithError(err).Error("Unable to close cache file after write")
		}
	}(file)

	// We don't lock the mutex here as the parent does that already
	if err = json.NewEncoder(file).Encode(s.displayedMessages); err != nil {
		s.log.WithError(err).Error("Unable to encode data to cache file")
	}
}
