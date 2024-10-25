//go:build debug

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
	"io"
	"os"

	"github.com/ProtonMail/gluon/async"
	"github.com/ProtonMail/proton-bridge/v3/internal/app"
	"github.com/ProtonMail/proton-bridge/v3/internal/locations"
	"github.com/ProtonMail/proton-bridge/v3/internal/vault"
	"github.com/ProtonMail/proton-bridge/v3/pkg/keychain"
	"github.com/urfave/cli/v2"
)

func main() {
	app := cli.NewApp()

	app.Commands = []*cli.Command{
		{
			Name:   "read",
			Action: readAction,
		},
		{
			Name:   "write",
			Action: writeAction,
		},
	}

	app.RunAndExitOnError()
}

func readAction(c *cli.Context) error {
	return app.WithLocations(func(locations *locations.Locations) error {
		return app.WithKeychainList(async.NoopPanicHandler{}, false, func(keychains *keychain.List) error {
			return app.WithVault(nil, locations, keychains, async.NoopPanicHandler{}, func(vault *vault.Vault, insecure, corrupt bool) error {
				if _, err := os.Stdout.Write(vault.ExportJSON()); err != nil {
					return fmt.Errorf("failed to write vault: %w", err)
				}

				return nil
			})
		})
	})
}

func writeAction(c *cli.Context) error {
	return app.WithLocations(func(locations *locations.Locations) error {
		return app.WithKeychainList(async.NoopPanicHandler{}, false, func(keychains *keychain.List) error {
			return app.WithVault(nil, locations, keychains, async.NoopPanicHandler{}, func(vault *vault.Vault, insecure, corrupt bool) error {
				b, err := io.ReadAll(os.Stdin)
				if err != nil {
					return fmt.Errorf("failed to read vault: %w", err)
				}

				vault.ImportJSON(b)

				return nil
			})
		})
	})
}
