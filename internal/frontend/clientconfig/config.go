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

// Package clientconfig provides automatic config of IMAP and SMTP.
// For now only for Apple Mail.
package clientconfig

import (
	"errors"

	"github.com/ProtonMail/proton-bridge/v2/internal/config/settings"
	"github.com/ProtonMail/proton-bridge/v2/internal/config/useragent"
	"github.com/ProtonMail/proton-bridge/v2/internal/frontend/types"
	"github.com/sirupsen/logrus"
)

type AutoConfig interface {
	Name() string
	Configure(imapPort int, smtpPort int, imapSSl, smtpSSL bool, user types.User, address string) error
}

var (
	available       = map[string]AutoConfig{} //nolint:gochecknoglobals
	ErrNotAvailable = errors.New("configuration not available")
)

const AppleMailClient = "Apple Mail"

func ConfigureAppleMail(user types.User, address string, s *settings.Settings) (needRestart bool, err error) {
	return configure(AppleMailClient, user, address, s)
}

func configure(configName string, user types.User, address string, s *settings.Settings) (needRestart bool, err error) {
	log := logrus.WithField("pkg", "client_config").WithField("client", configName)

	config, ok := available[configName]
	if !ok {
		return false, ErrNotAvailable
	}

	imapPort := s.GetInt(settings.IMAPPortKey)
	imapSSL := false
	smtpPort := s.GetInt(settings.SMTPPortKey)
	smtpSSL := s.GetBool(settings.SMTPSSLKey)

	if address == "" {
		address = user.GetPrimaryAddress()
	}

	if configName == AppleMailClient {
		// If configuring apple mail for Catalina or newer, users should use SSL.
		needRestart = false
		if !smtpSSL && useragent.IsCatalinaOrNewer() {
			smtpSSL = true
			s.SetBool(settings.SMTPSSLKey, true)
			log.Warn("Detected Catalina or newer with bad SMTP SSL settings, now using SSL, bridge needs to restart")
			needRestart = true
		}
	}

	return needRestart, config.Configure(imapPort, smtpPort, imapSSL, smtpSSL, user, address)
}
