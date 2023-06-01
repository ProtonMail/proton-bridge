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

package bridge

import (
	"fmt"
	"io"
	"strings"

	"github.com/ProtonMail/proton-bridge/v3/internal/safe"
	"github.com/ProtonMail/proton-bridge/v3/internal/useragent"
	"github.com/emersion/go-smtp"
	"github.com/sirupsen/logrus"
)

type smtpBackend struct {
	*Bridge
}

type smtpSession struct {
	*Bridge

	userID string
	authID string

	from string
	to   []string
}

func (be *smtpBackend) NewSession(*smtp.Conn) (smtp.Session, error) {
	return &smtpSession{Bridge: be.Bridge}, nil
}

func (s *smtpSession) AuthPlain(username, password string) error {
	return safe.RLockRet(func() error {
		for _, user := range s.users {
			addrID, err := user.CheckAuth(username, []byte(password))
			if err != nil {
				continue
			}

			s.userID = user.ID()
			s.authID = addrID

			if strings.Contains(s.Bridge.GetCurrentUserAgent(), useragent.DefaultUserAgent) {
				s.Bridge.setUserAgent(useragent.UnknownClient, useragent.DefaultVersion)
			}
			return nil
		}

		logrus.WithFields(logrus.Fields{
			"username": username,
			"pkg":      "smtp",
		}).Error("Incorrect login credentials.")

		return fmt.Errorf("invalid username or password")
	}, s.usersLock)
}

func (s *smtpSession) Reset() {
	s.from = ""
	s.to = nil
}

func (s *smtpSession) Logout() error {
	s.Reset()
	return nil
}

func (s *smtpSession) Mail(from string, _ *smtp.MailOptions) error {
	s.from = from
	return nil
}

func (s *smtpSession) Rcpt(to string) error {
	if len(to) > 0 {
		s.to = append(s.to, to)
	}

	return nil
}

func (s *smtpSession) Data(r io.Reader) error {
	err := safe.RLockRet(func() error {
		user, ok := s.users[s.userID]
		if !ok {
			return ErrNoSuchUser
		}

		return user.SendMail(s.authID, s.from, s.to, r)
	}, s.usersLock)

	if err != nil {
		logrus.WithField("pkg", "smtp").WithError(err).Error("Send mail failed.")
	}

	return err
}
