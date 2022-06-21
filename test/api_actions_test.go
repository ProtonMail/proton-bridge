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
)

func APIActionsFeatureContext(s *godog.ScenarioContext) {
	s.Step(`^the internet connection is lost$`, theInternetConnectionIsLost)
	s.Step(`^the internet connection is restored$`, theInternetConnectionIsRestored)
	s.Step(`^(\d+) second[s]? pass$`, secondsPass)
	s.Step(`^the body of draft "([^"]*)" for "([^"]*)" has changed to "([^"]*)"$`, draftBodyChanged)
}

func theInternetConnectionIsLost() error {
	ctx.GetPMAPIController().TurnInternetConnectionOff()
	return nil
}

func theInternetConnectionIsRestored() error {
	ctx.GetPMAPIController().TurnInternetConnectionOn()
	return nil
}

func secondsPass(seconds int) error {
	time.Sleep(time.Duration(seconds) * time.Second)
	return nil
}

func draftBodyChanged(bddMessageID, bddUserID, body string) error {
	account := ctx.GetTestAccount(bddUserID)
	if account == nil {
		return godog.ErrPending
	}

	messageID, err := ctx.GetAPIMessageID(account.Username(), bddMessageID)
	if err != nil {
		return internalError(err, "getting apiID for %s", bddMessageID)
	}

	err = ctx.GetPMAPIController().SetDraftBody(account.Username(), messageID, body)
	if err != nil {
		return internalError(err, "cannot set body of %s", messageID)
	}

	return nil
}
