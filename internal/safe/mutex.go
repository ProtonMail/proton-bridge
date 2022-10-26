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
