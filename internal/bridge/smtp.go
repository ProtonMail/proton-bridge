package bridge

import (
	"crypto/tls"
	"fmt"

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

	return nil
}

func (bridge *Bridge) restartSMTP() error {
	if err := bridge.closeSMTP(); err != nil {
		return err
	}

	smtpServer, err := newSMTPServer(bridge.smtpBackend, bridge.tlsConfig)
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

func newSMTPServer(smtpBackend *smtpBackend, tlsConfig *tls.Config) (*smtp.Server, error) {
	smtpServer := smtp.NewServer(smtpBackend)

	smtpServer.TLSConfig = tlsConfig
	smtpServer.Domain = constants.Host
	smtpServer.AllowInsecureAuth = true
	smtpServer.MaxLineLength = 1 << 16

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
