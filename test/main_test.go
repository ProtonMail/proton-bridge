// Copyright (c) 2022 Proton AG
//
// This file is part of Proton Mail Bridge.Bridge.
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

package tests

import (
	"flag"
	"os"
	"testing"

	"github.com/ProtonMail/proton-bridge/v2/internal/constants"
	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
)

var opt = godog.Options{ //nolint:gochecknoglobals
	Output: colors.Colored(os.Stdout),
	Format: "progress", // can define default values
}

func init() { //nolint:gochecknoinits
	godog.BindCommandLineFlags("godog.", &opt)

	// This would normally be done using ldflags but `godog` command doesn't support that.
	constants.Version = os.Getenv("BRIDGE_VERSION")
}

func TestMain(m *testing.M) {
	flag.Parse()
	opt.Paths = flag.Args()

	status := godog.TestSuite{
		Name:                 "bridge-integration-tests",
		TestSuiteInitializer: SuiteInitializer,
		ScenarioInitializer:  ScenarioInitializer,
		Options:              &opt,
	}.Run()

	if st := m.Run(); st > status {
		status = st
	}

	os.Exit(status)
}
