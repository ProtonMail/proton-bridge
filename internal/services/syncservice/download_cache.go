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

package syncservice

import (
	"sync"

	"github.com/ProtonMail/go-proton-api"
)

type DownloadCache struct {
	messageLock    sync.RWMutex
	messages       map[string]proton.Message
	attachmentLock sync.RWMutex
	attachments    map[string][]byte
}

func newDownloadCache() *DownloadCache {
	return &DownloadCache{
		messages:    make(map[string]proton.Message, 64),
		attachments: make(map[string][]byte, 64),
	}
}

func (s *DownloadCache) StoreMessage(message proton.Message) {
	s.messageLock.Lock()
	defer s.messageLock.Unlock()

	s.messages[message.ID] = message
}

func (s *DownloadCache) StoreAttachment(id string, data []byte) {
	s.attachmentLock.Lock()
	defer s.attachmentLock.Unlock()

	s.attachments[id] = data
}

func (s *DownloadCache) DeleteMessages(id ...string) {
	s.messageLock.Lock()
	defer s.messageLock.Unlock()

	for _, id := range id {
		delete(s.messages, id)
	}
}

func (s *DownloadCache) DeleteAttachments(id ...string) {
	s.attachmentLock.Lock()
	defer s.attachmentLock.Unlock()

	for _, id := range id {
		delete(s.attachments, id)
	}
}

func (s *DownloadCache) GetMessage(id string) (proton.Message, bool) {
	s.messageLock.RLock()
	defer s.messageLock.RUnlock()

	v, ok := s.messages[id]

	return v, ok
}

func (s *DownloadCache) GetAttachment(id string) ([]byte, bool) {
	s.attachmentLock.RLock()
	defer s.attachmentLock.RUnlock()

	v, ok := s.attachments[id]

	return v, ok
}

func (s *DownloadCache) Clear() {
	s.messageLock.Lock()
	s.messages = make(map[string]proton.Message, 64)
	s.messageLock.Unlock()

	s.attachmentLock.Lock()
	s.attachments = make(map[string][]byte, 64)
	s.attachmentLock.Unlock()
}

func (s *DownloadCache) Count() (int, int) {
	var (
		messageCount    int
		attachmentCount int
	)

	s.messageLock.Lock()
	messageCount = len(s.messages)
	s.messageLock.Unlock()

	s.attachmentLock.Lock()
	attachmentCount = len(s.attachments)
	s.attachmentLock.Unlock()

	return messageCount, attachmentCount
}
