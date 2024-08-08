// Copyright (c) 2024 Proton AG
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

package app

import (
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/ProtonMail/gluon/async"
	"github.com/ProtonMail/proton-bridge/v3/internal/bridge"
	"github.com/ProtonMail/proton-bridge/v3/internal/constants"
	"github.com/ProtonMail/proton-bridge/v3/internal/cookies"
	"github.com/ProtonMail/proton-bridge/v3/internal/crash"
	"github.com/ProtonMail/proton-bridge/v3/internal/events"
	"github.com/ProtonMail/proton-bridge/v3/internal/focus"
	"github.com/ProtonMail/proton-bridge/v3/internal/frontend/theme"
	"github.com/ProtonMail/proton-bridge/v3/internal/locations"
	"github.com/ProtonMail/proton-bridge/v3/internal/logging"
	"github.com/ProtonMail/proton-bridge/v3/internal/sentry"
	"github.com/ProtonMail/proton-bridge/v3/internal/useragent"
	"github.com/ProtonMail/proton-bridge/v3/internal/vault"
	"github.com/ProtonMail/proton-bridge/v3/pkg/keychain"
	"github.com/ProtonMail/proton-bridge/v3/pkg/restarter"
	"github.com/elastic/go-sysinfo"
	"github.com/pkg/profile"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

// Visible flags.
const (
	flagCPUProfile      = "cpu-prof"
	flagCPUProfileShort = "p"

	flagTraceProfile      = "trace-prof"
	flagTraceProfileShort = "t"

	flagMemProfile      = "mem-prof"
	flagMemProfileShort = "m"

	flagLogLevel      = "log-level"
	flagLogLevelShort = "l"

	flagGRPC      = "grpc"
	flagGRPCShort = "g"

	flagCLI      = "cli"
	flagCLIShort = "c"

	flagNonInteractive      = "noninteractive"
	flagNonInteractiveShort = "n"

	flagLogIMAP = "log-imap"
	flagLogSMTP = "log-smtp"
)

// Hidden flags.
const (
	flagLauncher            = "launcher"
	flagNoWindow            = "no-window"
	flagParentPID           = "parent-pid"
	flagSoftwareRenderer    = "software-renderer"
	flagEnableKeychainTest  = "enable-keychain-test"
	flagDisableKeychainTest = "disable-keychain-test"
	FlagSessionID           = "session-id"
)

const (
	appUsage     = "Proton Mail IMAP and SMTP Bridge"
	appShortName = "bridge"
)

var cliFlagEnableKeychainTest = &cli.BoolFlag{ //nolint:gochecknoglobals
	Name:   flagEnableKeychainTest,
	Usage:  "Enable the keychain test",
	Hidden: true,
	Value:  false,
} //nolint:gochecknoglobals

var cliFlagDisableKeychainTest = &cli.BoolFlag{ //nolint:gochecknoglobals
	Name:   flagDisableKeychainTest,
	Usage:  "Disable the keychain test",
	Hidden: true,
	Value:  false,
}

func New() *cli.App {
	app := cli.NewApp()

	app.Name = constants.FullAppName
	app.Usage = appUsage
	app.Flags = []cli.Flag{
		&cli.BoolFlag{
			Name:    flagCPUProfile,
			Aliases: []string{flagCPUProfileShort},
			Usage:   "Generate CPU profile",
		},
		&cli.BoolFlag{
			Name:    flagTraceProfile,
			Aliases: []string{flagTraceProfileShort},
			Usage:   "Generate Trace profile",
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
			Name:    flagGRPC,
			Aliases: []string{flagGRPCShort},
			Usage:   "Start the gRPC service",
		},
		&cli.BoolFlag{
			Name:    flagCLI,
			Aliases: []string{flagCLIShort},
			Usage:   "Start the command line interface",
		},
		&cli.BoolFlag{
			Name:    flagNonInteractive,
			Aliases: []string{flagNonInteractiveShort},
			Usage:   "Start the app in non-interactive mode",
		},
		&cli.StringFlag{
			Name:  flagLogIMAP,
			Usage: "Enable logging of IMAP communications (all|client|server) (may contain decrypted data!)",
		},
		&cli.BoolFlag{
			Name:  flagLogSMTP,
			Usage: "Enable logging of SMTP communications (may contain decrypted data!)",
		},

		// Hidden flags
		&cli.BoolFlag{
			Name:   flagNoWindow,
			Usage:  "Don't show window after start",
			Hidden: true,
		},
		&cli.StringFlag{
			Name:   flagLauncher,
			Usage:  "The launcher used to start the app",
			Hidden: true,
		},
		&cli.IntFlag{
			Name:   flagParentPID,
			Usage:  "Process ID of the parent",
			Hidden: true,
			Value:  -1,
		},
		&cli.BoolFlag{
			Name:   flagSoftwareRenderer, // This flag is ignored by bridge, but should be passed to launcher in case of restart, so it need to be accepted by the CLI parser.
			Usage:  "GUI is using software renderer",
			Hidden: true,
			Value:  false,
		},
		&cli.StringFlag{
			Name:   FlagSessionID,
			Hidden: true,
		},
		// the two flags below were introduced by BRIDGE-116
		cliFlagEnableKeychainTest,
		cliFlagDisableKeychainTest,
	}

	app.Action = run

	return app
}

func run(c *cli.Context) error {
	// Get the current bridge version.
	version, err := semver.NewVersion(constants.Version)
	if err != nil {
		return fmt.Errorf("could not create version: %w", err)
	}

	// Create a user agent that will be used for all requests.
	identifier := useragent.New()

	// Create a new Sentry client that will be used to report crashes etc.
	reporter := sentry.NewReporter(constants.FullAppName, identifier)

	// Determine the exe that should be used to restart/autostart the app.
	// By default, this is the launcher, if used. Otherwise, we try to get
	// the current exe, and fall back to os.Args[0] if that fails.
	var exe string

	if launcher := c.String(flagLauncher); launcher != "" {
		exe = launcher
	} else if executable, err := os.Executable(); err == nil {
		exe = executable
	} else {
		exe = os.Args[0]
	}

	var logCloser io.Closer
	defer func() {
		_ = logging.Close(logCloser)
	}()

	// Restart the app if requested.
	err = withRestarter(exe, func(restarter *restarter.Restarter) error {
		// Handle crashes with various actions.
		return withCrashHandler(restarter, reporter, func(crashHandler *crash.Handler, quitCh <-chan struct{}) error {
			migrationErr := migrateOldVersions()

			// Run with profiling if requested.
			return withProfiler(c, func() error {
				// Load the locations where we store our files.
				return WithLocations(func(locations *locations.Locations) error {
					// Migrate the keychain helper.
					if err := migrateKeychainHelper(locations); err != nil {
						logrus.WithError(err).Error("Failed to migrate keychain helper")
					}

					// Initialize logging.
					return withLogging(c, crashHandler, locations, func(closer io.Closer) error {
						logCloser = closer

						// If there was an error during migration, log it now.
						if migrationErr != nil {
							logrus.WithError(migrationErr).Error("Failed to migrate old app data")
						}

						// Ensure we are the only instance running.
						settings, err := locations.ProvideSettingsPath()
						if err != nil {
							logrus.WithError(err).Error("Failed to get settings path")
						}

						return withSingleInstance(settings, locations.GetLockFile(), version, func() error {
							// Look for available keychains
							skipKeychainTest := checkSkipKeychainTest(c, settings)
							return WithKeychainList(crashHandler, skipKeychainTest, func(keychains *keychain.List) error {
								// Unlock the encrypted vault.
								return WithVault(locations, keychains, crashHandler, func(v *vault.Vault, insecure, corrupt bool) error {
									if !v.Migrated() {
										// Migrate old settings into the vault.
										if err := migrateOldSettings(v); err != nil {
											logrus.WithError(err).Error("Failed to migrate old settings")
										}

										// Migrate old accounts into the vault.
										if err := migrateOldAccounts(locations, keychains, v); err != nil {
											logrus.WithError(err).Error("Failed to migrate old accounts")
										}

										// The vault has been migrated.
										if err := v.SetMigrated(); err != nil {
											logrus.WithError(err).Error("Failed to mark vault as migrated")
										}
									}

									logrus.WithFields(logrus.Fields{
										"lastVersion": v.GetLastVersion().String(),
										"showAllMail": v.GetShowAllMail(),
										"updateCh":    v.GetUpdateChannel(),
										"autoUpdate":  v.GetAutoUpdate(),
										"rollout":     v.GetUpdateRollout(),
										"DoH":         v.GetProxyAllowed(),
									}).Info("Vault loaded")

									// Load the cookies from the vault.
									return withCookieJar(v, func(cookieJar http.CookieJar) error {
										// Create a new bridge instance.
										return withBridge(c, exe, locations, version, identifier, crashHandler, reporter, v, cookieJar, keychains, func(b *bridge.Bridge, eventCh <-chan events.Event) error {
											if insecure {
												logrus.Warn("The vault key could not be retrieved; the vault will not be encrypted")
												b.PushError(bridge.ErrVaultInsecure)
											}

											if corrupt {
												logrus.Warn("The vault is corrupt and has been wiped")
												b.PushError(bridge.ErrVaultCorrupt)
											}

											// Remove old updates files
											b.RemoveOldUpdates()

											// Run the frontend.
											return runFrontend(c, crashHandler, restarter, locations, b, eventCh, quitCh, c.Int(flagParentPID))
										})
									})
								})
							})
						})
					})
				})
			})
		})
	})

	// if an error occurs, it must be logged now because we're about to close the log file.
	if err != nil {
		logrus.Fatal(err)
	}

	return err
}

// If there's another instance already running, try to raise it and exit.
func withSingleInstance(settingPath, lockFile string, version *semver.Version, fn func() error) error {
	logrus.Debug("Checking for other instances")
	defer logrus.Debug("Single instance stopped")

	lock, err := checkSingleInstance(settingPath, lockFile, version)
	if err != nil {
		logrus.Info("Another instance is already running; raising it")

		if ok := focus.TryRaise(settingPath); !ok {
			return fmt.Errorf("another instance is already running but it could not be raised")
		}

		logrus.Info("The other instance has been raised")

		return nil
	}

	defer func() {
		if err := lock.Close(); err != nil {
			logrus.WithError(err).Error("Failed to close lock file")
		}
	}()

	return fn()
}

// Initialize our logging system.
func withLogging(c *cli.Context, crashHandler *crash.Handler, locations *locations.Locations, fn func(closer io.Closer) error) error {
	logrus.Debug("Initializing logging")
	defer logrus.Debug("Logging stopped")

	// Get a place to keep our logs.
	logsPath, err := locations.ProvideLogsPath()
	if err != nil {
		return fmt.Errorf("could not provide logs path: %w", err)
	}

	logrus.WithField("path", logsPath).Debug("Received logs path")

	// Initialize logging.
	sessionID := logging.NewSessionIDFromString(c.String(FlagSessionID))
	var closer io.Closer
	if closer, err = logging.Init(
		logsPath,
		sessionID,
		logging.BridgeShortAppName,
		logging.DefaultMaxLogFileSize,
		logging.DefaultPruningSize,
		c.String(flagLogLevel),
	); err != nil {
		return fmt.Errorf("could not initialize logging: %w", err)
	}

	// Ensure we dump a stack trace if we crash.
	crashHandler.AddRecoveryAction(logging.DumpStackTrace(logsPath, sessionID, appShortName))

	logrus.
		WithField("appName", constants.FullAppName).
		WithField("version", constants.Version).
		WithField("revision", constants.Revision).
		WithField("tag", constants.Tag).
		WithField("build", constants.BuildTime).
		WithField("runtime", runtime.GOOS).
		WithField("args", os.Args).
		WithField("SentryID", sentry.GetProtectedHostname()).
		Info("Run app")

	now := time.Now()
	logrus.
		WithField("timeZone", now.Format("MST")).
		WithField("offset", now.Format("-07:00:00")).
		Info("Time zone info")

	host, err := sysinfo.Host()
	if err != nil {
		logrus.WithError(err).Error("Could not retrieve operating system info")
	} else {
		osInfo := host.Info().OS
		logrus.
			WithField("name", osInfo.Name).
			WithField("version", osInfo.Version).
			WithField("build", osInfo.Build).
			Info("Operating system info")
	}

	return fn(closer)
}

// WithLocations provides access to locations where we store our files.
func WithLocations(fn func(*locations.Locations) error) error {
	logrus.Debug("Creating locations")
	defer logrus.Debug("Locations stopped")

	// Create a locations provider to determine where to store our files.
	provider, err := locations.NewDefaultProvider(filepath.Join(constants.VendorName, constants.ConfigName))
	if err != nil {
		return fmt.Errorf("could not create locations provider: %w", err)
	}

	// Create a new locations object that will be used to provide paths to store files.
	return fn(locations.New(provider, constants.ConfigName))
}

// Start profiling if requested.
func withProfiler(c *cli.Context, fn func() error) error {
	defer logrus.Debug("Profiler stopped")

	if c.Bool(flagCPUProfile) {
		logrus.Debug("Running with CPU profiling")
		defer profile.Start(profile.CPUProfile, profile.ProfilePath(".")).Stop()
	}

	if c.Bool(flagTraceProfile) {
		logrus.Debug("Running with Trace profiling")
		defer profile.Start(profile.TraceProfile, profile.ProfilePath(".")).Stop()
	}

	if c.Bool(flagMemProfile) {
		logrus.Debug("Running with memory profiling")
		defer profile.Start(profile.MemProfile, profile.MemProfileAllocs, profile.ProfilePath(".")).Stop()
	}

	return fn()
}

// Restart the app if necessary.
func withRestarter(exe string, fn func(*restarter.Restarter) error) error {
	logrus.Debug("Creating restarter")
	defer logrus.Debug("Restarter stopped")

	restarter := restarter.New(exe)
	defer restarter.Restart()

	return fn(restarter)
}

// Handle crashes if they occur.
func withCrashHandler(restarter *restarter.Restarter, reporter *sentry.Reporter, fn func(*crash.Handler, <-chan struct{}) error) error {
	logrus.Debug("Creating crash handler")
	defer logrus.Debug("Crash handler stopped")

	crashHandler := crash.NewHandler(crash.ShowErrorNotification(constants.FullAppName))
	defer async.HandlePanic(crashHandler)

	// On crash, send crash report to Sentry.
	crashHandler.AddRecoveryAction(reporter.ReportException)

	// On crash, notify the user and restart the app.
	crashHandler.AddRecoveryAction(crash.ShowErrorNotification(constants.FullAppName))

	// On crash, restart the app.
	crashHandler.AddRecoveryAction(func(any) error { restarter.Set(true, true); return nil })

	// quitCh is closed when the app is quitting.
	quitCh := make(chan struct{})

	// On crash, quit the app.
	crashHandler.AddRecoveryAction(func(any) error { close(quitCh); return nil })

	return fn(crashHandler, quitCh)
}

// Use a custom cookie jar to persist values across runs.
func withCookieJar(vault *vault.Vault, fn func(http.CookieJar) error) error {
	logrus.Debug("Creating cookie jar")
	defer logrus.Debug("Cookie jar stopped")

	// Create the underlying cookie jar.
	jar, err := cookiejar.New(nil)
	if err != nil {
		return fmt.Errorf("could not create cookie jar: %w", err)
	}

	// Create the cookie jar which persists to the vault.
	persister, err := cookies.NewCookieJar(jar, vault)
	if err != nil {
		return fmt.Errorf("could not create cookie jar: %w", err)
	}

	if err := setDeviceCookies(persister); err != nil {
		return fmt.Errorf("could not set device cookies: %w", err)
	}

	// Persist the cookies to the vault when we close.
	defer func() {
		logrus.Debug("Persisting cookies")

		if err := persister.PersistCookies(); err != nil {
			logrus.WithError(err).Error("Failed to persist cookies")
		}
	}()

	return fn(persister)
}

// WithKeychainList init the list of usable keychains.
func WithKeychainList(panicHandler async.PanicHandler, skipKeychainTest bool, fn func(*keychain.List) error) error {
	logrus.Debug("Creating keychain list")
	defer logrus.Debug("Keychain list stop")
	defer async.HandlePanic(panicHandler)
	return fn(keychain.NewList(skipKeychainTest))
}

func setDeviceCookies(jar *cookies.Jar) error {
	url, err := url.Parse(constants.APIHost)
	if err != nil {
		return err
	}

	for name, value := range map[string]string{
		"hhn": sentry.GetProtectedHostname(),
		"tz":  sentry.GetTimeZone(),
		"lng": sentry.GetSystemLang(),
		"clr": string(theme.DefaultTheme()),
	} {
		jar.SetCookies(url, []*http.Cookie{{Name: name, Value: value, Secure: true}})
	}

	return nil
}

func checkSkipKeychainTest(c *cli.Context, settingsDir string) bool {
	if runtime.GOOS != "darwin" {
		return false
	}

	enable := c.Bool(flagEnableKeychainTest)
	disable := c.Bool(flagDisableKeychainTest)

	skip, err := vault.GetShouldSkipKeychainTest(settingsDir)
	if err != nil {
		logrus.WithError(err).Error("Could not load keychain settings.")
	}

	if (!enable) && (!disable) {
		return skip
	}

	// if both switches are passed, 'enable' has priority
	if disable {
		skip = true
	}
	if enable {
		skip = false
	}

	if err := vault.SetShouldSkipKeychainTest(settingsDir, skip); err != nil {
		logrus.WithError(err).Error("Could not save keychain settings.")
	}

	return skip
}
