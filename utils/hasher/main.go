// Copyright (c) 2023 Proton AG
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
	"os"

	"github.com/ProtonMail/proton-bridge/v3/internal/updater"
	"github.com/ProtonMail/proton-bridge/v3/internal/versioner"
	"github.com/ProtonMail/proton-bridge/v3/pkg/sum"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)
	if err := createApp().Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}

func createApp() *cli.App {
	app := cli.NewApp()

	app.Name = "hasher"
	app.Usage = "Generate the recursive hash of a directory"
	app.Action = computeSum
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:     "root",
			Aliases:  []string{"C"},
			Usage:    "The root directory from which to begin recursive hashing",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "output",
			Aliases:  []string{"o"},
			Usage:    "The file to save the sum in",
			Required: true,
		},
		&cli.BoolFlag{
			Name:    "verify",
			Aliases: []string{"v"},
			Usage:   "Verify the update folder is properly hashed and signed.",
		},
	}

	return app
}

func computeSum(c *cli.Context) error {
	if c.Bool("verify") {
		kr, err := updater.GetDefaultKeyring()
		if err != nil {
			logrus.WithError(err).Fatal("Failed to load key before verify")
		}

		if err := versioner.VerifyUpdateFolder(kr, c.String("root")); err != nil {
			logrus.WithError(err).Fatal("Failed to verify")
		}

		logrus.WithField("path", c.String("root")).Info("Signature OK")
	}

	b, err := sum.RecursiveSum(c.String("root"), c.String("output"))
	if err != nil {
		return err
	}

	f, err := os.Create(c.String("output"))
	if err != nil {
		return err
	}

	if _, err := f.Write(b); err != nil {
		return err
	}

	return f.Close()
}
