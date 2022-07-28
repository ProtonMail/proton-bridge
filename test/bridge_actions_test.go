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
	"github.com/cucumber/godog"
)

func BridgeActionsFeatureContext(s *godog.ScenarioContext) {
	s.Step(`^bridge starts$`, bridgeStarts)
	s.Step(`^bridge syncs "([^"]*)"$`, bridgeSyncsUser)
	s.Step(`^All mail mailbox is hidden$`, allMailMailboxIsHidden)
	s.Step(`^All mail mailbox is visible$`, allMailMailboxIsVisible)
}

func bridgeStarts() error {
	ctx.SetLastError(ctx.RestartBridge())
	return nil
}

func bridgeSyncsUser(bddUserID string) error {
	account := ctx.GetTestAccount(bddUserID)
	if account == nil {
		return godog.ErrPending
	}
	if err := ctx.WaitForSync(account.Username()); err != nil {
		return internalError(err, "waiting for sync")
	}
	ctx.SetLastError(ctx.GetTestingError())
	return nil
}

func allMailMailboxIsHidden() error {
	ctx.GetBridge().SetIsAllMailVisible(false)
	return nil
}

func allMailMailboxIsVisible() error {
	ctx.GetBridge().SetIsAllMailVisible(true)
	return nil
}
