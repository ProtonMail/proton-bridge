package safe

import "sync"

type Slice[Val any] struct {
	data []Val
	lock sync.RWMutex
}

func NewSlice[Val any](from []Val) *Slice[Val] {
	s := &Slice[Val]{
		data: make([]Val, len(from)),
	}

	copy(s.data, from)

	return s
}

func (s *Slice[Val]) Get(fn func(data []Val)) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	fn(s.data)
}

func (s *Slice[Val]) GetErr(fn func(data []Val) error) error {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return fn(s.data)
}

func (s *Slice[Val]) Set(data []Val) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.data = data
}

func GetSlice[Val, Ret any](s *Slice[Val], fn func(data []Val) Ret) Ret {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return fn(s.data)
}

func GetSliceErr[Val, Ret any](s *Slice[Val], fn func(data []Val) (Ret, error)) (Ret, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return fn(s.data)
}
