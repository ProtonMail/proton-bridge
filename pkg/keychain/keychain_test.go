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

package keychain

import (
	"encoding/base64"
	"testing"

	"github.com/docker/docker-credential-helpers/credentials"
	"github.com/stretchr/testify/require"
)

var suffix = []byte("\x00avoidFix\x00\x00\x00\x00\x00\x00\x00") //nolint:gochecknoglobals

var testData = map[string]string{ //nolint:gochecknoglobals
	"user1": base64.StdEncoding.EncodeToString(append([]byte("data1"), suffix...)),
	"user2": base64.StdEncoding.EncodeToString(append([]byte("data2"), suffix...)),
}

func TestInsertReadRemove(t *testing.T) {
	keychain := newKeychain(newTestHelper(), hostURL("bridge"))

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

type testHelper map[string]*credentials.Credentials

func newTestHelper() testHelper {
	return make(testHelper)
}

func (h testHelper) Add(creds *credentials.Credentials) error {
	h[creds.ServerURL] = creds
	return nil
}

func (h testHelper) Delete(url string) error {
	delete(h, url)
	return nil
}

func (h testHelper) Get(url string) (string, string, error) {
	creds := h[url]

	return creds.Username, creds.Secret, nil
}

func (h testHelper) List() (map[string]string, error) {
	list := make(map[string]string)

	for url, creds := range h {
		list[url] = creds.Username
	}

	return list, nil
}
