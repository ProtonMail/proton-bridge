// Copyright (c) 2021 Proton Technologies AG
//
// This file is part of ProtonMail Bridge.Bridge.
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

package tests

import (
	"os"

	"github.com/ProtonMail/proton-bridge/test/context"
	"github.com/cucumber/godog"
)

const (
	timeFormat = "2006-01-02T15:04:05"
)

func FeatureContext(s *godog.Suite) {
	s.BeforeScenario(beforeScenario)
	s.AfterScenario(afterScenario)

	APIActionsFeatureContext(s)
	APIChecksFeatureContext(s)
	APISetupFeatureContext(s)

	BridgeActionsFeatureContext(s)

	CommonChecksFeatureContext(s)

	IMAPActionsAuthFeatureContext(s)
	IMAPActionsMailboxFeatureContext(s)
	IMAPActionsMessagesFeatureContext(s)
	IMAPChecksFeatureContext(s)
	IMAPSetupFeatureContext(s)

	SMTPActionsAuthFeatureContext(s)
	SMTPChecksFeatureContext(s)
	SMTPSetupFeatureContext(s)

	StoreActionsFeatureContext(s)
	StoreChecksFeatureContext(s)
	StoreSetupFeatureContext(s)

	TransferActionsFeatureContext(s)
	TransferChecksFeatureContext(s)
	TransferSetupFeatureContext(s)

	UsersActionsFeatureContext(s)
	UsersSetupFeatureContext(s)
	UsersChecksFeatureContext(s)
}

var ctx *context.TestContext //nolint[gochecknoglobals]

func beforeScenario(scenario interface{}) {
	// bridge or ie. With godog 0.10.x and later it can be determined from
	// scenario.Uri and its file location.
	app := os.Getenv("TEST_APP")
	ctx = context.New(app)
}

func afterScenario(scenario interface{}, err error) {
	if err != nil {
		for _, user := range ctx.GetUsers().GetUsers() {
			store := user.GetStore()
			if store != nil {
				store.TestDumpDB(ctx.GetTestingT())
			}
		}
	}
	ctx.Cleanup()
	if err != nil {
		ctx.GetPMAPIController().PrintCalls()
	}
}
