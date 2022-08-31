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

import "sync"

type safeMapOfStrings struct {
	data  map[string]string
	mutex sync.RWMutex
}

func newSafeMapOfString() safeMapOfStrings {
	return safeMapOfStrings{
		data:  map[string]string{},
		mutex: sync.RWMutex{},
	}
}

func (m *safeMapOfStrings) get(key string) string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.data[key]
}

func (m *safeMapOfStrings) set(key, value string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.data[key] = value
}
