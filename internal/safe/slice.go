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
