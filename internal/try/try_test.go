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
// along with Proton Mail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package try

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTry(t *testing.T) {
	res, err := CatchVal(func() (string, error) {
		return "foo", nil
	})
	require.NoError(t, err)
	require.Equal(t, "foo", res)
}

func TestTryCatch(t *testing.T) {
	tryErr := fmt.Errorf("oops")

	res, err := CatchVal(
		func() (string, error) {
			return "", tryErr
		},
		func() error {
			return nil
		},
	)
	require.ErrorIs(t, err, tryErr)
	require.Zero(t, res)
}

func TestTryCatchError(t *testing.T) {
	tryErr := fmt.Errorf("oops")

	res, err := CatchVal(
		func() (string, error) {
			return "", tryErr
		},
		func() error {
			return fmt.Errorf("catch error")
		},
	)
	require.ErrorIs(t, err, tryErr)
	require.Zero(t, res)
}

func TestTryPanic(t *testing.T) {
	res, err := CatchVal(
		func() (string, error) {
			panic("oops")
		},
		func() error {
			return nil
		},
	)
	require.ErrorContains(t, err, "panic: oops")
	require.Zero(t, res)
}

func TestTryCatchPanic(t *testing.T) {
	tryErr := fmt.Errorf("oops")

	res, err := CatchVal(
		func() (string, error) {
			return "", tryErr
		},
		func() error {
			panic("oops")
		},
	)
	require.ErrorIs(t, err, tryErr)
	require.Zero(t, res)
}
