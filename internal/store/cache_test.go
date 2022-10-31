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

package store

import (
	"testing"

	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestIsCachedCrashRecovers(t *testing.T) {
	r := require.New(t)
	m, clear := initMocks(t)
	defer clear()

	m.newStoreNoEvents(t, true, &pmapi.Message{
		ID:      "msg1",
		Subject: "subject",
	})

	r.False(m.store.IsCached("msg1"))

	m.store.cache = nil
	r.False(m.store.IsCached("msg1"))
}

var wantLiteral = []byte("Mime-Version: 1.0\r\nContent-Transfer-Encoding: quoted-printable\r\nContent-Type: \r\nReferences:  <msg1@protonmail.internalid>\r\nX-Pm-Date: Thu, 01 Jan 1970 00:00:00 +0000\r\nX-Pm-External-Id: <>\r\nX-Pm-Internal-Id: msg1\r\nX-Original-Date: Mon, 01 Jan 0001 00:00:00 +0000\r\nDate: Fri, 13 Aug 1982 00:00:00 +0000\r\nMessage-Id: <msg1@protonmail.internalid>\r\nSubject: subject\r\n\r\n")

func TestGetCachedMessageOK(t *testing.T) {
	r := require.New(t)
	m, clear := initMocks(t)
	defer clear()

	messageID := "msg1"

	m.newStoreNoEvents(t, true, &pmapi.Message{
		ID:      messageID,
		Subject: "subject",
		Flags:   pmapi.FlagReceived,
		Body:    "body",
	})

	// Have build job
	m.client.EXPECT().
		KeyRingForAddressID(gomock.Any()).
		Return(testPrivateKeyRing, nil).
		Times(1)

	haveLiteral, err := m.store.getCachedMessage(messageID)
	r.NoError(err)
	r.Equal(wantLiteral, haveLiteral)

	r.True(m.store.IsCached(messageID))

	// No build job
	haveLiteral, err = m.store.getCachedMessage(messageID)
	r.NoError(err)
	r.Equal(wantLiteral, haveLiteral)
	r.True(m.store.IsCached(messageID))
}

func TestGetCachedMessageCacheLocked(t *testing.T) {
	r := require.New(t)
	m, clear := initMocks(t)
	defer clear()

	messageID := "msg1"

	m.newStoreNoEvents(t, true, &pmapi.Message{
		ID:      messageID,
		Subject: "subject",
		Flags:   pmapi.FlagReceived,
		Body:    "body",
	})

	// Have build job
	m.client.EXPECT().
		KeyRingForAddressID(gomock.Any()).
		Return(testPrivateKeyRing, nil).
		Times(1)
	haveLiteral, err := m.store.getCachedMessage(messageID)
	r.NoError(err)
	r.Equal(wantLiteral, haveLiteral)
	r.True(m.store.IsCached(messageID))

	// Lock cache
	m.store.cache.Lock(m.store.user.ID())

	// Have build job again due to failure
	m.client.EXPECT().
		KeyRingForAddressID(gomock.Any()).
		Return(testPrivateKeyRing, nil).
		Times(1)

	haveLiteral, err = m.store.getCachedMessage(messageID)
	r.NoError(err)
	r.Equal(wantLiteral, haveLiteral)
	r.True(m.store.IsCached(messageID))

	// No build job
	haveLiteral, err = m.store.getCachedMessage(messageID)
	r.NoError(err)
	r.Equal(wantLiteral, haveLiteral)
	r.True(m.store.IsCached(messageID))
}
