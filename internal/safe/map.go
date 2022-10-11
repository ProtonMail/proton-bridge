package safe

import (
	"sync"

	"golang.org/x/exp/maps"
)

type Map[Key comparable, Val any] struct {
	data map[Key]Val
	lock sync.RWMutex
}

func NewMap[Key comparable, Val any](from map[Key]Val) *Map[Key, Val] {
	m := &Map[Key, Val]{
		data: make(map[Key]Val),
	}

	for key, val := range from {
		m.Set(key, val)
	}

	return m
}

func (m *Map[Key, Val]) Has(key Key) bool {
	m.lock.RLock()
	defer m.lock.RUnlock()

	_, ok := m.data[key]
	return ok
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
	m.lock.RLock()
	defer m.lock.RUnlock()

	val, ok := m.data[key]
	if !ok {
		return false, nil
	}

	return true, fn(val)
}

func (m *Map[Key, Val]) Set(key Key, val Val) {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.data[key] = val
}

func (m *Map[Key, Val]) Delete(key Key) {
	m.lock.Lock()
	defer m.lock.Unlock()

	delete(m.data, key)
}

func (m *Map[Key, Val]) Iter(fn func(key Key, val Val)) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	for key, val := range m.data {
		fn(key, val)
	}
}

func (m *Map[Key, Val]) Keys(fn func(keys []Key)) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	fn(maps.Keys(m.data))
}

func (m *Map[Key, Val]) Values(fn func(vals []Val)) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	fn(maps.Values(m.data))
}

func GetMap[Key comparable, Val, Ret any](m *Map[Key, Val], key Key, fn func(Val) Ret, fallback func() Ret) Ret {
	m.lock.RLock()
	defer m.lock.RUnlock()

	val, ok := m.data[key]
	if !ok {
		return fallback()
	}

	return fn(val)
}

func GetMapErr[Key comparable, Val, Ret any](m *Map[Key, Val], key Key, fn func(Val) (Ret, error), fallback func() (Ret, error)) (Ret, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	val, ok := m.data[key]
	if !ok {
		return fallback()
	}

	return fn(val)
}

func FindMap[Key comparable, Val, Ret any](m *Map[Key, Val], cmp func(Val) bool, fn func(Val) Ret, fallback func() Ret) Ret {
	m.lock.RLock()
	defer m.lock.RUnlock()

	for _, val := range m.data {
		if cmp(val) {
			return fn(val)
		}
	}

	return fallback()
}

func FindMapErr[Key comparable, Val, Ret any](m *Map[Key, Val], cmp func(Val) bool, fn func(Val) (Ret, error), fallback func() (Ret, error)) (Ret, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	for _, val := range m.data {
		if cmp(val) {
			return fn(val)
		}
	}

	return fallback()
}
