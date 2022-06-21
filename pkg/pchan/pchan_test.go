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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPChanConcurrentPush(t *testing.T) {
	ch := New()

	var wg sync.WaitGroup

	// We are going to test with 5 additional goroutines.
	wg.Add(5)

	// Start 5 concurrent pushes.
	go func() { defer wg.Done(); ch.Push(1, 1) }()
	go func() { defer wg.Done(); ch.Push(2, 2) }()
	go func() { defer wg.Done(); ch.Push(3, 3) }()
	go func() { defer wg.Done(); ch.Push(4, 4) }()
	go func() { defer wg.Done(); ch.Push(5, 5) }()

	// Wait for the items to be pushed.
	wg.Wait()

	// All 5 should now be ready for popping.
	require.Len(t, ch.items, 5)

	// They should be popped in priority order.
	assert.Equal(t, 5, getValue(t, ch))
	assert.Equal(t, 4, getValue(t, ch))
	assert.Equal(t, 3, getValue(t, ch))
	assert.Equal(t, 2, getValue(t, ch))
	assert.Equal(t, 1, getValue(t, ch))
}

func TestPChanConcurrentPop(t *testing.T) {
	ch := New()

	var wg sync.WaitGroup

	// We are going to test with 5 additional goroutines.
	wg.Add(5)

	// Make a list to store the results in.
	var res list

	// Start 5 concurrent pops; these consume any items pushed.
	go func() { defer wg.Done(); res.append(getValue(t, ch)) }()
	go func() { defer wg.Done(); res.append(getValue(t, ch)) }()
	go func() { defer wg.Done(); res.append(getValue(t, ch)) }()
	go func() { defer wg.Done(); res.append(getValue(t, ch)) }()
	go func() { defer wg.Done(); res.append(getValue(t, ch)) }()

	// Push and block; items should be popped immediately by the waiting goroutines.
	ch.Push(1, 1).Wait()
	ch.Push(2, 2).Wait()
	ch.Push(3, 3).Wait()
	ch.Push(4, 4).Wait()
	ch.Push(5, 5).Wait()

	// Wait for all items to be popped then close the result channel.
	wg.Wait()

	assert.True(t, sort.IntsAreSorted(res.items))
}

func TestPChanClose(t *testing.T) {
	ch := New()

	go ch.Push(1, 1)

	valOpen, _, okOpen := ch.Pop()
	assert.True(t, okOpen)
	assert.Equal(t, 1, valOpen)

	ch.Close()

	valClose, _, okClose := ch.Pop()
	assert.False(t, okClose)
	assert.Nil(t, valClose)
}

type list struct {
	items []int
	mut   sync.Mutex
}

func (l *list) append(val int) {
	l.mut.Lock()
	defer l.mut.Unlock()

	l.items = append(l.items, val)
}

func getValue(t *testing.T, ch *PChan) int {
	val, _, ok := ch.Pop()

	assert.True(t, ok)

	return val.(int) //nolint:forcetypeassert
}
