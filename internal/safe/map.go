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
	"golang.org/x/exp/slices"
)

type Map[Key comparable, Val any] struct {
	data  map[Key]Val
	order []Key
	sort  func(a, b Key, data map[Key]Val) bool
	lock  sync.RWMutex
}

func NewMap[Key comparable, Val any](sort func(a, b Key, data map[Key]Val) bool) *Map[Key, Val] {
	return &Map[Key, Val]{
		data: make(map[Key]Val),
		sort: sort,
	}
}

func NewMapFrom[Key comparable, Val any](from map[Key]Val, sort func(a, b Key, data map[Key]Val) bool) *Map[Key, Val] {
	m := NewMap(sort)

	for key, val := range from {
		m.Set(key, val)
	}

	return m
}

func (m *Map[Key, Val]) Index(idx int, fn func(Key, Val)) bool {
	m.lock.RLock()
	defer m.lock.RUnlock()

	if idx < 0 || idx >= len(m.order) {
		return false
	}

	fn(m.order[idx], m.data[m.order[idx]])

	return true
}

func (m *Map[Key, Val]) Has(key Key) bool {
	return m.HasFunc(func(k Key, v Val) bool {
		return k == key
	})
}

func (m *Map[Key, Val]) HasFunc(fn func(key Key, val Val) bool) bool {
	m.lock.RLock()
	defer m.lock.RUnlock()

	for key, val := range m.data {
		if fn(key, val) {
			return true
		}
	}

	return false
}

func (m *Map[Key, Val]) Get(key Key, fn func(Val)) bool {
	m.lock.RLock()
	defer m.lock.RUnlock()

	val, ok := m.data[key]
	if !ok {
		return false
	}

	fn(val)

	return true
}

func (m *Map[Key, Val]) GetErr(key Key, fn func(Val) error) (bool, error) {
	var err error

	ok := m.Get(key, func(val Val) {
		err = fn(val)
	})

	return ok, err
}

func (m *Map[Key, Val]) GetDelete(key Key, fn func(Val)) bool {
	m.lock.Lock()
	defer m.lock.Unlock()

	val, ok := m.data[key]
	if !ok {
		return false
	}

	fn(val)

	delete(m.data, key)

	if idx := xslices.Index(m.order, key); idx >= 0 {
		m.order = append(m.order[:idx], m.order[idx+1:]...)
	} else {
		panic("order and data out of sync")
	}

	return true
}

func (m *Map[Key, Val]) GetDeleteErr(key Key, fn func(Val) error) (bool, error) {
	var err error

	ok := m.GetDelete(key, func(val Val) {
		err = fn(val)
	})

	return ok, err
}

func (m *Map[Key, Val]) Set(key Key, val Val) bool {
	m.lock.Lock()
	defer m.lock.Unlock()

	var had bool

	if _, ok := m.data[key]; ok {
		had = true
	}

	m.data[key] = val

	if idx := xslices.Index(m.order, key); idx >= 0 {
		m.order[idx] = key
	} else {
		m.order = append(m.order, key)
	}

	if m.sort != nil {
		slices.SortFunc(m.order, func(a, b Key) bool {
			return m.sort(a, b, m.data)
		})
	}

	return had
}

func (m *Map[Key, Val]) SetFrom(key Key, other Key) bool {
	m.lock.Lock()
	defer m.lock.Unlock()

	var had bool

	if _, ok := m.data[key]; ok {
		had = true
	}

	m.data[key] = m.data[other]

	if idx := xslices.Index(m.order, key); idx >= 0 {
		m.order[idx] = key
	} else {
		m.order = append(m.order, key)
	}

	if m.sort != nil {
		slices.SortFunc(m.order, func(a, b Key) bool {
			return m.sort(a, b, m.data)
		})
	}

	return had
}

func (m *Map[Key, Val]) Iter(fn func(key Key, val Val)) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	for _, key := range m.order {
		fn(key, m.data[key])
	}
}

func (m *Map[Key, Val]) IterKeys(fn func(Key)) {
	m.Iter(func(key Key, _ Val) {
		fn(key)
	})
}

func (m *Map[Key, Val]) IterKeysErr(fn func(Key) error) error {
	var err error

	m.IterKeys(func(key Key) {
		if err != nil {
			return
		}

		err = fn(key)
	})

	return err
}

func (m *Map[Key, Val]) IterValues(fn func(Val)) {
	m.Iter(func(_ Key, val Val) {
		fn(val)
	})
}

func (m *Map[Key, Val]) IterValuesErr(fn func(Val) error) error {
	var err error

	m.IterValues(func(val Val) {
		if err != nil {
			return
		}

		err = fn(val)
	})

	return err
}

func (m *Map[Key, Val]) Values(fn func(vals []Val)) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	vals := make([]Val, len(m.order))

	for i, key := range m.order {
		vals[i] = m.data[key]
	}

	fn(vals)
}

func (m *Map[Key, Val]) ValuesErr(fn func(vals []Val) error) error {
	var err error

	m.Values(func(vals []Val) {
		err = fn(vals)
	})

	return err
}

func (m *Map[Key, Val]) MapErr(fn func(map[Key]Val) error) error {
	m.lock.RLock()
	defer m.lock.RUnlock()

	return fn(m.data)
}

func MapGetRet[Key comparable, Val, Ret any](m *Map[Key, Val], key Key, fn func(Val) Ret) (Ret, bool) {
	var ret Ret

	ok := m.Get(key, func(val Val) {
		ret = fn(val)
	})

	return ret, ok
}

func MapValuesRet[Key comparable, Val, Ret any](m *Map[Key, Val], fn func([]Val) Ret) Ret {
	var ret Ret

	m.Values(func(vals []Val) {
		ret = fn(vals)
	})

	return ret
}

func MapValuesRetErr[Key comparable, Val, Ret any](m *Map[Key, Val], fn func([]Val) (Ret, error)) (Ret, error) {
	var ret Ret

	err := m.ValuesErr(func(vals []Val) error {
		var err error

		ret, err = fn(vals)

		return err
	})

	return ret, err
}
