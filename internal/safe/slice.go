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

package safe

import (
	"sync"

	"github.com/bradenaw/juniper/xslices"
)

type Slice[Val comparable] struct {
	data []Val
	lock sync.RWMutex
}

func NewSlice[Val comparable](from ...Val) *Slice[Val] {
	s := &Slice[Val]{
		data: make([]Val, len(from)),
	}

	copy(s.data, from)

	return s
}

func (s *Slice[Val]) Iter(fn func(val Val)) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	for _, val := range s.data {
		fn(val)
	}
}

func (s *Slice[Val]) Append(val Val) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.data = append(s.data, val)
}

func (s *Slice[Val]) Delete(val Val) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.data = xslices.Filter(s.data, func(v Val) bool {
		return v != val
	})
}
