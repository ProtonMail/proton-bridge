package bridge

import (
	"crypto/tls"
	"fmt"
	"github.com/ProtonMail/proton-bridge/v2/internal/logging"
	"net"
	"strconv"

	"github.com/ProtonMail/proton-bridge/v2/internal/constants"
	"github.com/emersion/go-sasl"
	"github.com/emersion/go-smtp"
	"github.com/sirupsen/logrus"
)

func (bridge *Bridge) serveSMTP() error {
	smtpListener, err := newListener(bridge.vault.GetSMTPPort(), bridge.vault.GetSMTPSSL(), bridge.tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to create SMTP listener: %w", err)
	}

	go func() {
		if err := bridge.smtpServer.Serve(smtpListener); err != nil {
			logrus.WithError(err).Error("SMTP server stopped")
		}
	}()

	_, port, err := net.SplitHostPort(smtpListener.Addr().String())
	if err != nil {
		return fmt.Errorf("failed to get SMTP listener address: %w", err)
	}

	portInt, err := strconv.Atoi(port)
	if err != nil {
		return fmt.Errorf("failed to convert SMTP listener port to int: %w", err)
	}

	if portInt != bridge.vault.GetSMTPPort() {
		if err := bridge.vault.SetSMTPPort(portInt); err != nil {
			return fmt.Errorf("failed to update SMTP port in vault: %w", err)
		}
	}

	return nil
}

func (bridge *Bridge) restartSMTP() error {
	if err := bridge.closeSMTP(); err != nil {
		return err
	}

	smtpServer, err := newSMTPServer(bridge.smtpBackend, bridge.tlsConfig, bridge.logSMTPCommands)
	if err != nil {
		return err
	}

	bridge.smtpServer = smtpServer

	return bridge.serveSMTP()
}

func (bridge *Bridge) closeSMTP() error {
	if err := bridge.smtpServer.Close(); err != nil {
		logrus.WithError(err).Warn("Failed to close SMTP server")
	}

	// Don't close the SMTP listener -- it's closed by the server.

	return nil
}

func newSMTPServer(smtpBackend *smtpBackend, tlsConfig *tls.Config, shouldLog bool) (*smtp.Server, error) {
	smtpServer := smtp.NewServer(smtpBackend)

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

	smtpServer.EnableAuth(sasl.Login, func(conn *smtp.Conn) sasl.Server {
		return sasl.NewLoginServer(func(address, password string) error {
			user, err := conn.Server().Backend.Login(nil, address, password)
			if err != nil {
				return err
			}

			conn.SetSession(user)

			return nil
		})
	})

	return smtpServer, nil
}
