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

import "sync"

type Value[T any] struct {
	data T
	lock sync.RWMutex
}

func NewValue[T any](data T) *Value[T] {
	return &Value[T]{
		data: data,
	}
}

func (s *Value[T]) Load(fn func(data T)) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	fn(s.data)
}

func (s *Value[T]) LoadErr(fn func(data T) error) error {
	var err error

	s.Load(func(data T) {
		err = fn(data)
	})

	return err
}

func (s *Value[T]) Save(data T) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.data = data
}

func (s *Value[T]) Mod(fn func(data *T)) {
	s.lock.Lock()
	defer s.lock.Unlock()

	fn(&s.data)
}

func LoadRet[T, Ret any](s *Value[T], fn func(data T) Ret) Ret {
	var ret Ret

	s.Load(func(data T) {
		ret = fn(data)
	})

	return ret
}

func LoadRetErr[T, Ret any](s *Value[T], fn func(data T) (Ret, error)) (Ret, error) {
	var ret Ret

	err := s.LoadErr(func(data T) error {
		var err error

		ret, err = fn(data)

		return err
	})

	return ret, err
}
