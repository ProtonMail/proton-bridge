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

package users

import (
	"testing"

	r "github.com/stretchr/testify/require"
)

func _TestNeverLongStorePath(t *testing.T) { // nolint[unused]
	r.Fail(t, "not implemented")
}

func TestClearStoreWithStore(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	user := testNewUser(m)
	defer cleanUpUserData(user)

	r.Nil(t, user.store.Close())
	user.store = nil
	r.Nil(t, user.clearStore())
}

func TestClearStoreWithoutStore(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	user := testNewUser(m)
	defer cleanUpUserData(user)

	r.NotNil(t, user.store)
	r.Nil(t, user.clearStore())
}
