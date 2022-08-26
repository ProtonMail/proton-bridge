package bridge

import (
	"context"
	"crypto/tls"
	"fmt"

	"github.com/Masterminds/semver/v3"
	"github.com/ProtonMail/gluon"
	imapEvents "github.com/ProtonMail/gluon/events"
	"github.com/ProtonMail/proton-bridge/v2/internal/constants"
	"github.com/sirupsen/logrus"
)

const (
	defaultClientName    = "UnknownClient"
	defaultClientVersion = "0.0.1"
)

func (bridge *Bridge) GetIMAPPort() int {
	return bridge.vault.GetIMAPPort()
}

func (bridge *Bridge) SetIMAPPort(newPort int) error {
	if newPort == bridge.vault.GetIMAPPort() {
		return nil
	}

	if err := bridge.vault.SetIMAPPort(newPort); err != nil {
		return err
	}

	return bridge.restartIMAP(context.Background())
}

func (bridge *Bridge) GetIMAPSSL() bool {
	return bridge.vault.GetIMAPSSL()
}

func (bridge *Bridge) SetIMAPSSL(newSSL bool) error {
	if newSSL == bridge.vault.GetIMAPSSL() {
		return nil
	}

	if err := bridge.vault.SetIMAPSSL(newSSL); err != nil {
		return err
	}

	return bridge.restartIMAP(context.Background())
}

func (bridge *Bridge) serveIMAP() error {
	imapListener, err := newListener(bridge.vault.GetIMAPPort(), bridge.vault.GetIMAPSSL(), bridge.tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to create IMAP listener: %w", err)
	}

	bridge.imapListener = imapListener

	return bridge.imapServer.Serve(context.Background(), bridge.imapListener)
}

func (bridge *Bridge) restartIMAP(ctx context.Context) error {
	if err := bridge.imapListener.Close(); err != nil {
		logrus.WithError(err).Warn("Failed to close IMAP listener")
	}

	return bridge.serveIMAP()
}

func (bridge *Bridge) closeIMAP(ctx context.Context) error {
	if err := bridge.imapServer.Close(ctx); err != nil {
		logrus.WithError(err).Warn("Failed to close IMAP server")
	}

	if err := bridge.imapListener.Close(); err != nil {
		logrus.WithError(err).Warn("Failed to close IMAP listener")
	}

	return nil
}

func (bridge *Bridge) handleIMAPEvent(event imapEvents.Event) {
	switch event := event.(type) {
	case imapEvents.SessionAdded:
		if !bridge.identifier.HasClient() {
			bridge.identifier.SetClient(defaultClientName, defaultClientVersion)
		}

	case imapEvents.IMAPID:
		bridge.identifier.SetClient(event.IMAPID.Name, event.IMAPID.Version)
	}
}

func newIMAPServer(gluonDir string, version *semver.Version, tlsConfig *tls.Config) (*gluon.Server, error) {
	imapServer, err := gluon.New(
		gluon.WithTLS(tlsConfig),
		gluon.WithDataDir(gluonDir),
		gluon.WithVersionInfo(
			int(version.Major()),
			int(version.Minor()),
			int(version.Patch()),
			constants.FullAppName,
			"TODO",
			"TODO",
		),
		gluon.WithLogger(
			logrus.StandardLogger().WriterLevel(logrus.InfoLevel),
			logrus.StandardLogger().WriterLevel(logrus.InfoLevel),
		),
	)
	if err != nil {
		return nil, err
	}

	return imapServer, nil
}
