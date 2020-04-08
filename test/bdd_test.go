// Copyright (c) 2020 Proton Technologies AG
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
	"github.com/ProtonMail/proton-bridge/test/context"
	"github.com/cucumber/godog"
)

func FeatureContext(s *godog.Suite) {
	s.BeforeScenario(beforeScenario)
	s.AfterScenario(afterScenario)

	APIChecksFeatureContext(s)

	BridgeActionsFeatureContext(s)
	BridgeChecksFeatureContext(s)
	BridgeSetupFeatureContext(s)

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
}

var ctx *context.TestContext //nolint[gochecknoglobals]

func beforeScenario(scenario interface{}) {
	ctx = context.New()
}

func afterScenario(scenario interface{}, err error) {
	if err != nil {
		for _, user := range ctx.GetBridge().GetUsers() {
			user.GetStore().TestDumpDB(ctx.GetTestingT())
		}
	}
	ctx.Cleanup()
	if err != nil {
		ctx.GetPMAPIController().PrintCalls()
	}
}
