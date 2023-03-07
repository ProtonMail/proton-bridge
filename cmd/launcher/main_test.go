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
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.

package main

import (
	"testing"

	"github.com/bradenaw/juniper/xslices"
	"github.com/stretchr/testify/assert"
)

func TestSliceContains(t *testing.T) {
	assert.True(t, sliceContains([]string{"a", "b", "c"}, "a"))
	assert.True(t, sliceContains([]int{1, 2, 3}, 2))
	assert.False(t, sliceContains([]string{"a", "b", "c"}, "A"))
	assert.False(t, sliceContains([]int{1, 2, 3}, 4))
	assert.False(t, sliceContains([]string{}, "a"))
	assert.True(t, sliceContains([]string{"a", "a"}, "a"))
}

func TestFindAndStrip(t *testing.T) {
	list := []string{"a", "b", "c", "c", "b", "c"}

	result, found := findAndStrip(list, "a")
	assert.True(t, found)
	assert.True(t, xslices.Equal(result, []string{"b", "c", "c", "b", "c"}))

	result, found = findAndStrip(list, "c")
	assert.True(t, found)
	assert.True(t, xslices.Equal(result, []string{"a", "b", "b"}))

	result, found = findAndStrip([]string{"c", "c", "c"}, "c")
	assert.True(t, found)
	assert.True(t, xslices.Equal(result, []string{}))

	result, found = findAndStrip(list, "A")
	assert.False(t, found)
	assert.True(t, xslices.Equal(result, list))

	result, found = findAndStrip([]string{}, "a")
	assert.False(t, found)
	assert.True(t, xslices.Equal(result, []string{}))
}

func TestFindAndStripWait(t *testing.T) {
	result, found, values := findAndStripWait([]string{"a", "b", "c"})
	assert.False(t, found)
	assert.True(t, xslices.Equal(result, []string{"a", "b", "c"}))
	assert.True(t, xslices.Equal(values, []string{}))

	result, found, values = findAndStripWait([]string{"a", "--wait", "b"})
	assert.True(t, found)
	assert.True(t, xslices.Equal(result, []string{"a"}))
	assert.True(t, xslices.Equal(values, []string{"b"}))

	result, found, values = findAndStripWait([]string{"a", "--wait", "b", "--wait", "c"})
	assert.True(t, found)
	assert.True(t, xslices.Equal(result, []string{"a"}))
	assert.True(t, xslices.Equal(values, []string{"b", "c"}))

	result, found, values = findAndStripWait([]string{"a", "--wait", "b", "--wait", "c", "--wait", "d"})
	assert.True(t, found)
	assert.True(t, xslices.Equal(result, []string{"a"}))
	assert.True(t, xslices.Equal(values, []string{"b", "c", "d"}))
}
