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

func (s *Value[T]) Get(fn func(data T)) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	fn(s.data)
}

func (s *Value[T]) GetErr(fn func(data T) error) error {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return fn(s.data)
}

func (s *Value[T]) Set(data T) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.data = data
}

func GetType[T, Ret any](s *Value[T], fn func(data T) Ret) Ret {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return fn(s.data)
}

func GetTypeErr[T, Ret any](s *Value[T], fn func(data T) (Ret, error)) (Ret, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return fn(s.data)
}
