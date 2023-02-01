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
	"context"
	"crypto/tls"
	"fmt"

	"github.com/ProtonMail/proton-bridge/v3/internal/events"
	"github.com/ProtonMail/proton-bridge/v3/internal/logging"

	"github.com/ProtonMail/proton-bridge/v3/internal/constants"
	"github.com/emersion/go-sasl"
	"github.com/emersion/go-smtp"
	"github.com/sirupsen/logrus"
)

func (bridge *Bridge) serveSMTP() error {
	port, err := func() (int, error) {
		logrus.Info("Starting SMTP server")

		smtpListener, err := newListener(bridge.vault.GetSMTPPort(), bridge.vault.GetSMTPSSL(), bridge.tlsConfig)
		if err != nil {
			return 0, fmt.Errorf("failed to create SMTP listener: %w", err)
		}

		bridge.smtpListener = smtpListener

		bridge.tasks.Once(func(context.Context) {
			if err := bridge.smtpServer.Serve(smtpListener); err != nil {
				logrus.WithError(err).Info("SMTP server stopped")
			}
		})

		if err := bridge.vault.SetSMTPPort(getPort(smtpListener.Addr())); err != nil {
			return 0, fmt.Errorf("failed to store SMTP port in vault: %w", err)
		}

		return getPort(smtpListener.Addr()), nil
	}()

	if err != nil {
		bridge.publish(events.SMTPServerError{
			Error: err,
		})

		return err
	}

	bridge.publish(events.SMTPServerReady{
		Port: port,
	})

	return nil
}

func (bridge *Bridge) restartSMTP() error {
	logrus.Info("Restarting SMTP server")

	if err := bridge.closeSMTP(); err != nil {
		return fmt.Errorf("failed to close SMTP: %w", err)
	}

	bridge.publish(events.SMTPServerStopped{})

	bridge.smtpServer = newSMTPServer(bridge, bridge.tlsConfig, bridge.logSMTP)

	return bridge.serveSMTP()
}

// We close the listener ourselves even though it's also closed by smtpServer.Close().
// This is because smtpServer.Serve() is called in a separate goroutine and might be executed
// after we've already closed the server. However, go-smtp has a bug; it blocks on the listener
// even after the server has been closed. So we close the listener ourselves to unblock it.
func (bridge *Bridge) closeSMTP() error {
	logrus.Info("Closing SMTP server")

	if bridge.smtpListener != nil {
		if err := bridge.smtpListener.Close(); err != nil {
			return fmt.Errorf("failed to close SMTP listener: %w", err)
		}
	}

	if err := bridge.smtpServer.Close(); err != nil {
		logrus.WithError(err).Debug("Failed to close SMTP server (expected -- we close the listener ourselves)")
	}

	bridge.publish(events.SMTPServerStopped{})

	return nil
}

func newSMTPServer(bridge *Bridge, tlsConfig *tls.Config, logSMTP bool) *smtp.Server {
	logrus.WithField("logSMTP", logSMTP).Info("Creating SMTP server")

	smtpServer := smtp.NewServer(&smtpBackend{Bridge: bridge})

	smtpServer.TLSConfig = tlsConfig
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

	if logSMTP {
		log := logrus.WithField("protocol", "SMTP")
		log.Warning("================================================")
		log.Warning("THIS LOG WILL CONTAIN **DECRYPTED** MESSAGE DATA")
		log.Warning("================================================")

		smtpServer.Debug = logging.NewSMTPDebugLogger()
	}

	return smtpServer
}
