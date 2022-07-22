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

// Package base implements a common application base currently shared by bridge and IE.
// The base includes the following:
//  - access to standard filesystem locations like config, cache, logging dirs
//  - an extensible crash handler
//  - versioned cache directory
//  - persistent settings
//  - event listener
//  - credentials store
//  - pmapi Manager
// In addition, the base initialises logging and reacts to command line arguments
// which control the log verbosity and enable cpu/memory profiling.
package base

import (
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/ProtonMail/go-autostart"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/v2/internal/api"
	"github.com/ProtonMail/proton-bridge/v2/internal/config/cache"
	"github.com/ProtonMail/proton-bridge/v2/internal/config/settings"
	"github.com/ProtonMail/proton-bridge/v2/internal/config/tls"
	"github.com/ProtonMail/proton-bridge/v2/internal/config/useragent"
	"github.com/ProtonMail/proton-bridge/v2/internal/constants"
	"github.com/ProtonMail/proton-bridge/v2/internal/cookies"
	"github.com/ProtonMail/proton-bridge/v2/internal/crash"
	"github.com/ProtonMail/proton-bridge/v2/internal/events"
	"github.com/ProtonMail/proton-bridge/v2/internal/locations"
	"github.com/ProtonMail/proton-bridge/v2/internal/logging"
	"github.com/ProtonMail/proton-bridge/v2/internal/sentry"
	"github.com/ProtonMail/proton-bridge/v2/internal/updater"
	"github.com/ProtonMail/proton-bridge/v2/internal/users/credentials"
	"github.com/ProtonMail/proton-bridge/v2/internal/versioner"
	"github.com/ProtonMail/proton-bridge/v2/pkg/keychain"
	"github.com/ProtonMail/proton-bridge/v2/pkg/listener"
	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

const (
	flagCPUProfile      = "cpu-prof"
	flagCPUProfileShort = "p"
	flagMemProfile      = "mem-prof"
	flagMemProfileShort = "m"
	flagLogLevel        = "log-level"
	flagLogLevelShort   = "l"
	// FlagCLI indicate to start with command line interface.
	FlagCLI      = "cli"
	flagCLIShort = "c"
	flagRestart  = "restart"
	FlagLauncher = "launcher"
	FlagNoWindow = "no-window"
)

type Base struct {
	SentryReporter *sentry.Reporter
	CrashHandler   *crash.Handler
	Locations      *locations.Locations
	Settings       *settings.Settings
	Lock           *os.File
	Cache          *cache.Cache
	Listener       listener.Listener
	Creds          *credentials.Store
	CM             pmapi.Manager
	CookieJar      *cookies.Jar
	UserAgent      *useragent.UserAgent
	Updater        *updater.Updater
	Versioner      *versioner.Versioner
	TLS            *tls.TLS
	Autostart      *autostart.App

	Name    string // the app's name
	usage   string // the app's usage description
	command string // the command used to launch the app (either the exe path or the launcher path)
	restart bool   // whether the app is currently set to restart

	teardown []func() error // actions to perform when app is exiting
}

func New( //nolint:funlen
	appName,
	appUsage,
	configName,
	updateURLName,
	keychainName,
	cacheVersion string,
) (*Base, error) {
	userAgent := useragent.New()

	sentryReporter := sentry.NewReporter(appName, constants.Version, userAgent)

	crashHandler := crash.NewHandler(
		sentryReporter.ReportException,
		crash.ShowErrorNotification(appName),
	)
	defer crashHandler.HandlePanic()

	rand.Seed(time.Now().UnixNano())
	os.Args = StripProcessSerialNumber(os.Args)

	locationsProvider, err := locations.NewDefaultProvider(filepath.Join(constants.VendorName, configName))
	if err != nil {
		return nil, err
	}

	locations := locations.New(locationsProvider, configName)

	logsPath, err := locations.ProvideLogsPath()
	if err != nil {
		return nil, err
	}
	if err := logging.Init(logsPath); err != nil {
		return nil, err
	}
	crashHandler.AddRecoveryAction(logging.DumpStackTrace(logsPath))

	if err := migrateFiles(configName); err != nil {
		logrus.WithError(err).Warn("Old config files could not be migrated")
	}

	if err := locations.Clean(); err != nil {
		return nil, err
	}

	settingsPath, err := locations.ProvideSettingsPath()
	if err != nil {
		return nil, err
	}
	settingsObj := settings.New(settingsPath)

	lock, err := checkSingleInstance(locations.GetLockFile(), settingsObj)
	if err != nil {
		logrus.WithError(err).Warnf("%v is already running", appName)
		return nil, api.CheckOtherInstanceAndFocus(settingsObj.GetInt(settings.APIPortKey))
	}

	if err := migrateRebranding(settingsObj, keychainName); err != nil {
		logrus.WithError(err).Warn("Rebranding migration failed")
	}

	cachePath, err := locations.ProvideCachePath()
	if err != nil {
		return nil, err
	}
	cache, err := cache.New(cachePath, cacheVersion)
	if err != nil {
		return nil, err
	}
	if err := cache.RemoveOldVersions(); err != nil {
		return nil, err
	}

	listener := listener.New()
	events.SetupEvents(listener)

	// If we can't load the keychain for whatever reason,
	// we signal to frontend and supply a dummy keychain that always returns errors.
	kc, err := keychain.NewKeychain(settingsObj, keychainName)
	if err != nil {
		listener.Emit(events.CredentialsErrorEvent, err.Error())
		kc = keychain.NewMissingKeychain()
	}

	cfg := pmapi.NewConfig(configName, constants.Version)
	cfg.GetUserAgent = userAgent.String
	cfg.UpgradeApplicationHandler = func() { listener.Emit(events.UpgradeApplicationEvent, "") }
	cfg.TLSIssueHandler = func() { listener.Emit(events.TLSCertIssue, "") }

	cm := pmapi.New(cfg)

	sentryReporter.SetClientFromManager(cm)

	cm.AddConnectionObserver(pmapi.NewConnectionObserver(
		func() { listener.Emit(events.InternetConnChangedEvent, events.InternetOff) },
		func() { listener.Emit(events.InternetConnChangedEvent, events.InternetOn) },
	))

	jar, err := cookies.NewCookieJar(settingsObj)
	if err != nil {
		return nil, err
	}

	cm.SetCookieJar(jar)

	key, err := crypto.NewKeyFromArmored(updater.DefaultPublicKey)
	if err != nil {
		return nil, err
	}

	kr, err := crypto.NewKeyRing(key)
	if err != nil {
		return nil, err
	}

	updatesDir, err := locations.ProvideUpdatesPath()
	if err != nil {
		return nil, err
	}

	versioner := versioner.New(updatesDir)
	installer := updater.NewInstaller(versioner)
	updater := updater.New(
		cm,
		installer,
		settingsObj,
		kr,
		semver.MustParse(constants.Version),
		updateURLName,
		runtime.GOOS,
	)

	exe, err := os.Executable()
	if err != nil {
		return nil, err
	}

	autostart := &autostart.App{
		Name:        startupNameForRebranding(appName),
		DisplayName: appName,
		Exec:        []string{exe, "--" + FlagNoWindow},
	}

	return &Base{
		SentryReporter: sentryReporter,
		CrashHandler:   crashHandler,
		Locations:      locations,
		Settings:       settingsObj,
		Lock:           lock,
		Cache:          cache,
		Listener:       listener,
		Creds:          credentials.NewStore(kc),
		CM:             cm,
		CookieJar:      jar,
		UserAgent:      userAgent,
		Updater:        updater,
		Versioner:      versioner,
		TLS:            tls.New(settingsPath),
		Autostart:      autostart,

		Name:  appName,
		usage: appUsage,

		// By default, the command is the app's executable.
		// This can be changed at runtime by using the "--launcher" flag.
		command: exe,
	}, nil
}

func (b *Base) NewApp(mainLoop func(*Base, *cli.Context) error) *cli.App {
	app := cli.NewApp()

	app.Name = b.Name
	app.Usage = b.usage
	app.Version = constants.Version
	app.Action = b.wrapMainLoop(mainLoop)
	app.Flags = []cli.Flag{
		&cli.BoolFlag{
			Name:    flagCPUProfile,
			Aliases: []string{flagCPUProfileShort},
			Usage:   "Generate CPU profile",
		},
		&cli.BoolFlag{
			Name:    flagMemProfile,
			Aliases: []string{flagMemProfileShort},
			Usage:   "Generate memory profile",
		},
		&cli.StringFlag{
			Name:    flagLogLevel,
			Aliases: []string{flagLogLevelShort},
			Usage:   "Set the log level (one of panic, fatal, error, warn, info, debug)",
		},
		&cli.BoolFlag{
			Name:    FlagCLI,
			Aliases: []string{flagCLIShort},
			Usage:   "Use command line interface",
		},
		&cli.BoolFlag{
			Name:  FlagNoWindow,
			Usage: "Don't show window after start",
		},
		&cli.StringFlag{
			Name:   flagRestart,
			Usage:  "The number of times the application has already restarted",
			Hidden: true,
		},
		&cli.StringFlag{
			Name:   FlagLauncher,
			Usage:  "The launcher to use to restart the application",
			Hidden: true,
		},
	}

	return app
}

// SetToRestart sets the app to restart the next time it is closed.
func (b *Base) SetToRestart() {
	b.restart = true
}

// AddTeardownAction adds an action to perform during app teardown.
func (b *Base) AddTeardownAction(fn func() error) {
	b.teardown = append(b.teardown, fn)
}

func (b *Base) wrapMainLoop(appMainLoop func(*Base, *cli.Context) error) cli.ActionFunc { //nolint:funlen
	return func(c *cli.Context) error {
		defer b.CrashHandler.HandlePanic()
		defer func() { _ = b.Lock.Close() }()

		// If launcher was used to start the app, use that for restart
		// and autostart.
		if launcher := c.String(FlagLauncher); launcher != "" {
			b.command = launcher
			// Bridge supports no-window option which we should use
			// for autostart.
			b.Autostart.Exec = []string{launcher, "--" + FlagNoWindow}
		}

		if c.Bool(flagCPUProfile) {
			startCPUProfile()
			defer pprof.StopCPUProfile()
		}

		if c.Bool(flagMemProfile) {
			defer makeMemoryProfile()
		}

		logging.SetLevel(c.String(flagLogLevel))
		b.CM.SetLogging(logrus.WithField("pkg", "pmapi"), logrus.GetLevel() == logrus.TraceLevel)

		logrus.
			WithField("appName", b.Name).
			WithField("version", constants.Version).
			WithField("revision", constants.Revision).
			WithField("build", constants.BuildTime).
			WithField("runtime", runtime.GOOS).
			WithField("args", os.Args).
			Info("Run app")

		b.CrashHandler.AddRecoveryAction(func(interface{}) error {
			sentry.Flush(2 * time.Second)

			if c.Int(flagRestart) > maxAllowedRestarts {
				logrus.
					WithField("restart", c.Int("restart")).
					Warn("Not restarting, already restarted too many times")
				os.Exit(1)

				return nil
			}

			return b.restartApp(true)
		})

		if err := appMainLoop(b, c); err != nil {
			return err
		}

		if err := b.doTeardown(); err != nil {
			return err
		}

		if b.restart {
			return b.restartApp(false)
		}

		return nil
	}
}

func (b *Base) doTeardown() error {
	for _, action := range b.teardown {
		if err := action(); err != nil {
			return err
		}
	}

	return nil
}
