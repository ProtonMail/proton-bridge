package app

import (
	"fmt"

	"github.com/ProtonMail/proton-bridge/v2/internal/crash"
	"github.com/ProtonMail/proton-bridge/v2/internal/locations"
	"github.com/ProtonMail/proton-bridge/v2/internal/logging"
	"github.com/urfave/cli/v2"
)

func initLogging(c *cli.Context, locations *locations.Locations, crashHandler *crash.Handler) error {
	// Get a place to keep our logs.
	logsPath, err := locations.ProvideLogsPath()
	if err != nil {
		return fmt.Errorf("could not provide logs path: %w", err)
	}

	// Initialize logging.
	if err := logging.Init(logsPath, c.String(flagLogLevel)); err != nil {
		return fmt.Errorf("could not initialize logging: %w", err)
	}

	// Ensure we dump a stack trace if we crash.
	crashHandler.AddRecoveryAction(logging.DumpStackTrace(logsPath))

	return nil
}
