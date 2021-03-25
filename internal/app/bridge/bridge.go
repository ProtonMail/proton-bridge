// Copyright (c) 2021 Proton Technologies AG
//
// This file is part of ProtonMail Bridge.
//
// ProtonMail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ProtonMail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ProtonMail Bridge.  If not, see <https://www.gnu.org/licenses/>.

// Package bridge implements the bridge CLI application.
package bridge

import (
	"crypto/tls"
	"time"

	"github.com/ProtonMail/proton-bridge/internal/api"
	"github.com/ProtonMail/proton-bridge/internal/app/base"
	"github.com/ProtonMail/proton-bridge/internal/bridge"
	"github.com/ProtonMail/proton-bridge/internal/config/settings"
	pkgTLS "github.com/ProtonMail/proton-bridge/internal/config/tls"
	"github.com/ProtonMail/proton-bridge/internal/constants"
	"github.com/ProtonMail/proton-bridge/internal/frontend"
	"github.com/ProtonMail/proton-bridge/internal/frontend/types"
	"github.com/ProtonMail/proton-bridge/internal/imap"
	"github.com/ProtonMail/proton-bridge/internal/smtp"
	"github.com/ProtonMail/proton-bridge/internal/updater"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

const (
	flagLogIMAP        = "log-imap"
	flagLogSMTP        = "log-smtp"
	flagNoWindow       = "no-window"
	flagNonInteractive = "noninteractive"
)

func New(base *base.Base) *cli.App {
	app := base.NewApp(run)

	app.Flags = append(app.Flags, []cli.Flag{
		&cli.StringFlag{
			Name:  flagLogIMAP,
			Usage: "Enable logging of IMAP communications (all|client|server) (may contain decrypted data!)"},
		&cli.BoolFlag{
			Name:  flagLogSMTP,
			Usage: "Enable logging of SMTP communications (may contain decrypted data!)"},
		&cli.BoolFlag{
			Name:  flagNoWindow,
			Usage: "Don't show window after start"},
		&cli.BoolFlag{
			Name:  flagNonInteractive,
			Usage: "Start Bridge entirely noninteractively"},
	}...)

	return app
}

func run(b *base.Base, c *cli.Context) error { // nolint[funlen]
	tlsConfig, err := loadTLSConfig(b)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to load TLS config")
	}
	bridge := bridge.New(b.Locations, b.Cache, b.Settings, b.SentryReporter, b.CrashHandler, b.Listener, b.CM, b.Creds, b.Updater, b.Versioner)
	imapBackend := imap.NewIMAPBackend(b.CrashHandler, b.Listener, b.Cache, bridge)
	smtpBackend := smtp.NewSMTPBackend(b.CrashHandler, b.Listener, b.Settings, bridge)

	go func() {
		defer b.CrashHandler.HandlePanic()
		api.NewAPIServer(b.Settings, b.Listener).ListenAndServe()
	}()

	go func() {
		defer b.CrashHandler.HandlePanic()
		imapPort := b.Settings.GetInt(settings.IMAPPortKey)
		imap.NewIMAPServer(
			b.CrashHandler,
			c.String(flagLogIMAP) == "client" || c.String(flagLogIMAP) == "all",
			c.String(flagLogIMAP) == "server" || c.String(flagLogIMAP) == "all",
			imapPort, tlsConfig, imapBackend, b.UserAgent, b.Listener).ListenAndServe()
	}()

	go func() {
		defer b.CrashHandler.HandlePanic()
		smtpPort := b.Settings.GetInt(settings.SMTPPortKey)
		useSSL := b.Settings.GetBool(settings.SMTPSSLKey)
		smtp.NewSMTPServer(
			c.Bool(flagLogSMTP),
			smtpPort, useSSL, tlsConfig, smtpBackend, b.Listener).ListenAndServe()
	}()

	// Bridge supports no-window option which we should use for autostart.
	b.Autostart.Exec = append(b.Autostart.Exec, "--"+flagNoWindow)

	// We want to remove old versions if the app exits successfully.
	b.AddTeardownAction(b.Versioner.RemoveOldVersions)

	// We want cookies to be saved to disk so they are loaded the next time.
	b.AddTeardownAction(b.CookieJar.PersistCookies)

	var frontendMode string

	switch {
	case c.Bool(base.FlagCLI):
		frontendMode = "cli"
	case c.Bool(flagNonInteractive):
		return <-(make(chan error)) // Block forever.
	default:
		frontendMode = "qt"
	}

	f := frontend.New(
		constants.Version,
		constants.BuildVersion,
		b.Name,
		frontendMode,
		!c.Bool(flagNoWindow),
		b.CrashHandler,
		b.Locations,
		b.Settings,
		b.Listener,
		b.Updater,
		b.UserAgent,
		bridge,
		smtpBackend,
		b.Autostart,
		b,
	)

	// Watch for updates routine
	go func() {
		ticker := time.NewTicker(constants.UpdateCheckInterval)

		for {
			checkAndHandleUpdate(b.Updater, f, b.Settings.GetBool(settings.AutoUpdateKey))
			<-ticker.C
		}
	}()

	return f.Loop()
}

func loadTLSConfig(b *base.Base) (*tls.Config, error) {
	if !b.TLS.HasCerts() {
		if err := generateTLSCerts(b); err != nil {
			return nil, err
		}
	}

	tlsConfig, err := b.TLS.GetConfig()
	if err == nil {
		return tlsConfig, nil
	}

	logrus.WithError(err).Error("Failed to load TLS config, regenerating certificates")

	if err := generateTLSCerts(b); err != nil {
		return nil, err
	}

	return b.TLS.GetConfig()
}

func generateTLSCerts(b *base.Base) error {
	template, err := pkgTLS.NewTLSTemplate()
	if err != nil {
		return errors.Wrap(err, "failed to generate TLS template")
	}

	if err := b.TLS.GenerateCerts(template); err != nil {
		return errors.Wrap(err, "failed to generate TLS certs")
	}

	if err := b.TLS.InstallCerts(); err != nil {
		return errors.Wrap(err, "failed to install TLS certs")
	}

	return nil
}

func checkAndHandleUpdate(u types.Updater, f frontend.Frontend, autoUpdate bool) {
	log := logrus.WithField("pkg", "app/bridge")
	version, err := u.Check()
	if err != nil {
		log.WithError(err).Error("An error occurred while checking for updates")
		return
	}

	f.WaitUntilFrontendIsReady()

	// Update links in UI
	f.SetVersion(version)

	if !u.IsUpdateApplicable(version) {
		log.Info("No need to update")
		return
	}

	log.WithField("version", version.Version).Info("An update is available")

	if !autoUpdate {
		f.NotifyManualUpdate(version, u.CanInstall(version))
		return
	}

	if !u.CanInstall(version) {
		log.Info("A manual update is required")
		f.NotifySilentUpdateError(updater.ErrManualUpdateRequired)
		return
	}

	if err := u.InstallUpdate(version); err != nil {
		if errors.Cause(err) == updater.ErrDownloadVerify {
			log.WithError(err).Warning("Skipping update installation due to temporary error")
		} else {
			log.WithError(err).Error("The update couldn't be installed")
			f.NotifySilentUpdateError(err)
		}

		return
	}

	f.NotifySilentUpdateInstalled()
}
