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
	"os"

	"github.com/ProtonMail/proton-bridge/v2/pkg/sum"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)
	if err := createApp().Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}

func createApp() *cli.App { //nolint:funlen
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
	}

	return app
}

func computeSum(c *cli.Context) error {
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
