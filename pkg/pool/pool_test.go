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

package pool_test

import (
	"testing"

	"github.com/ProtonMail/proton-bridge/v2/pkg/pool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPool(t *testing.T) {
	pool := pool.New(2, func(req interface{}, prio int) (interface{}, error) { return req, nil })

	job1, done1 := pool.NewJob("echo", 1)
	defer done1()

	job2, done2 := pool.NewJob("this", 1)
	defer done2()

	res2, err := job2.GetResult()
	require.NoError(t, err)

	res1, err := job1.GetResult()
	require.NoError(t, err)

	assert.Equal(t, "echo", res1)
	assert.Equal(t, "this", res2)
}
