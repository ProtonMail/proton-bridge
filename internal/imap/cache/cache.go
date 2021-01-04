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

package cache

import (
	"bytes"
	"sort"
	"sync"
	"time"

	pkgMsg "github.com/ProtonMail/proton-bridge/pkg/message"
)

type key struct {
	ID        string
	Timestamp int64
	Size      int
}

type oldestFirst []key

func (s oldestFirst) Len() int           { return len(s) }
func (s oldestFirst) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s oldestFirst) Less(i, j int) bool { return s[i].Timestamp < s[j].Timestamp }

type cachedMessage struct {
	key
	data      []byte
	structure pkgMsg.BodyStructure
}

//nolint[gochecknoglobals]
var (
	cacheTimeLimit = int64(1 * 60 * 60 * 1000) // milliseconds
	cacheSizeLimit = 100 * 1000 * 1000         // B - MUST be larger than email max size limit (~ 25 MB)
	mailCache      = make(map[string]cachedMessage)

	// cacheMutex takes care of one single operation, whereas buildMutex takes
	// care of the whole action doing multiple operations. buildMutex will protect
	// you from asking server or decrypting or building the same message more
	// than once. When first request to build the message comes, it will block
	// all other build requests. When the first one is done, all others are
	// handled by cache, not doing anything twice. With cacheMutex we are safe
	// only to not mess up with the cache, but we could end up downloading and
	// building message twice.
	cacheMutex = &sync.Mutex{}
	buildMutex = &sync.Mutex{}
	buildLocks = map[string]interface{}{}
)

func (m *cachedMessage) isValidOrDel() bool {
	if m.key.Timestamp+cacheTimeLimit < timestamp() {
		delete(mailCache, m.key.ID)
		return false
	}
	return true
}

func timestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func Clear() {
	mailCache = make(map[string]cachedMessage)
}

// BuildLock locks per message level, not on global level.
// Multiple different messages can be building at once.
func BuildLock(messageID string) {
	for {
		buildMutex.Lock()
		if _, ok := buildLocks[messageID]; ok { // if locked, wait
			buildMutex.Unlock()
			time.Sleep(10 * time.Millisecond)
		} else { // if unlocked, lock it
			buildLocks[messageID] = struct{}{}
			buildMutex.Unlock()
			return
		}
	}
}

func BuildUnlock(messageID string) {
	buildMutex.Lock()
	defer buildMutex.Unlock()
	delete(buildLocks, messageID)
}

func LoadMail(mID string) (reader *bytes.Reader, structure *pkgMsg.BodyStructure) {
	reader = &bytes.Reader{}
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	if message, ok := mailCache[mID]; ok && message.isValidOrDel() {
		reader = bytes.NewReader(message.data)
		structure = &message.structure

		// Update timestamp to keep emails which are used often.
		message.Timestamp = timestamp()
	}
	return
}

func SaveMail(mID string, msg []byte, structure *pkgMsg.BodyStructure) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	newMessage := cachedMessage{
		key: key{
			ID:        mID,
			Timestamp: timestamp(),
			Size:      len(msg),
		},
		data:      msg,
		structure: *structure,
	}

	// Remove old and reduce size.
	totalSize := 0
	messageList := []key{}
	for _, message := range mailCache {
		if message.isValidOrDel() {
			messageList = append(messageList, message.key)
			totalSize += message.key.Size
		}
	}
	sort.Sort(oldestFirst(messageList))
	var oldest key
	for totalSize+newMessage.key.Size >= cacheSizeLimit {
		oldest, messageList = messageList[0], messageList[1:]
		delete(mailCache, oldest.ID)
		totalSize -= oldest.Size
	}

	// Write new.
	mailCache[mID] = newMessage
}
