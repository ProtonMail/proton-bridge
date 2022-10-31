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

package imap

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestUpdatesCanDelete(t *testing.T) {
	u := newIMAPUpdates(nil)

	can, _ := u.CanDelete("mbox")
	require.True(t, can)

	u.forbidExpunge("mbox")
	u.allowExpunge("mbox")

	can, _ = u.CanDelete("mbox")
	require.True(t, can)
}

func TestUpdatesCannotDelete(t *testing.T) {
	u := newIMAPUpdates(nil)

	u.forbidExpunge("mbox")
	can, wait := u.CanDelete("mbox")
	require.False(t, can)

	ch := make(chan time.Duration)
	go func() {
		start := time.Now()
		wait()
		ch <- time.Since(start)
		close(ch)
	}()

	time.Sleep(200 * time.Millisecond)
	u.allowExpunge("mbox")
	duration := <-ch

	require.True(t, duration > 200*time.Millisecond)
}
