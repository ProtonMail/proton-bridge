// Copyright (c) 2020 Proton Technologies AG
//
// This file is part of ProtonMail Bridge.Bridge.
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

package fakeapi

import (
	"errors"
)

type fakeSession struct {
	username          string
	uid, refreshToken string
	hasFullScope      bool
}

var errWrongNameOrPassword = errors.New("Incorrect login credentials. Please try again") //nolint[stylecheck]

func (cntrl *Controller) createSessionIfAuthorized(username, password string) (*fakeSession, error) {
	// get user
	user, ok := cntrl.usersByUsername[username]
	if !ok || user.password != password {
		return nil, errWrongNameOrPassword
	}

	// create session
	session := &fakeSession{
		username:     username,
		uid:          cntrl.tokenGenerator.next("uid"),
		hasFullScope: !user.has2FA,
	}
	cntrl.refreshTheTokensForSession(session)

	cntrl.sessionsByUID[session.uid] = session
	return session, nil
}

func (cntrl *Controller) refreshTheTokensForSession(session *fakeSession) {
	session.refreshToken = cntrl.tokenGenerator.next("refresh")
}

func (cntrl *Controller) deleteSession(uid string) {
	delete(cntrl.sessionsByUID, uid)
}
