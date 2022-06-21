// Copyright (c) 2022 Proton AG
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
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.

package pchan

import (
	"sort"
	"sync"
)

type PChan struct {
	lock        sync.Mutex
	items       []*Item
	ready, done chan struct{}
	once        sync.Once
}

type Item struct {
	ch   *PChan
	val  interface{}
	prio int
	done sync.WaitGroup
}

func (item *Item) Wait() {
	item.done.Wait()
}

func (item *Item) GetPriority() int {
	item.ch.lock.Lock()
	defer item.ch.lock.Unlock()

	return item.prio
}

func (item *Item) SetPriority(priority int) {
	item.ch.lock.Lock()
	defer item.ch.lock.Unlock()

	item.prio = priority

	sort.Slice(item.ch.items, func(i, j int) bool {
		return item.ch.items[i].prio < item.ch.items[j].prio
	})
}

func New() *PChan {
	return &PChan{
		ready: make(chan struct{}),
		done:  make(chan struct{}),
	}
}

func (ch *PChan) Push(val interface{}, prio int) *Item {
	defer ch.notify()

	return ch.push(val, prio)
}

func (ch *PChan) Pop() (interface{}, int, bool) {
	select {
	case <-ch.ready:
		val, prio := ch.pop()
		return val, prio, true

	case <-ch.done:
		return nil, 0, false
	}
}

func (ch *PChan) Close() {
	ch.once.Do(func() { close(ch.done) })
}

func (ch *PChan) push(val interface{}, prio int) *Item {
	ch.lock.Lock()
	defer ch.lock.Unlock()

	item := &Item{
		ch:   ch,
		val:  val,
		prio: prio,
	}

	item.done.Add(1)

	ch.items = append(ch.items, item)

	return item
}

func (ch *PChan) pop() (interface{}, int) {
	ch.lock.Lock()
	defer ch.lock.Unlock()

	sort.Slice(ch.items, func(i, j int) bool {
		return ch.items[i].prio < ch.items[j].prio
	})

	var item *Item

	item, ch.items = ch.items[len(ch.items)-1], ch.items[:len(ch.items)-1]

	defer item.done.Done()

	return item.val, item.prio
}

func (ch *PChan) notify() {
	go func() { ch.ready <- struct{}{} }()
}
