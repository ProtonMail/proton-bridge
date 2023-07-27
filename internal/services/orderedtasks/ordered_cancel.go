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

package orderedtasks

import (
	"context"

	"github.com/ProtonMail/gluon/async"
	gl "github.com/ProtonMail/gluon/logging"
	"github.com/bradenaw/juniper/xslices"
	"golang.org/x/exp/slices"
)

type task struct {
	done func()
	ch   chan struct{}
}

func (t *task) cancelAndWait() {
	t.cancel()
	t.wait()
}

func (t *task) cancel() {
	t.done()
}

func (t *task) wait() {
	<-t.ch
}

// OrderedCancelGroup cancels go routines in reverse order that they are launched and waits for completion before
// advancing to the next one.
type OrderedCancelGroup struct {
	cancels      []*task
	panicHandler async.PanicHandler
}

func NewOrderedCancelGroup(handler async.PanicHandler) *OrderedCancelGroup {
	return &OrderedCancelGroup{panicHandler: handler}
}

func (o *OrderedCancelGroup) Go(ctx context.Context, userID, debugName string, f func(ctx context.Context)) {
	ctx, cancel := context.WithCancel(ctx)
	task := &task{done: cancel, ch: make(chan struct{})}

	go func() {
		gl.DoAnnotated(ctx, func(ctx context.Context) {
			defer async.HandlePanic(o.panicHandler)
			defer close(task.ch)

			f(ctx)
		}, gl.Labels{"group": debugName, "user": userID})
	}()

	o.cancels = append(o.cancels, task)
}

func (o *OrderedCancelGroup) reversed() []*task {
	s := slices.Clone(o.cancels)
	xslices.Reverse(s)

	return s
}

func (o *OrderedCancelGroup) CancelAndWait() {
	for _, t := range o.reversed() {
		t.cancelAndWait()
	}

	o.cancels = nil
}

func (o *OrderedCancelGroup) Cancel() {
	for _, t := range o.reversed() {
		t.cancel()
	}
}

func (o *OrderedCancelGroup) Wait() {
	for _, t := range o.reversed() {
		t.wait()
	}

	o.cancels = nil
}
