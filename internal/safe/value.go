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
