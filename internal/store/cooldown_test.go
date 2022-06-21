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

package store

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCooldownExponentialWait(t *testing.T) {
	ms := time.Millisecond
	sec := time.Second

	testData := []struct {
		haveInitial, haveMax time.Duration
		haveBase             int
		wantWaitTimes        []time.Duration
	}{
		{
			haveInitial:   1 * sec,
			haveBase:      0,
			haveMax:       0 * sec,
			wantWaitTimes: []time.Duration{0 * sec},
		},
		{
			haveInitial:   0 * sec,
			haveBase:      1,
			haveMax:       0 * sec,
			wantWaitTimes: []time.Duration{0 * sec},
		},
		{
			haveInitial:   0 * sec,
			haveBase:      0,
			haveMax:       1 * sec,
			wantWaitTimes: []time.Duration{1 * sec},
		},
		{
			haveInitial:   0 * sec,
			haveBase:      1,
			haveMax:       1 * sec,
			wantWaitTimes: []time.Duration{1 * sec},
		},
		{
			haveInitial:   1 * sec,
			haveBase:      0,
			haveMax:       1 * sec,
			wantWaitTimes: []time.Duration{1 * sec},
		},
		{
			haveInitial:   1 * sec,
			haveBase:      2,
			haveMax:       1 * sec,
			wantWaitTimes: []time.Duration{1 * sec},
		},
		{
			haveInitial:   500 * ms,
			haveBase:      2,
			haveMax:       5 * sec,
			wantWaitTimes: []time.Duration{500 * ms, 1 * sec, 2 * sec, 4 * sec, 5 * sec},
		},
	}

	var testCooldown cooldown

	for _, td := range testData {
		testCooldown.setExponentialWait(td.haveInitial, td.haveBase, td.haveMax)
		assert.Equal(t, td.wantWaitTimes, testCooldown.waitTimes)
	}
}

func TestCooldownIncreaseAndReset(t *testing.T) {
	var testCooldown cooldown
	testCooldown.setWaitTimes(1*time.Second, 2*time.Second, 3*time.Second)
	assert.Equal(t, 0, testCooldown.waitIndex)

	assert.False(t, testCooldown.isTooSoon())
	assert.True(t, testCooldown.isTooSoon())
	assert.Equal(t, 0, testCooldown.waitIndex)

	testCooldown.reset()
	assert.Equal(t, 0, testCooldown.waitIndex)

	assert.False(t, testCooldown.isTooSoon())
	assert.True(t, testCooldown.isTooSoon())
	assert.Equal(t, 0, testCooldown.waitIndex)

	// increase at least N+1 times to check overflow
	testCooldown.increaseWaitTime()
	assert.True(t, testCooldown.isTooSoon())
	testCooldown.increaseWaitTime()
	assert.True(t, testCooldown.isTooSoon())
	testCooldown.increaseWaitTime()
	assert.True(t, testCooldown.isTooSoon())
	testCooldown.increaseWaitTime()
	assert.True(t, testCooldown.isTooSoon())

	assert.Equal(t, 2, testCooldown.waitIndex)
}

func TestCooldownNotSooner(t *testing.T) {
	var testCooldown cooldown
	waitTime := 100 * time.Millisecond
	testCooldown.setWaitTimes(waitTime)

	// First time it should never be too soon.
	assert.False(t, testCooldown.isTooSoon())

	// Only half of given wait time should be too soon.
	time.Sleep(waitTime / 2)
	assert.True(t, testCooldown.isTooSoon())

	// After given wait time it shouldn't be soon anymore.
	time.Sleep(waitTime/2 + time.Millisecond)
	assert.False(t, testCooldown.isTooSoon())
}
