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

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/Masterminds/semver/v3"
	"github.com/ProtonMail/proton-bridge/v2/internal/updater"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

type versionInfo struct {
	updater.VersionInfo

	Commit string
}

func main() {
	if err := createApp().Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}

func createApp() *cli.App { //nolint:funlen
	app := cli.NewApp()

	app.Name = "versioner"

	app.Usage = "Create and update version files"

	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:     "app",
			Usage:    "The app (bridge, importExport)",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "platform",
			Usage:    "The platform (windows, darwin, linux)",
			Required: true,
		},
	}

	app.Commands = []*cli.Command{{
		Name:   "update",
		Action: update,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "channel",
				Usage:    "The update channel (live/beta/...)",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "version",
				Usage: "The version of the app",
			},
			&cli.StringFlag{
				Name:  "min-auto",
				Usage: "The minimum version of the app that can autoupdate to this version",
			},
			&cli.StringFlag{
				Name:  "package",
				Usage: "The package file",
			},
			&cli.StringSliceFlag{
				Name:  "installer",
				Usage: "An installer that can be used to manually install the app (can be specified multiple times)",
			},
			&cli.StringFlag{
				Name:  "landing-page",
				Usage: "The landing page",
			},
			&cli.StringFlag{
				Name:  "release-notes-page",
				Usage: "The release notes page",
			},
			&cli.Float64Flag{
				Name:  "rollout",
				Usage: "What proportion of users should receive this update",
			},
			&cli.StringFlag{
				Name:  "commit",
				Usage: "What commit produced this update",
			},
		},
	}, {
		Name:   "dump",
		Action: dump,
	}}

	return app
}

func update(c *cli.Context) error {
	versions := fetch(c.String("app"), c.String("platform"))

	version := versions[c.String("channel")]

	if c.IsSet("version") {
		version.Version = semver.MustParse(c.String("version"))
	}

	if c.IsSet("min-auto") {
		version.MinAuto = semver.MustParse(c.String("min-auto"))
	}

	if c.IsSet("package") {
		version.Package = c.String("package")
	}

	if c.IsSet("installer") {
		version.Installers = c.StringSlice("installer")
	}

	if c.IsSet("landing-page") {
		version.LandingPage = c.String("landing-page")
	}

	if c.IsSet("release-notes-page") {
		version.ReleaseNotesPage = c.String("release-notes-page")
	}

	if c.IsSet("rollout") {
		version.RolloutProportion = c.Float64("rollout")
	}

	if c.IsSet("commit") {
		version.Commit = c.String("commit")
	}

	versions[c.String("channel")] = version

	return write(c.App.Writer, versions)
}

func dump(c *cli.Context) error {
	return write(c.App.Writer, fetch(c.String("app"), c.String("platform")))
}

func fetch(app, platform string) map[string]versionInfo {
	url := fmt.Sprintf(
		"%v/%v/version_%v.json",
		updater.Host, app, platform,
	)

	res, err := resty.New().R().Get(url)
	if err != nil {
		logrus.WithError(err).Error("Fetch failed.")
		return make(map[string]versionInfo)
	}

	var versionMap map[string]versionInfo

	if err := json.Unmarshal(res.Body(), &versionMap); err != nil {
		logrus.WithError(err).Error("Unmarshal failed.")
		return make(map[string]versionInfo)
	}

	return versionMap
}

func write(w io.Writer, versions map[string]versionInfo) error {
	enc := json.NewEncoder(w)

	enc.SetIndent("", "  ")

	return enc.Encode(versions)
}
