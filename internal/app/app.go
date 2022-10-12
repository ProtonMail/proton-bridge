package app

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"path/filepath"

	"github.com/ProtonMail/proton-bridge/v2/internal/bridge"
	"github.com/ProtonMail/proton-bridge/v2/internal/constants"
	"github.com/ProtonMail/proton-bridge/v2/internal/cookies"
	"github.com/ProtonMail/proton-bridge/v2/internal/crash"
	"github.com/ProtonMail/proton-bridge/v2/internal/focus"
	bridgeCLI "github.com/ProtonMail/proton-bridge/v2/internal/frontend/cli"
	"github.com/ProtonMail/proton-bridge/v2/internal/frontend/grpc"
	"github.com/ProtonMail/proton-bridge/v2/internal/locations"
	"github.com/ProtonMail/proton-bridge/v2/internal/sentry"
	"github.com/ProtonMail/proton-bridge/v2/internal/useragent"
	"github.com/ProtonMail/proton-bridge/v2/internal/vault"
	"github.com/ProtonMail/proton-bridge/v2/pkg/restarter"
	"github.com/pkg/profile"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

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

	// Hidden flags
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
	// If there's another instance already running, try to raise it and exit.
	if raised := focus.TryRaise(); raised {
		return nil
	}

	// Create a user agent that will be used for all requests.
	identifier := useragent.New()

	// Create a new Sentry client that will be used to report crashes etc.
	reporter := sentry.NewReporter(constants.FullAppName, constants.Version, identifier)

	// Run with profiling if requested.
	return withProfiler(c, func() error {
		// Restart the app if requested.
		return withRestarter(func(restarter *restarter.Restarter) error {
			// Handle crashes with various actions.
			return withCrashHandler(restarter, reporter, func(crashHandler *crash.Handler) error {
				// Load the locations where we store our files.
				return withLocations(func(locations *locations.Locations) error {
					// Initialize the logging.
					if err := initLogging(c, locations, crashHandler); err != nil {
						return fmt.Errorf("could not initialize logging: %w", err)
					}

					// Unlock the encrypted vault.
					return withVault(locations, func(vault *vault.Vault, insecure, corrupt bool) error {
						// Load the cookies from the vault.
						return withCookieJar(vault, func(cookieJar http.CookieJar) error {
							// Create a new bridge instance.
							return withBridge(c, locations, identifier, reporter, vault, cookieJar, func(b *bridge.Bridge) error {
								if insecure {
									logrus.Warn("The vault key could not be retrieved; the vault will not be encrypted")
									b.PushError(bridge.ErrVaultInsecure)
								}

								if corrupt {
									logrus.Warn("The vault is corrupt and has been wiped")
									b.PushError(bridge.ErrVaultCorrupt)
								}

								switch {
								case c.Bool(flagCLI):
									return bridgeCLI.New(b).Loop()

								case c.Bool(flagNonInteractive):
									select {}

								default:
									service, err := grpc.NewService(crashHandler, restarter, locations, b, !c.Bool(flagNoWindow))
									if err != nil {
										return fmt.Errorf("could not create service: %w", err)
									}

									return service.Loop()
								}
							})
						})
					})
				})
			})
		})
	})
}

func withLocations(fn func(*locations.Locations) error) error {
	// Create a locations provider to determine where to store our files.
	provider, err := locations.NewDefaultProvider(filepath.Join(constants.VendorName, constants.ConfigName))
	if err != nil {
		return fmt.Errorf("could not create locations provider: %w", err)
	}

	// Create a new locations object that will be used to provide paths to store files.
	locations := locations.New(provider, constants.ConfigName)

	// TODO: Add teardown actions (removing the lock file, etc.)

	return fn(locations)
}

func withProfiler(c *cli.Context, fn func() error) error {
	// Start CPU profile if requested.
	if c.Bool(flagCPUProfile) {
		defer profile.Start(profile.CPUProfile, profile.ProfilePath(".")).Stop()
	}

	// Start memory profile if requested.
	if c.Bool(flagMemProfile) {
		defer profile.Start(profile.MemProfile, profile.MemProfileAllocs, profile.ProfilePath(".")).Stop()
	}

	return fn()
}

func withRestarter(fn func(*restarter.Restarter) error) error {
	restarter := restarter.New()
	defer restarter.Restart()

	return fn(restarter)
}

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
