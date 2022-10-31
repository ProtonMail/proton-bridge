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

package users

import (
	"testing"

	"github.com/ProtonMail/proton-bridge/v2/internal/events"
	gomock "github.com/golang/mock/gomock"
	r "github.com/stretchr/testify/require"
)

func TestClearData(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	users := testNewUsersWithUsers(t, m)
	defer cleanUpUsersData(users)

	m.eventListener.EXPECT().Emit(events.UserRefreshEvent, "user")
	m.eventListener.EXPECT().Emit(events.UserRefreshEvent, "users")

	m.eventListener.EXPECT().Emit(events.CloseConnectionEvent, "user@pm.me")
	m.eventListener.EXPECT().Emit(events.CloseConnectionEvent, "users@pm.me")
	m.eventListener.EXPECT().Emit(events.CloseConnectionEvent, "anotheruser@pm.me")
	m.eventListener.EXPECT().Emit(events.CloseConnectionEvent, "alsouser@pm.me")

	m.pmapiClient.EXPECT().AuthDelete(gomock.Any())
	m.credentialsStore.EXPECT().Logout("user").Return(testCredentialsDisconnected, nil)

	m.pmapiClient.EXPECT().AuthDelete(gomock.Any())
	m.credentialsStore.EXPECT().Logout("users").Return(testCredentialsSplitDisconnected, nil)

	m.locator.EXPECT().Clear()

	r.NoError(t, users.ClearData())
}
