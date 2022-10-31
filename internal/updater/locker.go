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
	"sync/atomic"

	"github.com/pkg/errors"
)

var ErrOperationOngoing = errors.New("the operation is already ongoing")

// locker is an easy way to ensure we only perform one update at a time.
type locker struct {
	ongoing atomic.Value
}

func newLocker() *locker {
	l := &locker{}

	l.ongoing.Store(false)

	return l
}

func (l *locker) doOnce(fn func() error) error {
	if l.ongoing.Load().(bool) { //nolint:forcetypeassert
		return ErrOperationOngoing
	}

	l.ongoing.Store(true)
	defer func() { l.ongoing.Store(false) }()

	return fn()
}
