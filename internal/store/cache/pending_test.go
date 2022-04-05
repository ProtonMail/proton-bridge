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

package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPending(t *testing.T) {
	pending := newPending()

	pending.add("1")
	pending.add("2")
	pending.add("3")

	resCh := make(chan string)

	go func() { pending.wait("1"); resCh <- "1" }()
	go func() { pending.wait("2"); resCh <- "2" }()
	go func() { pending.wait("3"); resCh <- "3" }()

	pending.done("1")
	assert.Equal(t, "1", <-resCh)

	pending.done("2")
	assert.Equal(t, "2", <-resCh)

	pending.done("3")
	assert.Equal(t, "3", <-resCh)
}

func TestPendingUnknown(t *testing.T) {
	newPending().wait("this is not currently being waited")
}
