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

package pmapi

import (
	"io"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
)

// Client defines the interface of a PMAPI client.
type Client interface {
	Auth(username, password string, info *AuthInfo) (*Auth, error)
	AuthInfo(username string) (*AuthInfo, error)
	AuthRefresh(token string) (*Auth, error)
	Auth2FA(twoFactorCode string, auth *Auth) (*Auth2FA, error)
	AuthSalt() (salt string, err error)
	Logout()
	DeleteAuth() error
	IsConnected() bool
	ClearData()

	CurrentUser() (*User, error)
	UpdateUser() (*User, error)
	Unlock(passphrase []byte) (err error)
	ReloadKeys(passphrase []byte) (err error)
	IsUnlocked() bool

	GetAddresses() (addresses AddressList, err error)
	Addresses() AddressList
	ReorderAddresses(addressIDs []string) error

	GetEvent(eventID string) (*Event, error)

	SendMessage(string, *SendMessageReq) (sent, parent *Message, err error)
	CreateDraft(m *Message, parent string, action int) (created *Message, err error)
	Import([]*ImportMsgReq) ([]*ImportMsgRes, error)

	CountMessages(addressID string) ([]*MessagesCount, error)
	ListMessages(filter *MessagesFilter) ([]*Message, int, error)
	GetMessage(apiID string) (*Message, error)
	DeleteMessages(apiIDs []string) error
	LabelMessages(apiIDs []string, labelID string) error
	UnlabelMessages(apiIDs []string, labelID string) error
	MarkMessagesRead(apiIDs []string) error
	MarkMessagesUnread(apiIDs []string) error

	ListLabels() ([]*Label, error)
	CreateLabel(label *Label) (*Label, error)
	UpdateLabel(label *Label) (*Label, error)
	DeleteLabel(labelID string) error
	EmptyFolder(labelID string, addressID string) error

	ReportBugWithEmailClient(os, osVersion, title, description, username, email, emailClient string) error
	SendSimpleMetric(category, action, label string) error

	GetMailSettings() (MailSettings, error)
	GetContactEmailByEmail(string, int, int) ([]ContactEmail, error)
	GetContactByID(string) (Contact, error)
	DecryptAndVerifyCards([]Card) ([]Card, error)

	GetAttachment(id string) (att io.ReadCloser, err error)
	CreateAttachment(att *Attachment, r io.Reader, sig io.Reader) (created *Attachment, err error)
	DeleteAttachment(attID string) (err error)

	KeyRingForAddressID(string) (kr *crypto.KeyRing, err error)
	GetPublicKeysForEmail(string) ([]PublicKey, bool, error)
}
