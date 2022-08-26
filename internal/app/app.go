package app

import (
	"fmt"
	"path/filepath"

	"github.com/ProtonMail/proton-bridge/v2/internal/constants"
	"github.com/ProtonMail/proton-bridge/v2/internal/crash"
	"github.com/ProtonMail/proton-bridge/v2/internal/focus"
	bridgeCLI "github.com/ProtonMail/proton-bridge/v2/internal/frontend/cli"
	"github.com/ProtonMail/proton-bridge/v2/internal/frontend/grpc"
	"github.com/ProtonMail/proton-bridge/v2/internal/locations"
	"github.com/ProtonMail/proton-bridge/v2/internal/sentry"
	"github.com/ProtonMail/proton-bridge/v2/internal/useragent"
	"github.com/ProtonMail/proton-bridge/v2/pkg/restarter"
	"github.com/pkg/profile"
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

	flagNoWindow       = "no-window"
	flagNonInteractive = "non-interactive"
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
			Name:   flagNoWindow,
			Usage:  "Don't show window after start",
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

	// Start CPU profile if requested.
	if c.Bool(flagCPUProfile) {
		p := profile.Start(profile.CPUProfile, profile.ProfilePath("cpu.pprof"))
		defer p.Stop()
	}

	// Start memory profile if requested.
	if c.Bool(flagMemProfile) {
		p := profile.Start(profile.MemProfile, profile.MemProfileAllocs, profile.ProfilePath("mem.pprof"))
		defer p.Stop()
	}

	// Create the restarter.
	restarter := restarter.New()
	defer restarter.Restart()

	// Create a user agent that will be used for all requests.
	identifier := useragent.New()

	// Create a crash handler that will send crash reports to sentry.
	crashHandler := crash.NewHandler(
		sentry.NewReporter(constants.FullAppName, constants.Version, identifier).ReportException,
		crash.ShowErrorNotification(constants.FullAppName),
		func(r interface{}) error { restarter.Set(true, true); return nil },
	)
	defer crashHandler.HandlePanic()

	// Create a locations provider to determine where to store our files.
	provider, err := locations.NewDefaultProvider(filepath.Join(constants.VendorName, constants.ConfigName))
	if err != nil {
		return fmt.Errorf("could not create locations provider: %w", err)
	}

	// Create a new locations object that will be used to provide paths to store files.
	locations := locations.New(provider, constants.ConfigName)

	// Initialize the logging.
	if err := initLogging(c, locations, crashHandler); err != nil {
		return fmt.Errorf("could not initialize logging: %w", err)
	}

	// Create the bridge.
	bridge, err := newBridge(locations, identifier)
	if err != nil {
		return fmt.Errorf("could not create bridge: %w", err)
	}
	defer bridge.Close(c.Context)

	// Start the frontend.
	switch {
	case c.Bool(flagCLI):
		return bridgeCLI.New(bridge).Loop()

	case c.Bool(flagNonInteractive):
		select {}

	default:
		service, err := grpc.NewService(crashHandler, restarter, locations, bridge, !c.Bool(flagNoWindow))
		if err != nil {
			return fmt.Errorf("could not create service: %w", err)
		}

		return service.Loop()
	}
}
