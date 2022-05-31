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

// Package bridge implements the bridge CLI application.
package bridge

import (
	"crypto/tls"
	"time"

	"github.com/ProtonMail/proton-bridge/v2/internal/api"
	"github.com/ProtonMail/proton-bridge/v2/internal/app/base"
	pkgBridge "github.com/ProtonMail/proton-bridge/v2/internal/bridge"
	"github.com/ProtonMail/proton-bridge/v2/internal/config/settings"
	pkgTLS "github.com/ProtonMail/proton-bridge/v2/internal/config/tls"
	"github.com/ProtonMail/proton-bridge/v2/internal/constants"
	"github.com/ProtonMail/proton-bridge/v2/internal/frontend"
	"github.com/ProtonMail/proton-bridge/v2/internal/frontend/types"
	"github.com/ProtonMail/proton-bridge/v2/internal/imap"
	"github.com/ProtonMail/proton-bridge/v2/internal/smtp"
	"github.com/ProtonMail/proton-bridge/v2/internal/store"
	"github.com/ProtonMail/proton-bridge/v2/internal/store/cache"
	"github.com/ProtonMail/proton-bridge/v2/internal/updater"
	"github.com/ProtonMail/proton-bridge/v2/pkg/message"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

const (
	flagLogIMAP        = "log-imap"
	flagLogSMTP        = "log-smtp"
	flagNonInteractive = "noninteractive"

	// Memory cache was estimated by empirical usage in past and it was set to 100MB.
	// NOTE: This value must not be less than maximal size of one email (~30MB).
	inMemoryCacheLimnit = 100 * (1 << 20)
)

func New(base *base.Base) *cli.App {
	app := base.NewApp(mailLoop)

	app.Flags = append(app.Flags, []cli.Flag{
		&cli.StringFlag{
			Name:  flagLogIMAP,
			Usage: "Enable logging of IMAP communications (all|client|server) (may contain decrypted data!)",
		},
		&cli.BoolFlag{
			Name:  flagLogSMTP,
			Usage: "Enable logging of SMTP communications (may contain decrypted data!)",
		},
		&cli.BoolFlag{
			Name:  flagNonInteractive,
			Usage: "Start Bridge entirely noninteractively",
		},
	}...)

	return app
}

func mailLoop(b *base.Base, c *cli.Context) error { //nolint:funlen
	tlsConfig, err := loadTLSConfig(b)
	if err != nil {
		return err
	}

	// GODT-1481: Always turn off reporting of unencrypted recipient in v2.
	b.Settings.SetBool(settings.ReportOutgoingNoEncKey, false)

	cache, cacheErr := loadMessageCache(b)
	if cacheErr != nil {
		logrus.WithError(cacheErr).Error("Could not load local cache.")
	}

	builder := message.NewBuilder(
		b.Settings.GetInt(settings.FetchWorkers),
		b.Settings.GetInt(settings.AttachmentWorkers),
	)

	bridge := pkgBridge.New(
		b.Locations,
		b.Cache,
		b.Settings,
		b.SentryReporter,
		b.CrashHandler,
		b.Listener,
		cache,
		builder,
		b.CM,
		b.Creds,
		b.Updater,
		b.Versioner,
		b.Autostart,
	)
	imapBackend := imap.NewIMAPBackend(b.CrashHandler, b.Listener, b.Cache, b.Settings, bridge)
	smtpBackend := smtp.NewSMTPBackend(b.CrashHandler, b.Listener, b.Settings, bridge)

	if cacheErr != nil {
		bridge.AddError(pkgBridge.ErrLocalCacheUnavailable)
	}

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
			b.CrashHandler,
			c.Bool(flagLogSMTP),
			smtpPort, useSSL, tlsConfig, smtpBackend, b.Listener).ListenAndServe()
	}()

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
		!c.Bool(base.FlagNoWindow),
		b.CrashHandler,
		b.Locations,
		b.Settings,
		b.Listener,
		b.Updater,
		b.UserAgent,
		bridge,
		smtpBackend,
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

// loadMessageCache loads local cache in case it is enabled in settings and available.
// In any other case it is returning in-memory cache. Could also return an error in case
// local cache is enabled but unavailable (in-memory cache will be returned nevertheless).
func loadMessageCache(b *base.Base) (cache.Cache, error) {
	if !b.Settings.GetBool(settings.CacheEnabledKey) {
		return cache.NewInMemoryCache(inMemoryCacheLimnit), nil
	}

	var compressor cache.Compressor

	// NOTE(GODT-1158): Changing compression is not an option currently
	// available for user but, if user changes compression setting we have
	// to nuke the cache.
	if b.Settings.GetBool(settings.CacheCompressionKey) {
		compressor = &cache.GZipCompressor{}
	} else {
		compressor = &cache.NoopCompressor{}
	}

	var path string

	if customPath := b.Settings.Get(settings.CacheLocationKey); customPath != "" {
		path = customPath
	} else {
		path = b.Cache.GetDefaultMessageCacheDir()
		// Store path so it will allways persist if default location
		// will be changed in new version.
		b.Settings.Set(settings.CacheLocationKey, path)
	}

	// To prevent memory peaks we set maximal write concurency for store
	// build jobs.
	store.SetBuildAndCacheJobLimit(b.Settings.GetInt(settings.CacheConcurrencyWrite))

	messageCache, err := cache.NewOnDiskCache(path, compressor, cache.Options{
		MinFreeAbs:      uint64(b.Settings.GetInt(settings.CacheMinFreeAbsKey)),
		MinFreeRat:      b.Settings.GetFloat64(settings.CacheMinFreeRatKey),
		ConcurrentRead:  b.Settings.GetInt(settings.CacheConcurrencyRead),
		ConcurrentWrite: b.Settings.GetInt(settings.CacheConcurrencyWrite),
	})
	if err != nil {
		return cache.NewInMemoryCache(inMemoryCacheLimnit), err
	}

	return messageCache, nil
}
