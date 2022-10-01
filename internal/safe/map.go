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

func (m *Map[Key, Val]) Get(key Key, fn func(val Val)) bool {
	m.lock.RLock()
	defer m.lock.RUnlock()

	val, ok := m.data[key]
	if !ok {
		return false
	}

	fn(val)

	return true
}

func (m *Map[Key, Val]) GetErr(key Key, fn func(val Val) error) (bool, error) {
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

func GetMap[Key comparable, Val, Ret any](m *Map[Key, Val], key Key, fn func(val Val) Ret) (Ret, bool) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	val, ok := m.data[key]
	if !ok {
		return *new(Ret), false
	}

	return fn(val), true
}

func GetMapErr[Key comparable, Val, Ret any](m *Map[Key, Val], key Key, fn func(val Val) (Ret, error)) (Ret, bool, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	val, ok := m.data[key]
	if !ok {
		return *new(Ret), false, nil
	}

	ret, err := fn(val)

	return ret, true, err
}
