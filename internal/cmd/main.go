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

package cmd

import (
	"os"
	"runtime"

	"github.com/ProtonMail/proton-bridge/pkg/constants"
	"github.com/getsentry/raven-go"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var (
	log = logrus.WithField("pkg", "cmd") //nolint[gochecknoglobals]

	baseFlags = []cli.Flag{ //nolint[gochecknoglobals]
		cli.StringFlag{
			Name:  "log-level, l",
			Usage: "Set the log level (one of panic, fatal, error, warn, info, debug, debug-client, debug-server)"},
		cli.BoolFlag{
			Name:  "cli, c",
			Usage: "Use command line interface"},
		cli.StringFlag{
			Name:  "version-json, g",
			Usage: "Generate json version file"},
		cli.BoolFlag{
			Name:  "mem-prof, m",
			Usage: "Generate memory profile"},
		cli.BoolFlag{
			Name:  "cpu-prof, p",
			Usage: "Generate CPU profile"},
	}
)

// Main sets up Sentry, filters out unwanted args, creates app and runs it.
func Main(appName, usage string, extraFlags []cli.Flag, run func(*cli.Context) error) {
	if err := raven.SetDSN(constants.DSNSentry); err != nil {
		log.WithError(err).Errorln("Can not setup sentry DSN")
	}
	raven.SetRelease(constants.Revision)

	filterProcessSerialNumberFromArgs()
	filterRestartNumberFromArgs()

	app := newApp(appName, usage, extraFlags, run)

	logrus.SetLevel(logrus.InfoLevel)
	log.WithField("version", constants.Version).
		WithField("revision", constants.Revision).
		WithField("build", constants.BuildTime).
		WithField("runtime", runtime.GOOS).
		WithField("args", os.Args).
		WithField("appName", app.Name).
		Info("Run app")

	if err := app.Run(os.Args); err != nil {
		log.Error("Program exited with error: ", err)
	}
}

func newApp(appName, usage string, extraFlags []cli.Flag, run func(*cli.Context) error) *cli.App {
	app := cli.NewApp()
	app.Name = appName
	app.Usage = usage
	app.Version = constants.BuildVersion
	app.Flags = append(baseFlags, extraFlags...) //nolint[gocritic]
	app.Action = run
	return app
}
