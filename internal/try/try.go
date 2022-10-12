package try

import (
	"fmt"
	"sync"

	"github.com/bradenaw/juniper/xerrors"
	"github.com/sirupsen/logrus"
)

// Catch tries to execute the `try` function, and if it fails or panics,
// it executes the `handlers` functions in order.
func Catch(try func() error, handlers ...func() error) error {
	if _, err := CatchVal(func() (any, error) { return nil, try() }, handlers...); err != nil {
		return err
	}

	return nil
}

// CatchVal tries to execute the `try` function, and if it fails or panics,
// it executes the `handlers` functions in order.
func CatchVal[T any](try func() (T, error), handlers ...func() error) (res T, err error) {
	defer func() {
		if r := recover(); r != nil {
			catch(handlers...)
			err = xerrors.WithStack(fmt.Errorf("panic: %v", r))
		}
	}()

	if res, err = try(); err != nil {
		catch(handlers...)
		return res, err
	}

	return res, nil
}

func catch(handlers ...func() error) {
	defer func() {
		if r := recover(); r != nil {
			logrus.WithError(xerrors.WithStack(fmt.Errorf("panic: %v", r))).Error("Catch handler panicked")
		}
	}()

	for _, handler := range handlers {
		if err := handler(); err != nil {
			logrus.WithError(xerrors.WithStack(err)).Error("Catch handler failed")
		}
	}
}

type Group struct {
	mu sync.Mutex
}

func (wg *Group) GoTry(fn func(bool)) {
	if wg.mu.TryLock() {
		go func() {
			defer wg.mu.Unlock()
			fn(true)
		}()
	} else {
		go fn(false)
	}
}

func (wg *Group) Lock() {
	wg.mu.Lock()
}

func (wg *Group) Unlock() {
	wg.mu.Unlock()
}

func (wg *Group) Wait() {
	wg.mu.Lock()
	defer wg.mu.Unlock()
}