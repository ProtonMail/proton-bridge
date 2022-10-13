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
// along with Proton Mail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package bridge

import (
	"sync"

	"github.com/ProtonMail/proton-bridge/v2/internal/user"
	"github.com/emersion/go-smtp"
)

type smtpBackend struct {
	users     map[string]*user.User
	usersLock sync.RWMutex
}

func newSMTPBackend() (*smtpBackend, error) {
	return &smtpBackend{
		users: make(map[string]*user.User),
	}, nil
}

func (backend *smtpBackend) Login(state *smtp.ConnectionState, email, password string) (smtp.Session, error) {
	backend.usersLock.RLock()
	defer backend.usersLock.RUnlock()

	for _, user := range backend.users {
		session, err := user.NewSMTPSession(email, []byte(password))
		if err != nil {
			continue
		}

		return session, nil
	}

	return nil, ErrNoSuchUser
}

func (backend *smtpBackend) AnonymousLogin(state *smtp.ConnectionState) (smtp.Session, error) {
	return nil, ErrNotImplemented
}

// addUser adds the given user to the backend.
// It returns an error if a user with the same ID already exists.
func (backend *smtpBackend) addUser(newUser *user.User) error {
	backend.usersLock.Lock()
	defer backend.usersLock.Unlock()

	if _, ok := backend.users[newUser.ID()]; ok {
		return ErrUserAlreadyExists
	}

	backend.users[newUser.ID()] = newUser

	return nil
}

// removeUser removes the given user from the backend.
// It returns an error if the user doesn't exist.
func (backend *smtpBackend) removeUser(user *user.User) error {
	backend.usersLock.Lock()
	defer backend.usersLock.Unlock()

	if _, ok := backend.users[user.ID()]; !ok {
		return ErrNoSuchUser
	}

	delete(backend.users, user.ID())

	return nil
}
