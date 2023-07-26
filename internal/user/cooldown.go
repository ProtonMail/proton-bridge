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

package user

import (
	"math/rand"
	"time"
)

const SyncRetryCooldown = 20 * time.Second

type cooldownProvider interface {
	GetNextWaitTime() time.Duration
}

func jitter(max int) time.Duration {
	return time.Duration(rand.Intn(max)) * time.Second //nolint:gosec
}

type expCooldown struct {
	count int
}

func (c *expCooldown) GetNextWaitTime() time.Duration {
	waitTimes := []time.Duration{
		20 * time.Second,
		40 * time.Second,
		80 * time.Second,
		160 * time.Second,
		300 * time.Second,
		600 * time.Second,
	}

	last := len(waitTimes) - 1

	if c.count >= last {
		return waitTimes[last] + jitter(10)
	}

	c.count++

	return waitTimes[c.count-1] + jitter(10)
}

type noCooldown struct{}

func (c *noCooldown) GetNextWaitTime() time.Duration { return time.Millisecond }
