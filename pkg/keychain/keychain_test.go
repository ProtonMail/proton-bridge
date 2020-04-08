// Copyright (c) 2020 Proton Technologies AG
//
// This file is part of ProtonMail Bridge.
//
// ProtonMail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ProtonMail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ProtonMail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package keychain

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/require"
)

var suffix = []byte("\x00avoidFix\x00\x00\x00\x00\x00\x00\x00") //nolint[gochecknoglobals]

var testData = map[string]string{ //nolint[gochecknoglobals]
	"user1": base64.StdEncoding.EncodeToString(append([]byte("data1"), suffix...)),
	"user2": base64.StdEncoding.EncodeToString(append([]byte("data2"), suffix...)),
}

func TestSplitServiceAndID(t *testing.T) {
	acc, err := NewAccess("bridge")
	require.NoError(t, err)
	expectedUserID := "user"

	acc.KeychainURL = "Something/With/Several/Slashes/"
	acc.KeychainMacURL = acc.KeychainURL
	expectedServiceName := acc.KeychainURL
	serviceName, userID, err := splitServiceAndID(acc.KeychainName(expectedUserID))
	require.NoError(t, err)
	require.Equal(t, expectedUserID, userID)
	require.Equal(t, expectedServiceName, serviceName+"/")

	acc.KeychainURL = "SomethingWithoutSlash"
	acc.KeychainMacURL = acc.KeychainURL
	expectedServiceName = acc.KeychainURL
	serviceName, userID, err = splitServiceAndID(acc.KeychainName(expectedUserID))
	require.NoError(t, err)
	require.Equal(t, expectedUserID, userID)
	require.Equal(t, expectedServiceName, serviceName)
}

func TestInsertReadRemove(t *testing.T) { // nolint[funlen]
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	access, err := NewAccess("bridge")
	require.NoError(t, err)
	access.KeychainURL = "protonmail/testchain/users"
	access.KeychainMacURL = "ProtonMailTestChainService"

	// Clear before test.
	for id := range testData {
		// Keychain can be empty.
		_ = access.Delete(id)
	}

	for id, secret := range testData {
		expectedList, _ := access.List()
		// Add expected secrets.
		expectedSecret := secret
		require.NoError(t, access.Put(id, expectedSecret))

		// Check list.
		actualList, err := access.List()
		require.NoError(t, err)
		expectedList = append(expectedList, id)
		require.ElementsMatch(t, expectedList, actualList)

		// Get and check what was inserted.
		actualSecret, err := access.Get(id)
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
						require.NoError(t, access.Put(id, expectedSecret))
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
		actualList, err = access.List()
		require.NoError(t, err)
		require.ElementsMatch(t, expectedList, actualList)

		// Get and check what changed.
		actualSecret, err = access.Get(id)
		require.NoError(t, err)
		require.Equal(t, expectedSecret, actualSecret)

		if id != "user1" {
			// Remove.
			err = access.Delete(id)
			require.NoError(t, err)

			// Check removed.
			actualList, err = access.List()
			require.NoError(t, err)
			expectedList = expectedList[:len(expectedList)-1]
			require.ElementsMatch(t, expectedList, actualList)
		}
	}

	// Clear first.
	err = access.Delete("user1")
	require.NoError(t, err)

	actualList, err := access.List()
	require.NoError(t, err)
	for id := range testData {
		require.NotContains(t, actualList, id)
	}
}
