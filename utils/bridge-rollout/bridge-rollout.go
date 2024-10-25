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
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.

package main

import (
	"fmt"
	"os"

	"github.com/ProtonMail/gluon/async"
	"github.com/ProtonMail/proton-bridge/v3/internal/app"
	"github.com/ProtonMail/proton-bridge/v3/internal/locations"
	"github.com/ProtonMail/proton-bridge/v3/internal/vault"
	"github.com/ProtonMail/proton-bridge/v3/pkg/keychain"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func main() {
	logrus.SetLevel(logrus.ErrorLevel)
	app := cli.NewApp()

	app.Commands = []*cli.Command{
		{
			Name:   "get",
			Action: getRollout,
			Usage:  "get the bridge rollout value",
		},
		{
			Name:   "set",
			Action: setRollout,
			Flags: []cli.Flag{
				&cli.Float64Flag{
					Name:     "value",
					Usage:    "the rollout value",
					Required: true,
					Aliases:  []string{"v"},
				},
			},
			Usage: "set the bridge rollout value",
		},
	}
	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}

func getRollout(_ *cli.Context) error {
	return app.WithLocations(func(locations *locations.Locations) error {
		return app.WithKeychainList(async.NoopPanicHandler{}, false, func(keychains *keychain.List) error {
			return app.WithVault(nil, locations, keychains, async.NoopPanicHandler{}, func(vault *vault.Vault, _, _ bool) error {
				fmt.Println(vault.GetUpdateRollout())
				return nil
			})
		})
	})
}

func setRollout(c *cli.Context) error {
	return app.WithLocations(func(locations *locations.Locations) error {
		return app.WithKeychainList(async.NoopPanicHandler{}, false, func(keychains *keychain.List) error {
			return app.WithVault(nil, locations, keychains, async.NoopPanicHandler{}, func(vault *vault.Vault, _, _ bool) error {
				clamped := max(0.0, min(1.0, c.Float64("value")))
				if err := vault.SetUpdateRollout(clamped); err != nil {
					return err
				}
				return getRollout(c)
			})
		})
	})
}
