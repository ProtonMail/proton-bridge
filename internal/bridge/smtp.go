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
	"context"
	"crypto/tls"
	"fmt"

	"github.com/ProtonMail/proton-bridge/v2/internal/logging"
	"github.com/ProtonMail/proton-bridge/v2/internal/safe"
	"github.com/ProtonMail/proton-bridge/v2/internal/user"

	"github.com/ProtonMail/proton-bridge/v2/internal/constants"
	"github.com/emersion/go-smtp"
	"github.com/sirupsen/logrus"
)

func (bridge *Bridge) serveSMTP() error {
	smtpListener, err := newListener(bridge.vault.GetSMTPPort(), bridge.vault.GetSMTPSSL(), bridge.tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to create SMTP listener: %w", err)
	}

	bridge.smtpListener = smtpListener

	bridge.tasks.Once(func(ctx context.Context) {
		if err := bridge.smtpServer.Serve(smtpListener); err != nil {
			logrus.WithError(err).Debug("SMTP server stopped")
		}
	})

	if err := bridge.vault.SetSMTPPort(getPort(smtpListener.Addr())); err != nil {
		return fmt.Errorf("failed to set IMAP port: %w", err)
	}

	return nil
}

func (bridge *Bridge) restartSMTP() error {
	if err := bridge.closeSMTP(); err != nil {
		return fmt.Errorf("failed to close SMTP: %w", err)
	}

	bridge.smtpServer = newSMTPServer(bridge.users, bridge.tlsConfig, bridge.logSMTP)

	return bridge.serveSMTP()
}

// We close the listener ourselves even though it's also closed by smtpServer.Close().
// This is because smtpServer.Serve() is called in a separate goroutine and might be executed
// after we've already closed the server. However, go-smtp has a bug; it blocks on the listener
// even after the server has been closed. So we close the listener ourselves to unblock it.
func (bridge *Bridge) closeSMTP() error {
	if err := bridge.smtpListener.Close(); err != nil {
		return fmt.Errorf("failed to close SMTP listener: %w", err)
	}

	if err := bridge.smtpServer.Close(); err != nil {
		logrus.WithError(err).Debug("Failed to close SMTP server")
	}

	return nil
}

func newSMTPServer(users *safe.Map[string, *user.User], tlsConfig *tls.Config, shouldLog bool) *smtp.Server {
	smtpServer := smtp.NewServer(&smtpBackend{users})

	smtpServer.TLSConfig = tlsConfig
	smtpServer.Domain = constants.Host
	smtpServer.AllowInsecureAuth = true
	smtpServer.MaxLineLength = 1 << 16
	smtpServer.ErrorLog = logging.NewSMTPLogger()

	if shouldLog {
		log := logrus.WithField("protocol", "SMTP")
		log.Warning("================================================")
		log.Warning("THIS LOG WILL CONTAIN **DECRYPTED** MESSAGE DATA")
		log.Warning("================================================")
		smtpServer.Debug = logging.NewSMTPDebugLogger()
	}

	return smtpServer
}
