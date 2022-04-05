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

package confirmer

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConfirmerYes(t *testing.T) {
	c := New()

	req := c.NewRequest(1 * time.Second)

	go func() {
		assert.NoError(t, c.SetResult(req.ID(), true))
	}()

	res, err := req.Result()
	assert.NoError(t, err)
	assert.True(t, res)
}

func TestConfirmerNo(t *testing.T) {
	c := New()

	req := c.NewRequest(1 * time.Second)

	go func() {
		assert.NoError(t, c.SetResult(req.ID(), false))
	}()

	res, err := req.Result()
	assert.NoError(t, err)
	assert.False(t, res)
}

func TestConfirmerTimeout(t *testing.T) {
	c := New()

	req := c.NewRequest(1 * time.Second)

	go func() {
		time.Sleep(2 * time.Second)
		assert.NoError(t, c.SetResult(req.ID(), true))
	}()

	_, err := req.Result()
	assert.Error(t, err)
}

func TestConfirmerMultipleResultCalls(t *testing.T) {
	c := New()

	req := c.NewRequest(1 * time.Second)

	go func() {
		assert.NoError(t, c.SetResult(req.ID(), true))
	}()

	res, err := req.Result()
	assert.NoError(t, err)
	assert.True(t, res)

	_, errAgain := req.Result()
	assert.Error(t, errAgain)
}

func TestConfirmerMultipleSimultaneousResultCalls(t *testing.T) {
	c := New()

	req := c.NewRequest(1 * time.Second)

	go func() {
		time.Sleep(1 * time.Second)
		assert.NoError(t, c.SetResult(req.ID(), true))
	}()

	// We just check that nothing panics. We can't know which Result() will get the result though.

	go func() { _, _ = req.Result() }()
	go func() { _, _ = req.Result() }()
	go func() { _, _ = req.Result() }()

	_, _ = req.Result()
}

func TestConfirmerMultipleSetResultCalls(t *testing.T) {
	c := New()

	req := c.NewRequest(1 * time.Second)

	go func() {
		assert.NoError(t, c.SetResult(req.ID(), true))
		assert.Error(t, c.SetResult(req.ID(), true))
		assert.Error(t, c.SetResult(req.ID(), true))
		assert.Error(t, c.SetResult(req.ID(), true))
	}()

	res, err := req.Result()
	assert.NoError(t, err)
	assert.True(t, res)
}
