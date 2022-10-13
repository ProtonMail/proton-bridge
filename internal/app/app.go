package app

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"os"
	"path/filepath"

	"github.com/Masterminds/semver/v3"
	"github.com/ProtonMail/proton-bridge/v2/internal/bridge"
	"github.com/ProtonMail/proton-bridge/v2/internal/constants"
	"github.com/ProtonMail/proton-bridge/v2/internal/cookies"
	"github.com/ProtonMail/proton-bridge/v2/internal/crash"
	"github.com/ProtonMail/proton-bridge/v2/internal/events"
	"github.com/ProtonMail/proton-bridge/v2/internal/focus"
	"github.com/ProtonMail/proton-bridge/v2/internal/locations"
	"github.com/ProtonMail/proton-bridge/v2/internal/sentry"
	"github.com/ProtonMail/proton-bridge/v2/internal/useragent"
	"github.com/ProtonMail/proton-bridge/v2/internal/vault"
	"github.com/ProtonMail/proton-bridge/v2/pkg/restarter"
	"github.com/pkg/profile"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

// Visible flags
const (
	flagCPUProfile      = "cpu-prof"
	flagCPUProfileShort = "p"

	flagMemProfile      = "mem-prof"
	flagMemProfileShort = "m"

	flagLogLevel      = "log-level"
	flagLogLevelShort = "l"

	flagCLI      = "cli"
	flagCLIShort = "c"

	flagNonInteractive = "non-interactive"

	flagLogIMAP = "log-imap"
	flagLogSMTP = "log-smtp"
)

// Hidden flags
const (
	flagLauncher = "launcher"
	flagNoWindow = "no-window"
)

const (
	appUsage = "Proton Mail IMAP and SMTP Bridge"
)

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
			Name:    flagCLI,
			Aliases: []string{flagCLIShort},
			Usage:   "Use command line interface",
		},
		&cli.BoolFlag{
			Name:  flagNonInteractive,
			Usage: "Run the app in non-interactive mode",
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
		&cli.BoolFlag{
			Name:   flagLauncher,
			Usage:  "The launcher used to start the app",
			Hidden: true,
		},
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
	reporter := sentry.NewReporter(constants.FullAppName, constants.Version, identifier)

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

	// Run with profiling if requested.
	return withProfiler(c, func() error {
		// Restart the app if requested.
		return withRestarter(exe, func(restarter *restarter.Restarter) error {
			// Handle crashes with various actions.
			return withCrashHandler(restarter, reporter, func(crashHandler *crash.Handler) error {
				// Load the locations where we store our files.
				return withLocations(func(locations *locations.Locations) error {
					// Initialize logging.
					return withLogging(c, crashHandler, locations, func() error {
						// Ensure we are the only instance running.
						return withSingleInstance(locations, version, func() error {
							// Unlock the encrypted vault.
							return withVault(locations, func(vault *vault.Vault, insecure, corrupt bool) error {
								// Load the cookies from the vault.
								return withCookieJar(vault, func(cookieJar http.CookieJar) error {
									// Create a new bridge instance.
									return withBridge(c, exe, locations, version, identifier, reporter, vault, cookieJar, func(b *bridge.Bridge, eventCh <-chan events.Event) error {
										if insecure {
											logrus.Warn("The vault key could not be retrieved; the vault will not be encrypted")
											b.PushError(bridge.ErrVaultInsecure)
										}

										if corrupt {
											logrus.Warn("The vault is corrupt and has been wiped")
											b.PushError(bridge.ErrVaultCorrupt)
										}

										// Run the frontend.
										return runFrontend(c, crashHandler, restarter, locations, b, eventCh)
									})
								})
							})
						})
					})
				})
			})
		})
	})
}

// If there's another instance already running, try to raise it and exit.
func withSingleInstance(locations *locations.Locations, version *semver.Version, fn func() error) error {
	lock, err := checkSingleInstance(locations.GetLockFile(), version)
	if err != nil {
		if ok := focus.TryRaise(); !ok {
			return fmt.Errorf("another instance is already running but it could not be raised")
		}

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
func withLogging(c *cli.Context, crashHandler *crash.Handler, locations *locations.Locations, fn func() error) error {
	if err := initLogging(c, locations, crashHandler); err != nil {
		return fmt.Errorf("could not initialize logging: %w", err)
	}

	// TODO: Add teardown actions (clean the log directory?)

	return fn()
}

// Provide access to locations where we store our files.
func withLocations(fn func(*locations.Locations) error) error {
	// Create a locations provider to determine where to store our files.
	provider, err := locations.NewDefaultProvider(filepath.Join(constants.VendorName, constants.ConfigName))
	if err != nil {
		return fmt.Errorf("could not create locations provider: %w", err)
	}

	// Create a new locations object that will be used to provide paths to store files.
	locations := locations.New(provider, constants.ConfigName)
	defer locations.Clean()

	return fn(locations)
}

// Start profiling if requested.
func withProfiler(c *cli.Context, fn func() error) error {
	if c.Bool(flagCPUProfile) {
		defer profile.Start(profile.CPUProfile, profile.ProfilePath(".")).Stop()
	}

	if c.Bool(flagMemProfile) {
		defer profile.Start(profile.MemProfile, profile.MemProfileAllocs, profile.ProfilePath(".")).Stop()
	}

	return fn()
}

// Restart the app if necessary.
func withRestarter(exe string, fn func(*restarter.Restarter) error) error {
	restarter := restarter.New(exe)
	defer restarter.Restart()

	return fn(restarter)
}

// Handle crashes if they occur.
func withCrashHandler(restarter *restarter.Restarter, reporter *sentry.Reporter, fn func(*crash.Handler) error) error {
	crashHandler := crash.NewHandler(crash.ShowErrorNotification(constants.FullAppName))
	defer crashHandler.HandlePanic()

	// On crash, send crash report to Sentry.
	crashHandler.AddRecoveryAction(reporter.ReportException)

	// On crash, notify the user and restart the app.
	crashHandler.AddRecoveryAction(crash.ShowErrorNotification(constants.FullAppName))

	// On crash, restart the app.
	crashHandler.AddRecoveryAction(func(r any) error { restarter.Set(true, true); return nil })

	return fn(crashHandler)
}

// Use a custom cookie jar to persist values across runs.
func withCookieJar(vault *vault.Vault, fn func(http.CookieJar) error) error {
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

	// Persist the cookies to the vault when we close.
	defer func() {
		if err := persister.PersistCookies(); err != nil {
			logrus.WithError(err).Error("Failed to persist cookies")
		}
	}()

	return fn(persister)
}
