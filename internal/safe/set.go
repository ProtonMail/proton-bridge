package safe

import "golang.org/x/exp/maps"

type Set[Val comparable] Map[Val, struct{}]

func NewSet[Val comparable](vals ...Val) *Set[Val] {
	set := (*Set[Val])(NewMap[Val, struct{}](nil))

	for _, val := range vals {
		set.Insert(val)
	}

	return set
}

func (m *Set[Val]) Has(key Val) bool {
	m.lock.RLock()
	defer m.lock.RUnlock()

	_, ok := m.data[key]
	return ok
}

func (m *Set[Val]) Insert(key Val) {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.data[key] = struct{}{}
}

func (m *Set[Val]) Iter(fn func(key Val)) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	for key := range m.data {
		fn(key)
	}
}

func (m *Set[Val]) Values(fn func(vals []Val)) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	fn(maps.Keys(m.data))
}
