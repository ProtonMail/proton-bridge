// Copyright (c) 2024 Proton AG
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

package crash

import (
	"fmt"
	"testing"

	"github.com/ProtonMail/gluon/async"
	"github.com/stretchr/testify/assert"
)

func TestHandler(t *testing.T) {
	assert.NotPanics(t, func() {
		var s string

		h := NewHandler(
			func(r interface{}) error {
				s += fmt.Sprintf("1: %v\n", r)
				return nil
			},
			func(r interface{}) error {
				s += fmt.Sprintf("2: %v\n", r)
				return nil
			},
		)

		h.
			AddRecoveryAction(func(r interface{}) error {
				s += fmt.Sprintf("3: %v\n", r)
				return nil
			}).
			AddRecoveryAction(func(r interface{}) error {
				s += fmt.Sprintf("4: %v\n", r)
				return nil
			})

		defer func() {
			assert.Equal(t, "1: thing\n2: thing\n3: thing\n4: thing\n", s)
		}()

		defer async.HandlePanic(h)

		panic("thing")
	})
}
