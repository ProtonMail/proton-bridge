// Copyright (c) 2021 Proton Technologies AG
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

package cache

import (
	"fmt"
	"testing"
	"time"

	bckMsg "github.com/ProtonMail/proton-bridge/pkg/message"
	"github.com/stretchr/testify/require"
)

var bs = &bckMsg.BodyStructure{} //nolint[gochecknoglobals]
const testUID = "testmsg"

func TestSaveAndLoad(t *testing.T) {
	msg := []byte("Test message")

	SaveMail(testUID, msg, bs)
	require.Equal(t, mailCache[testUID].data, msg)

	reader, _ := LoadMail(testUID)
	require.Equal(t, reader.Len(), len(msg))
	stored := make([]byte, len(msg))
	_, _ = reader.Read(stored)
	require.Equal(t, stored, msg)
}

func TestMissing(t *testing.T) {
	reader, _ := LoadMail("non-existing")
	require.Equal(t, reader.Len(), 0)
}

func TestClearOld(t *testing.T) {
	cacheTimeLimit = 10
	msg := []byte("Test message")
	SaveMail(testUID, msg, bs)
	time.Sleep(100 * time.Millisecond)

	reader, _ := LoadMail(testUID)
	require.Equal(t, reader.Len(), 0)
}

func TestClearBig(t *testing.T) {
	r := require.New(t)
	wantMessage := []byte("Test message")

	wantCacheSize := 3
	nTestMessages := wantCacheSize * wantCacheSize
	cacheSizeLimit = wantCacheSize*len(wantMessage) + 1
	cacheTimeLimit = int64(1 << 20) // be sure the message will survive

	// It should never have more than nSize items.
	for i := 0; i < nTestMessages; i++ {
		time.Sleep(1 * time.Millisecond)
		SaveMail(fmt.Sprintf("%s%d", testUID, i), wantMessage, bs)
		r.LessOrEqual(len(mailCache), wantCacheSize, "cache too big when %d", i)
	}

	// Check that the oldest are deleted first.
	for i := 0; i < nTestMessages; i++ {
		iUID := fmt.Sprintf("%s%d", testUID, i)
		reader, _ := LoadMail(iUID)
		mail := mailCache[iUID]

		if i < (nTestMessages - wantCacheSize) {
			r.Zero(reader.Len(), "LoadMail should return empty, but have %s for %s time %d ", string(mail.data), iUID, mail.key.Timestamp)
		} else {
			stored := make([]byte, len(wantMessage))
			_, err := reader.Read(stored)
			r.NoError(err)
			r.Equal(wantMessage, stored, "LoadMail returned wrong message: %s for %s time %d", stored, iUID, mail.key.Timestamp)
		}
	}
}

func TestConcurency(t *testing.T) {
	msg := []byte("Test message")
	for i := 0; i < 10; i++ {
		go SaveMail(fmt.Sprintf("%s%d", testUID, i), msg, bs)
	}
}
