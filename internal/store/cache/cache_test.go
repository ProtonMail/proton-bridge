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
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOnDiskCacheNoCompression(t *testing.T) {
	cache, err := NewOnDiskCache(t.TempDir(), &NoopCompressor{}, Options{ConcurrentRead: runtime.NumCPU(), ConcurrentWrite: runtime.NumCPU()})
	require.NoError(t, err)

	testCache(t, cache)
}

func TestOnDiskCacheGZipCompression(t *testing.T) {
	cache, err := NewOnDiskCache(t.TempDir(), &GZipCompressor{}, Options{ConcurrentRead: runtime.NumCPU(), ConcurrentWrite: runtime.NumCPU()})
	require.NoError(t, err)

	testCache(t, cache)
}

func TestInMemoryCache(t *testing.T) {
	testCache(t, NewInMemoryCache(1<<20))
}

func testCache(t *testing.T, cache Cache) {
	assert.NoError(t, cache.Unlock("userID1", []byte("my secret passphrase")))
	assert.NoError(t, cache.Unlock("userID2", []byte("my other passphrase")))

	getSetCachedMessage(t, cache, "userID1", "messageID1", "some secret")
	assert.True(t, cache.Has("userID1", "messageID1"))

	getSetCachedMessage(t, cache, "userID2", "messageID2", "some other secret")
	assert.True(t, cache.Has("userID2", "messageID2"))

	assert.NoError(t, cache.Rem("userID1", "messageID1"))
	assert.False(t, cache.Has("userID1", "messageID1"))

	assert.NoError(t, cache.Rem("userID2", "messageID2"))
	assert.False(t, cache.Has("userID2", "messageID2"))

	assert.NoError(t, cache.Delete("userID1"))
	assert.NoError(t, cache.Delete("userID2"))
}

func getSetCachedMessage(t *testing.T, cache Cache, userID, messageID, secret string) {
	assert.NoError(t, cache.Set(userID, messageID, []byte(secret)))

	data, err := cache.Get(userID, messageID)
	assert.NoError(t, err)

	assert.Equal(t, []byte(secret), data)
}
