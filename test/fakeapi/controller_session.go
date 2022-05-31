// Copyright (c) 2022 Proton AG
//
// This file is part of Proton Mail Bridge.Bridge.
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

package fakeapi

import (
	"bytes"
	"errors"

	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
)

type fakeSession struct {
	username      string
	uid, acc, ref string
	hasFullScope  bool
}

var errWrongNameOrPassword = errors.New("Incorrect login credentials. Please try again") //nolint:stylecheck

func (ctl *Controller) checkAccessToken(uid, acc string) bool {
	session, ok := ctl.sessionsByUID[uid]
	if !ok {
		return false
	}

	return session.uid == uid && session.acc == acc
}

func (ctl *Controller) checkScope(uid string) bool {
	session, ok := ctl.sessionsByUID[uid]
	if !ok {
		return false
	}

	return session.hasFullScope
}

func (ctl *Controller) createSessionIfAuthorized(username string, password []byte) (*fakeSession, error) {
	user, ok := ctl.usersByUsername[username]
	if !ok || !bytes.Equal(user.password, password) {
		return nil, errWrongNameOrPassword
	}

	return ctl.createSession(username, !user.has2FA), nil
}

func (ctl *Controller) createSession(username string, hasFullScope bool) *fakeSession {
	session := &fakeSession{
		username:     username,
		uid:          ctl.tokenGenerator.next("uid"),
		acc:          ctl.tokenGenerator.next("acc"),
		ref:          ctl.tokenGenerator.next("ref"),
		hasFullScope: hasFullScope,
	}

	ctl.sessionsByUID[session.uid] = session
	return session
}

func (ctl *Controller) refreshSessionIfAuthorized(uid, ref string) (*fakeSession, error) {
	session, ok := ctl.sessionsByUID[uid]
	if !ok || session.uid != uid {
		return nil, pmapi.ErrAuthFailed{OriginalError: errors.New("bad uid")}
	}

	if ref != session.ref {
		return nil, pmapi.ErrAuthFailed{OriginalError: errors.New("bad refresh token")}
	}

	session.ref = ctl.tokenGenerator.next("ref")
	session.acc = ctl.tokenGenerator.next("acc")

	ctl.sessionsByUID[session.uid] = session

	return session, nil
}

func (ctl *Controller) deleteSession(uid string) {
	delete(ctl.sessionsByUID, uid)
}
