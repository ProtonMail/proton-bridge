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

package updater

import (
	"sync"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestLocker(t *testing.T) {
	l := newLocker()

	assert.NoError(t, l.doOnce(func() error {
		return nil
	}))
}

func TestLockerForwardsErrors(t *testing.T) {
	l := newLocker()

	assert.Error(t, l.doOnce(func() error {
		return errors.New("something went wrong")
	}))
}

func TestLockerAllowsOnlyOneOperation(t *testing.T) {
	l := newLocker()

	wg := &sync.WaitGroup{}

	wg.Add(1)
	go func() {
		assert.NoError(t, l.doOnce(func() error {
			time.Sleep(2 * time.Second)
			wg.Done()
			return nil
		}))
	}()

	time.Sleep(time.Second)

	err := l.doOnce(func() error { return nil })
	if assert.Error(t, err) {
		assert.Equal(t, ErrOperationOngoing, err)
	}

	wg.Wait()
}
