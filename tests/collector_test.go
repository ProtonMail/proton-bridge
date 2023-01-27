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

package tests

import (
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/ProtonMail/gluon/queue"
	"github.com/ProtonMail/proton-bridge/v3/internal/events"
)

type eventCollector struct {
	events map[reflect.Type]*queue.QueuedChannel[events.Event]
	fwdCh  []*queue.QueuedChannel[events.Event]
	lock   sync.Mutex
	wg     sync.WaitGroup
}

func newEventCollector() *eventCollector {
	return &eventCollector{
		events: make(map[reflect.Type]*queue.QueuedChannel[events.Event]),
	}
}

func (c *eventCollector) collectFrom(eventCh <-chan events.Event) <-chan events.Event {
	c.lock.Lock()
	defer c.lock.Unlock()

	fwdCh := queue.NewQueuedChannel[events.Event](0, 0)

	c.fwdCh = append(c.fwdCh, fwdCh)

	c.wg.Add(1)

	go func() {
		defer fwdCh.CloseAndDiscardQueued()
		defer c.wg.Done()

		for event := range eventCh {
			c.push(event)
		}
	}()

	return fwdCh.GetChannel()
}

func awaitType[T events.Event](c *eventCollector, ofType T, timeout time.Duration) (T, bool) {
	if event := c.await(ofType, timeout); event == nil {
		return *new(T), false //nolint:gocritic
	} else if event, ok := event.(T); !ok {
		panic(fmt.Errorf("unexpected event type %T", event))
	} else {
		return event, true
	}
}

func (c *eventCollector) await(ofType events.Event, timeout time.Duration) events.Event {
	select {
	case event := <-c.getEventCh(ofType):
		return event

	case <-time.After(timeout):
		return nil
	}
}

func (c *eventCollector) push(event events.Event) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if _, ok := c.events[reflect.TypeOf(event)]; !ok {
		c.events[reflect.TypeOf(event)] = queue.NewQueuedChannel[events.Event](0, 0)
	}

	c.events[reflect.TypeOf(event)].Enqueue(event)

	for _, eventCh := range c.fwdCh {
		eventCh.Enqueue(event)
	}
}

func (c *eventCollector) getEventCh(ofType events.Event) <-chan events.Event {
	c.lock.Lock()
	defer c.lock.Unlock()

	if _, ok := c.events[reflect.TypeOf(ofType)]; !ok {
		c.events[reflect.TypeOf(ofType)] = queue.NewQueuedChannel[events.Event](0, 0)
	}

	return c.events[reflect.TypeOf(ofType)].GetChannel()
}

func (c *eventCollector) close() {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.wg.Wait()

	for _, eventCh := range c.events {
		eventCh.CloseAndDiscardQueued()
	}
	c.events = nil
}
