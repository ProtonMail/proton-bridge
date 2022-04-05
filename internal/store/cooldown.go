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

import "time"

type cooldown struct {
	waitTimes []time.Duration
	waitIndex int
	lastTry   time.Time
}

func (c *cooldown) setExponentialWait(initial time.Duration, base int, maximum time.Duration) {
	waitTimes := []time.Duration{}
	t := initial
	if base > 1 {
		for t < maximum {
			waitTimes = append(waitTimes, t)
			t *= time.Duration(base)
		}
	}
	waitTimes = append(waitTimes, maximum)
	c.setWaitTimes(waitTimes...)
}

func (c *cooldown) setWaitTimes(newTimes ...time.Duration) {
	c.waitTimes = newTimes
	c.reset()
}

// isTooSoon™ returns whether the cooldown period is not yet over.
func (c *cooldown) isTooSoon() bool {
	if time.Since(c.lastTry) < c.waitTimes[c.waitIndex] {
		return true
	}
	c.lastTry = time.Now()
	return false
}

func (c *cooldown) increaseWaitTime() {
	c.lastTry = time.Now()
	if c.waitIndex+1 < len(c.waitTimes) {
		c.waitIndex++
	}
}

func (c *cooldown) reset() {
	c.waitIndex = 0
	c.lastTry = time.Time{}
}
