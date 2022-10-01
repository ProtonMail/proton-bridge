package safe

import "sync"

type Type[T any] struct {
	data T
	lock sync.RWMutex
}

func NewType[T any](data T) *Type[T] {
	return &Type[T]{
		data: data,
	}
}

func (s *Type[T]) Get(fn func(data T)) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	fn(s.data)
}

func (s *Type[T]) GetErr(fn func(data T) error) error {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return fn(s.data)
}

func (s *Type[T]) Set(data T) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.data = data
}

func GetType[T, Ret any](s *Type[T], fn func(data T) Ret) Ret {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return fn(s.data)
}

func GetTypeErr[T, Ret any](s *Type[T], fn func(data T) (Ret, error)) (Ret, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return fn(s.data)
}
