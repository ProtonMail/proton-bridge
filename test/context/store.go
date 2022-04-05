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

package context

import (
	"errors"
)

// PairMessageID sets pairing between BDD message ID and API message ID.
func (ctx *TestContext) PairMessageID(username, bddMessageID, realMessageID string) {
	if bddMessageID == "" {
		return
	}
	ctx.bddMessageIDsToAPIIDs[username+":"+bddMessageID] = realMessageID
}

// GetAPIMessageID returns API message ID for given BDD message ID.
func (ctx *TestContext) GetAPIMessageID(username, bddMessageID string) (string, error) {
	msgID, ok := ctx.bddMessageIDsToAPIIDs[username+":"+bddMessageID]
	if !ok {
		return "", errors.New("unknown bddMessageID")
	}
	return msgID, nil
}
