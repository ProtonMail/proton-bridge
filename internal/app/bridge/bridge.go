// Copyright (c) 2020 Proton Technologies AG
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
	"time"

	"github.com/ProtonMail/proton-bridge/internal/api"
	"github.com/ProtonMail/proton-bridge/internal/app/base"
	"github.com/ProtonMail/proton-bridge/internal/bridge"
	"github.com/ProtonMail/proton-bridge/internal/config/settings"
	"github.com/ProtonMail/proton-bridge/internal/constants"
	"github.com/ProtonMail/proton-bridge/internal/frontend"
	"github.com/ProtonMail/proton-bridge/internal/frontend/types"
	"github.com/ProtonMail/proton-bridge/internal/imap"
	"github.com/ProtonMail/proton-bridge/internal/smtp"
	"github.com/ProtonMail/proton-bridge/internal/updater"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func New(base *base.Base) *cli.App {
	app := base.NewApp(run)

	app.Flags = append(app.Flags, []cli.Flag{
		&cli.StringFlag{
			Name:  "log-imap",
			Usage: "Enable logging of IMAP communications (all|client|server) (may contain decrypted data!)"},
		&cli.BoolFlag{
			Name:  "log-smtp",
			Usage: "Enable logging of SMTP communications (may contain decrypted data!)"},
		&cli.BoolFlag{
			Name:  "no-window",
			Usage: "Don't show window after start"},
		&cli.BoolFlag{
			Name:  "noninteractive",
			Usage: "Start Bridge entirely noninteractively"},
	}...)

	return app
}

func run(b *base.Base, c *cli.Context) error { // nolint[funlen]
	tls, err := b.TLS.GetConfig()
	if err != nil {
		logrus.WithError(err).Fatal("Failed to create TLS config")
	}

	bridge := bridge.New(b.Locations, b.Cache, b.Settings, b.CrashHandler, b.Listener, b.CM, b.Creds)
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
			c.String("log-imap") == "client" || c.String("log-imap") == "all",
			c.String("log-imap") == "server" || c.String("log-imap") == "all",
			imapPort, tls, imapBackend, b.Listener).ListenAndServe()
	}()

	go func() {
		defer b.CrashHandler.HandlePanic()
		smtpPort := b.Settings.GetInt(settings.SMTPPortKey)
		useSSL := b.Settings.GetBool(settings.SMTPSSLKey)
		smtp.NewSMTPServer(
			c.Bool("log-smtp"),
			smtpPort, useSSL, tls, smtpBackend, b.Listener).ListenAndServe()
	}()

	var frontendMode string

	switch {
	case c.Bool("cli"):
		frontendMode = "cli"
	case c.Bool("noninteractive"):
		frontendMode = "noninteractive"
	default:
		frontendMode = "qt"
	}

	if frontendMode == "noninteractive" {
		<-(make(chan struct{}))
		return nil
	}

	// Bridge supports no-window option which we should use for autostart.
	b.Autostart.Exec = append(b.Autostart.Exec, "--no-window")

	// We want to remove old versions if the app exits successfully.
	b.AddTeardownAction(b.Versioner.RemoveOldVersions)

	// We want cookies to be saved to disk so they are loaded the next time.
	b.AddTeardownAction(b.CookieJar.PersistCookies)

	f := frontend.New(
		constants.Version,
		constants.BuildVersion,
		b.Name,
		frontendMode,
		!c.Bool("no-window"),
		b.CrashHandler,
		b.Locations,
		b.Settings,
		b.Listener,
		b.Updater,
		bridge,
		smtpBackend,
		b.Autostart,
		b,
	)

	// Watch for updates routine
	go func() {
		ticker := time.NewTicker(time.Hour)

		for {
			checkAndHandleUpdate(b.Updater, f, b.Settings.GetBool(settings.AutoUpdateKey))
			<-ticker.C
		}
	}()

	return f.Loop()
}

func checkAndHandleUpdate(u types.Updater, f frontend.Frontend, autoUpdate bool) {
	version, err := u.Check()
	if err != nil {
		logrus.WithError(err).Error("An error occurred while checking for updates")
		f.NotifySilentUpdateError(err)
		return
	}

	if !u.IsUpdateApplicable(version) {
		logrus.Debug("No need to update")
		return
	}

	logrus.WithField("version", version.Version).Info("An update is available")

	if !autoUpdate {
		f.NotifyManualUpdate(version, u.CanInstall(version))
		return
	}

	if !u.CanInstall(version) {
		logrus.Info("A manual update is required")
		f.NotifySilentUpdateError(updater.ErrManualUpdateRequired)
		return
	}

	if err := u.InstallUpdate(version); err != nil {
		logrus.WithError(err).Error("An error occurred while silent installing updates")
		f.NotifySilentUpdateError(err)
		return
	}

	f.NotifySilentUpdateInstalled()
}
