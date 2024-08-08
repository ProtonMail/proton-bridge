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

package keychain

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/require"
)

var suffix = []byte("\x00avoidFix\x00\x00\x00\x00\x00\x00\x00") //nolint:gochecknoglobals

var testData = map[string]string{ //nolint:gochecknoglobals
	"user1": base64.StdEncoding.EncodeToString(append([]byte("data1"), suffix...)),
	"user2": base64.StdEncoding.EncodeToString(append([]byte("data2"), suffix...)),
}

func TestInsertReadRemove(t *testing.T) {
	keychain := newKeychain(NewTestHelper(), hostURL("bridge"))

	for id, secret := range testData {
		expectedList, _ := keychain.List()
		// Add expected secrets.
		expectedSecret := secret
		require.NoError(t, keychain.Put(id, expectedSecret))

		// Check list.
		actualList, err := keychain.List()
		require.NoError(t, err)
		expectedList = append(expectedList, id)
		require.ElementsMatch(t, expectedList, actualList)

		// Get and check what was inserted.
		_, actualSecret, err := keychain.Get(id)
		require.NoError(t, err)
		require.Equal(t, expectedSecret, actualSecret)

		// Put what changed.

		expectedSecret = "edited_" + id
		expectedSecret = base64.StdEncoding.EncodeToString(append([]byte(expectedSecret), suffix...))

		nJobs := 100
		nWorkers := 3
		jobs := make(chan interface{}, nJobs)
		done := make(chan interface{})
		for i := 0; i < nWorkers; i++ {
			go func() {
				for {
					_, more := <-jobs
					if more {
						require.NoError(t, keychain.Put(id, expectedSecret))
					} else {
						done <- nil
						return
					}
				}
			}()
		}

		for i := 0; i < nJobs; i++ {
			jobs <- nil
		}
		close(jobs)
		for i := 0; i < nWorkers; i++ {
			<-done
		}

		// Check list.
		actualList, err = keychain.List()
		require.NoError(t, err)
		require.ElementsMatch(t, expectedList, actualList)

		// Get and check what changed.
		_, actualSecret, err = keychain.Get(id)
		require.NoError(t, err)
		require.Equal(t, expectedSecret, actualSecret)

		if id != "user1" {
			// Remove.
			err = keychain.Delete(id)
			require.NoError(t, err)

			// Check removed.
			actualList, err = keychain.List()
			require.NoError(t, err)
			expectedList = expectedList[:len(expectedList)-1]
			require.ElementsMatch(t, expectedList, actualList)
		}
	}

	// Clear first.
	require.NoError(t, keychain.Delete("user1"))

	actualList, err := keychain.List()
	require.NoError(t, err)
	for id := range testData {
		require.NotContains(t, actualList, id)
	}
}

func TestIsErrKeychainNoItem(t *testing.T) {
	r := require.New(t)
	helpers := NewList(false).GetHelpers()

	for helperName := range helpers {
		kc, err := NewKeychain(helperName, "bridge-test", helpers, helperName)
		r.NoError(err)

		_, _, err = kc.Get("non-existing")
		r.True(IsErrKeychainNoItem(err), "failed for %s with error %w", helperName, err)
	}
}
