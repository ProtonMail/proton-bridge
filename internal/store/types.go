// Copyright (c) 2020 Proton Technologies AG
//
// This file is part of ProtonMail Bridge.
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

package store

import (
	"io"

	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
)

type PanicHandler interface {
	HandlePanic()
}

// PMAPIProvider is subset of pmapi.Client for use by the Store.
type PMAPIProvider interface {
	CurrentUser() (*pmapi.User, error)
	Addresses() pmapi.AddressList

	GetEvent(eventID string) (*pmapi.Event, error)

	CountMessages(addressID string) ([]*pmapi.MessagesCount, error)
	ListMessages(filter *pmapi.MessagesFilter) ([]*pmapi.Message, int, error)
	GetMessage(apiID string) (*pmapi.Message, error)
	Import([]*pmapi.ImportMsgReq) ([]*pmapi.ImportMsgRes, error)
	DeleteMessages(apiIDs []string) error
	LabelMessages(apiIDs []string, labelID string) error
	UnlabelMessages(apiIDs []string, labelID string) error
	MarkMessagesRead(apiIDs []string) error
	MarkMessagesUnread(apiIDs []string) error

	CreateDraft(m *pmapi.Message, parent string, action int) (created *pmapi.Message, err error)
	CreateAttachment(att *pmapi.Attachment, r io.Reader, sig io.Reader) (created *pmapi.Attachment, err error)
	SendMessage(messageID string, req *pmapi.SendMessageReq) (sent, parent *pmapi.Message, err error)

	ListLabels() ([]*pmapi.Label, error)
	CreateLabel(label *pmapi.Label) (*pmapi.Label, error)
	UpdateLabel(label *pmapi.Label) (*pmapi.Label, error)
	DeleteLabel(labelID string) error
	EmptyFolder(labelID string, addressID string) error
}

// BridgeUser is subset of bridge.User for use by the Store.
type BridgeUser interface {
	ID() string
	GetAddressID(address string) (string, error)
	IsConnected() bool
	IsCombinedAddressMode() bool
	GetPrimaryAddress() string
	GetStoreAddresses() []string
	UpdateUser() error
	CloseConnection(string)
	Logout() error
}
