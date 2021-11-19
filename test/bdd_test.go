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
	"context"
	"os"

	testContext "github.com/ProtonMail/proton-bridge/test/context"
	"github.com/cucumber/godog"
)

const (
	timeFormat = "2006-01-02T15:04:05"
)

func SuiteInitializer(s *godog.TestSuiteContext) {
	s.BeforeSuite(testContext.BeforeRun)
	s.AfterSuite(testContext.AfterRun)
}

func ScenarioInitializer(s *godog.ScenarioContext) {
	s.Before(beforeScenario)
	s.After(afterScenario)

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

var ctx *testContext.TestContext //nolint[gochecknoglobals]

func beforeScenario(scenarioCtx context.Context, _ *godog.Scenario) (context.Context, error) {
	// NOTE(GODT-219) It would be possible to optimised the usage of godog with our context.
	app := os.Getenv("TEST_APP")
	ctx = testContext.New(app)
	return scenarioCtx, nil
}

func afterScenario(scenarioCtx context.Context, _ *godog.Scenario, err error) (context.Context, error) {
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

	return scenarioCtx, err
}
