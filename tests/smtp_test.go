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

package tests

import (
	"fmt"
	"net/smtp"
	"strings"

	"github.com/ProtonMail/proton-bridge/v3/internal/constants"
	"github.com/cucumber/godog"
)

func (s *scenario) userConnectsSMTPClient(username, clientID string) error {
	return s.t.newSMTPClient(s.t.getUserID(username), clientID)
}

func (s *scenario) userConnectsSMTPClientOnPort(username, clientID string, port int) error {
	return s.t.newSMTPClientOnPort(s.t.getUserID(username), clientID, port)
}

func (s *scenario) userConnectsAndAuthenticatesSMTPClient(username, clientID string) error {
	return s.userConnectsAndAuthenticatesSMTPClientWithAddress(username, clientID, s.t.getUserAddrs(s.t.getUserID(username))[0])
}

func (s *scenario) userConnectsAndAuthenticatesSMTPClientWithAddress(username, clientID, address string) error {
	if err := s.t.newSMTPClient(s.t.getUserID(username), clientID); err != nil {
		return err
	}

	userID, client := s.t.getSMTPClient(clientID)

	s.t.pushError(client.Auth(smtp.PlainAuth("", address, s.t.getUserBridgePass(userID), constants.Host)))

	return nil
}

func (s *scenario) smtpClientCanAuthenticate(clientID string) error {
	userID, client := s.t.getSMTPClient(clientID)

	if err := client.Auth(smtp.PlainAuth("", s.t.getUserAddrs(userID)[0], s.t.getUserBridgePass(userID), constants.Host)); err != nil {
		return fmt.Errorf("expected no error, got %v", err)
	}

	return nil
}

func (s *scenario) smtpClientCannotAuthenticate(clientID string) error {
	userID, client := s.t.getSMTPClient(clientID)

	if err := client.Auth(smtp.PlainAuth("", s.t.getUserAddrs(userID)[0], s.t.getUserBridgePass(userID), constants.Host)); err == nil {
		return fmt.Errorf("expected error, got nil")
	}

	return nil
}

func (s *scenario) smtpClientCannotAuthenticateWithIncorrectUsername(clientID string) error {
	userID, client := s.t.getSMTPClient(clientID)

	if err := client.Auth(smtp.PlainAuth("", s.t.getUserAddrs(userID)[0]+"bad", s.t.getUserBridgePass(userID), constants.Host)); err == nil {
		return fmt.Errorf("expected error, got nil")
	}

	return nil
}

func (s *scenario) smtpClientCannotAuthenticateWithIncorrectPassword(clientID string) error {
	userID, client := s.t.getSMTPClient(clientID)

	if err := client.Auth(smtp.PlainAuth("", s.t.getUserAddrs(userID)[0], s.t.getUserBridgePass(userID)+"bad", constants.Host)); err == nil {
		return fmt.Errorf("expected error, got nil")
	}

	return nil
}

func (s *scenario) smtpClientSendsMailFrom(clientID, from string) error {
	_, client := s.t.getSMTPClient(clientID)

	s.t.pushError(client.Mail(from))

	return nil
}

func (s *scenario) smtpClientSendsRcptTo(clientID, to string) error {
	_, client := s.t.getSMTPClient(clientID)

	s.t.pushError(client.Rcpt(to))

	return nil
}

func (s *scenario) smtpClientSendsData(clientID string, data *godog.DocString) error {
	_, client := s.t.getSMTPClient(clientID)

	rc, err := client.Data()
	if err != nil {
		s.t.pushError(err)
	} else if _, err := rc.Write([]byte(data.Content)); err != nil {
		s.t.pushError(err)
	} else if err := rc.Close(); err != nil {
		s.t.pushError(err)
	}

	return nil
}

func (s *scenario) smtpClientSendsReset(clientID string) error {
	_, client := s.t.getSMTPClient(clientID)

	s.t.pushError(client.Reset())

	return nil
}

func (s *scenario) smtpClientSendsTheFollowingMessageFromTo(clientID, from, to string, message *godog.DocString) error {
	_, client := s.t.getSMTPClient(clientID)

	s.t.pushError(func() error {
		if err := client.Mail(from); err != nil {
			return err
		}

		for _, to := range strings.Split(to, ", ") {
			if err := client.Rcpt(to); err != nil {
				return err
			}
		}

		wc, err := client.Data()
		if err != nil {
			return err
		}

		if _, err := wc.Write([]byte(message.Content)); err != nil {
			return err
		}

		return wc.Close()
	}())

	return nil
}

func (s *scenario) smtpClientLogsOut(clientID string) error {
	_, client := s.t.getSMTPClient(clientID)

	s.t.pushError(client.Quit())

	return nil
}
