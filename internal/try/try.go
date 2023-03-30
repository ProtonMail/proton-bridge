// Copyright (c) 2023 Proton AG
//
// This file is part of Proton Mail Bridge.
//
// Proton Mail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Proton Mail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Proton Mail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package try

import (
	"fmt"
	"sync"

	"github.com/ProtonMail/gluon/async"
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
			logrus.WithError(err).Error("Catch handler failed")
		}
	}
}

type Group struct {
	mu           sync.Mutex
	panicHandler async.PanicHandler
}

func MakeGroup(panicHandler async.PanicHandler) Group {
	return Group{panicHandler: panicHandler}
}

func (wg *Group) GoTry(fn func(bool)) {
	if wg.mu.TryLock() {
		go func() {
			defer async.HandlePanic(wg.panicHandler)
			defer wg.mu.Unlock()
			fn(true)
		}()
	} else {
		go func() {
			defer async.HandlePanic(wg.panicHandler)
			fn(false)
		}()
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
