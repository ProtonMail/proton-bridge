// Copyright (c) 2023 Proton AG
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

package smtp

import (
	"context"
	"io"
	"sync"
)

type Accounts struct {
	accountsLock sync.RWMutex
	accounts     map[string]*Service
}

func NewAccounts() *Accounts {
	return &Accounts{
		accounts: make(map[string]*Service),
	}
}

func (s *Accounts) AddAccount(account *Service) {
	s.accountsLock.Lock()
	defer s.accountsLock.Unlock()

	s.accounts[account.UserID()] = account
}

func (s *Accounts) RemoveAccount(account *Service) {
	s.accountsLock.Lock()
	defer s.accountsLock.Unlock()

	delete(s.accounts, account.UserID())
}

func (s *Accounts) CheckAuth(user string, password []byte) (string, string, error) {
	s.accountsLock.RLock()
	defer s.accountsLock.RUnlock()

	for id, service := range s.accounts {
		addrID, err := service.user.CheckAuth(user, password)
		if err != nil {
			continue
		}

		service.user.ReportSMTPAuthSuccess(context.Background())
		return id, addrID, nil
	}

	for _, service := range s.accounts {
		service.user.ReportSMTPAuthFailed(user)
	}

	return "", "", ErrNoSuchUser
}

func (s *Accounts) SendMail(ctx context.Context, userID, addrID, from string, to []string, r io.Reader) error {
	if len(to) == 0 {
		return ErrInvalidRecipient
	}

	s.accountsLock.RLock()
	defer s.accountsLock.RUnlock()

	service, ok := s.accounts[userID]
	if !ok {
		return ErrNoSuchUser
	}

	return service.SendMail(ctx, addrID, from, to, r)
}
