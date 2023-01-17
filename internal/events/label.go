// Copyright (c) 2023 Proton AG
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
// along with Proton Mail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package events

import (
	"fmt"

	"github.com/ProtonMail/proton-bridge/v3/internal/logging"
)

type UserLabelCreated struct {
	eventBase

	UserID  string
	LabelID string
	Name    string
}

func (event UserLabelCreated) String() string {
	return fmt.Sprintf("UserLabelCreated: UserID: %s, LabelID: %s, Name: %s", event.UserID, event.LabelID, logging.Sensitive(event.Name))
}

type UserLabelUpdated struct {
	eventBase

	UserID  string
	LabelID string
	Name    string
}

func (event UserLabelUpdated) String() string {
	return fmt.Sprintf("UserLabelUpdated: UserID: %s, LabelID: %s, Name: %s", event.UserID, event.LabelID, logging.Sensitive(event.Name))
}

type UserLabelDeleted struct {
	eventBase

	UserID  string
	LabelID string
}

func (event UserLabelDeleted) String() string {
	return fmt.Sprintf("UserLabelDeleted: UserID: %s, LabelID: %s", event.UserID, event.LabelID)
}
