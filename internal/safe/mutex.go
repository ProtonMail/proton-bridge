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

package safe

import (
	"sync"
	"sync/atomic"

	"golang.org/x/exp/slices"
)

var nextMutexID uint64 // nolint:gochecknoglobals

// Mutex is a mutex that can be locked and unlocked.
type Mutex interface {
	Lock()
	Unlock()

	getMutexID() uint64
}

// NewMutex returns a new mutex.
func NewMutex() Mutex {
	return &mutex{
		mutexID: atomic.AddUint64(&nextMutexID, 1),
	}
}

type mutex struct {
	sync.Mutex

	mutexID uint64
}

func (m *mutex) getMutexID() uint64 {
	return m.mutexID
}

// RWMutex is a mutex that can be locked and unlocked for reading and writing.
type RWMutex interface {
	Mutex

	RLock()
	RUnlock()
}

// NewRWMutex returns a new read-write mutex.
func NewRWMutex() RWMutex {
	return &rwMutex{
		mutexID: atomic.AddUint64(&nextMutexID, 1),
	}
}

type rwMutex struct {
	sync.RWMutex

	mutexID uint64
}

func (m *rwMutex) getMutexID() uint64 {
	return m.mutexID
}

// Lock locks one or more mutexes for writing and calls the given function.
// The mutexes are locked in a deterministic order to avoid deadlocks.
func Lock(fn func(), m ...Mutex) {
	if len(m) == 0 {
		panic("no mutexes provided")
	}

	slices.SortFunc(m, func(a, b Mutex) bool {
		return a.getMutexID() < b.getMutexID()
	})

	for _, m := range m {
		m.Lock()
		defer m.Unlock()
	}

	fn()
}

// LockRet locks one or more mutexes for writing and calls the given function, returning a value.
func LockRet[T any](fn func() T, m ...Mutex) T {
	var ret T

	Lock(func() {
		ret = fn()
	}, m...)

	return ret
}

// LockRetErr locks one or more mutexes for writing and calls the given function, returning a value and an error.
func LockRetErr[T any](fn func() (T, error), m ...Mutex) (T, error) {
	var ret T

	err := LockRet(func() error {
		var err error

		ret, err = fn()

		return err
	}, m...)

	return ret, err
}

// RLock locks one or more mutexes for reading and calls the given function.
// The mutexes are locked in a deterministic order to avoid deadlocks.
func RLock(fn func(), m ...RWMutex) {
	if len(m) == 0 {
		panic("no mutexes provided")
	}

	slices.SortFunc(m, func(a, b RWMutex) bool {
		return a.getMutexID() < b.getMutexID()
	})

	for _, m := range m {
		m.RLock()
		defer m.RUnlock()
	}

	fn()
}

// RLockRet locks one or more mutexes for reading and calls the given function, returning a value.
func RLockRet[T any](fn func() T, m ...RWMutex) T {
	var ret T

	RLock(func() {
		ret = fn()
	}, m...)

	return ret
}

// RLockRetErr locks one or more mutexes for reading and calls the given function, returning a value and an error.
func RLockRetErr[T any](fn func() (T, error), m ...RWMutex) (T, error) {
	var err error

	ret := RLockRet(func() T {
		var ret T

		ret, err = fn()

		return ret
	}, m...)

	return ret, err
}
