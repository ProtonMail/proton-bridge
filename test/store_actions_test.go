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
	"time"

	"github.com/cucumber/godog"
	"github.com/stretchr/testify/assert"
)

func StoreActionsFeatureContext(s *godog.ScenarioContext) {
	s.Step(`^the event loop of "([^"]*)" loops once$`, theEventLoopLoops)
	s.Step(`^"([^"]*)" receives an address event$`, receivesAnAddressEvent)
}

func theEventLoopLoops(username string) error {
	acc := ctx.GetTestAccount(username)
	if acc == nil {
		return godog.ErrPending
	}
	store, err := ctx.GetStore(acc.Username())
	if err != nil {
		return internalError(err, "getting store of user %s", username)
	}
	store.TestPollNow()
	return nil
}

func receivesAnAddressEvent(username string) error {
	acc := ctx.GetTestAccount(username)
	if acc == nil {
		return godog.ErrPending
	}
	store, err := ctx.GetStore(acc.Username())
	if err != nil {
		return internalError(err, "getting store of user %s", username)
	}
	assert.Eventually(ctx.GetTestingT(), func() bool {
		store.TestPollNow()
		return len(store.TestGetLastEvent().Addresses) > 0
	}, 5*time.Second, time.Second)
	return nil
}
