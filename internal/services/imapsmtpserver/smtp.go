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

package imapsmtpserver

import (
	"crypto/tls"

	"github.com/ProtonMail/proton-bridge/v3/internal/constants"
	"github.com/ProtonMail/proton-bridge/v3/internal/identifier"
	"github.com/ProtonMail/proton-bridge/v3/internal/logging"
	smtpservice "github.com/ProtonMail/proton-bridge/v3/internal/services/smtp"
	"github.com/emersion/go-sasl"
	"github.com/emersion/go-smtp"
	"github.com/sirupsen/logrus"
)

type SMTPSettingsProvider interface {
	TLSConfig() *tls.Config
	Log() bool
	Port() int
	SetPort(int) error
	UseSSL() bool
	Identifier() identifier.UserAgentUpdater
}

func newSMTPServer(accounts *smtpservice.Accounts, settings SMTPSettingsProvider) *smtp.Server {
	logrus.WithField("logSMTP", settings.Log()).Info("Creating SMTP server")

	smtpServer := smtp.NewServer(smtpservice.NewBackend(accounts, settings.Identifier()))

	smtpServer.TLSConfig = settings.TLSConfig()
	smtpServer.Domain = constants.Host
	smtpServer.AllowInsecureAuth = true
	smtpServer.MaxLineLength = 1 << 16
	smtpServer.ErrorLog = logging.NewSMTPLogger()

	// go-smtp suppors SASL PLAIN but not LOGIN. We need to add LOGIN support ourselves.
	smtpServer.EnableAuth(sasl.Login, func(conn *smtp.Conn) sasl.Server {
		return sasl.NewLoginServer(func(username, password string) error {
			return conn.Session().AuthPlain(username, password)
		})
	})

	if settings.Log() {
		log := logrus.WithField("protocol", "SMTP")
		log.Warning("================================================")
		log.Warning("THIS LOG WILL CONTAIN **DECRYPTED** MESSAGE DATA")
		log.Warning("================================================")

		smtpServer.Debug = logging.NewSMTPDebugLogger()
	}

	return smtpServer
}
