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
	"time"
)

type Accounts struct {
	accountsLock sync.RWMutex
	accounts     map[string]*smtpAccountState
}

const maxFailedCommands = 3
const defaultErrTimeout = 20 * time.Second
const successiveErrInterval = time.Second

func NewAccounts() *Accounts {
	return &Accounts{
		accounts: make(map[string]*smtpAccountState),
	}
}

func (s *Accounts) AddAccount(account *Service) {
	s.accountsLock.Lock()
	defer s.accountsLock.Unlock()

	s.accounts[account.UserID()] = &smtpAccountState{
		service:    account,
		errTimeout: defaultErrTimeout,
	}
}

func (s *Accounts) RemoveAccount(account *Service) {
	s.accountsLock.Lock()
	defer s.accountsLock.Unlock()

	delete(s.accounts, account.UserID())
}

func (s *Accounts) CheckAuth(user string, password []byte) (string, string, error) {
	s.accountsLock.RLock()
	defer s.accountsLock.RUnlock()

	for id, account := range s.accounts {
		addrID, err := account.service.checkAuth(context.Background(), user, password)
		if err != nil {
			continue
		}

		account.service.telemetry.ReportSMTPAuthSuccess(context.Background())
		return id, addrID, nil
	}

	for _, service := range s.accounts {
		service.service.telemetry.ReportSMTPAuthFailed(user)
	}

	return "", "", ErrNoSuchUser
}

func (s *Accounts) SendMail(ctx context.Context, userID, addrID, from string, to []string, r io.Reader) error {
	if len(to) == 0 {
		return ErrInvalidRecipient
	}

	s.accountsLock.RLock()
	defer s.accountsLock.RUnlock()

	requestTime := time.Now()

	account, ok := s.accounts[userID]
	if !ok {
		return ErrNoSuchUser
	}

	if err := account.canMakeRequest(requestTime); err != nil {
		return err
	}

	err := account.service.SendMail(ctx, addrID, from, to, r)
	account.handleSMTPErr(requestTime, err)

	return err
}

type smtpAccountState struct {
	service    *Service
	errTimeout time.Duration

	errLock     sync.Mutex
	errCounter  int
	lastRequest time.Time
}

func (s *smtpAccountState) canMakeRequest(requestTime time.Time) error {
	s.errLock.Lock()
	defer s.errLock.Unlock()

	if s.errCounter >= maxFailedCommands {
		if requestTime.Sub(s.lastRequest) >= s.errTimeout {
			s.errCounter = 0
			return nil
		}

		return ErrTooManyErrors
	}

	return nil
}

func (s *smtpAccountState) handleSMTPErr(requestTime time.Time, err error) {
	s.errLock.Lock()
	defer s.errLock.Unlock()

	if err == nil || requestTime.Sub(s.lastRequest) > successiveErrInterval {
		s.errCounter = 0
	} else {
		s.errCounter++
	}

	s.lastRequest = requestTime
}
