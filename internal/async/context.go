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

package async

import (
	"context"
	"sync"
)

// Abortable collects groups of functions that can be aborted by calling Abort.
type Abortable struct {
	abortFunc []context.CancelFunc
	abortLock sync.RWMutex
}

func (a *Abortable) Do(ctx context.Context, fn func(context.Context)) {
	fn(a.newCancelCtx(ctx))
}

func (a *Abortable) Abort() {
	a.abortLock.RLock()
	defer a.abortLock.RUnlock()

	for _, fn := range a.abortFunc {
		fn()
	}
}

func (a *Abortable) newCancelCtx(ctx context.Context) context.Context {
	a.abortLock.Lock()
	defer a.abortLock.Unlock()

	ctx, cancel := context.WithCancel(ctx)

	a.abortFunc = append(a.abortFunc, cancel)

	return ctx
}

// RangeContext iterates over the given channel until the context is canceled or the
// channel is closed.
func RangeContext[T any](ctx context.Context, ch <-chan T, fn func(T)) {
	for {
		select {
		case v, ok := <-ch:
			if !ok {
				return
			}

			fn(v)

		case <-ctx.Done():
			return
		}
	}
}

// ForwardContext forwards all values from the src channel to the dst channel until the
// context is canceled or the src channel is closed.
func ForwardContext[T any](ctx context.Context, dst chan<- T, src <-chan T) {
	RangeContext(ctx, src, func(v T) {
		select {
		case dst <- v:
		case <-ctx.Done():
		}
	})
}
