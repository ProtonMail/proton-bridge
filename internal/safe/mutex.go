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

type Mutex interface {
	Lock()
	Unlock()
}

func Lock(fn func(), m ...Mutex) {
	if len(m) == 0 {
		panic("no mutexes provided")
	}

	for _, m := range m {
		m.Lock()
		defer m.Unlock()
	}

	fn()
}

func LockRet[T any](fn func() T, m ...Mutex) T {
	var ret T

	Lock(func() {
		ret = fn()
	}, m...)

	return ret
}

func LockRetErr[T any](fn func() (T, error), m ...Mutex) (T, error) {
	var ret T

	err := LockRet(func() error {
		var err error

		ret, err = fn()

		return err
	}, m...)

	return ret, err
}

type RWMutex interface {
	Mutex

	RLock()
	RUnlock()
}

func RLock(fn func(), m ...RWMutex) {
	if len(m) == 0 {
		panic("no mutexes provided")
	}

	for _, m := range m {
		m.RLock()
		defer m.RUnlock()
	}

	fn()
}

func RLockRet[T any](fn func() T, m ...RWMutex) T {
	var ret T

	RLock(func() {
		ret = fn()
	}, m...)

	return ret
}

func RLockRetErr[T any](fn func() (T, error), m ...RWMutex) (T, error) {
	var err error

	ret := RLockRet(func() T {
		var ret T

		ret, err = fn()

		return ret
	}, m...)

	return ret, err
}
