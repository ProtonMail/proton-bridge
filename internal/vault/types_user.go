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

package vault

import "github.com/ProtonMail/gluon/imap"

// UserData holds information about a single bridge user.
// The user may or may not be logged in.
type UserData struct {
	UserID   string
	Username string

	GluonKey    []byte
	GluonIDs    map[string]string
	UIDValidity map[string]imap.UID
	BridgePass  []byte // raw token represented as byte slice (needs to be encoded)
	AddressMode AddressMode

	AuthUID string
	AuthRef string
	KeyPass []byte

	SyncStatus SyncStatus
	EventID    string
}

type AddressMode int

const (
	CombinedMode AddressMode = iota
	SplitMode
)

func (mode AddressMode) String() string {
	switch mode {
	case CombinedMode:
		return "combined"

	case SplitMode:
		return "split"

	default:
		return "unknown"
	}
}

type SyncStatus struct {
	HasLabels        bool
	HasMessages      bool
	LastMessageID    string
	FailedMessageIDs []string
}

func (status SyncStatus) IsComplete() bool {
	return status.HasLabels && status.HasMessages
}

func newDefaultUser(userID, username, authUID, authRef string, keyPass []byte) UserData {
	return UserData{
		UserID:   userID,
		Username: username,

		GluonKey:    newRandomToken(32),
		GluonIDs:    make(map[string]string),
		UIDValidity: make(map[string]imap.UID),
		BridgePass:  newRandomToken(16),
		AddressMode: CombinedMode,

		AuthUID: authUID,
		AuthRef: authRef,
		KeyPass: keyPass,
	}
}
