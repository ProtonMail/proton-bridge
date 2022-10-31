// Copyright (c) 2022 Proton AG
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
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.

package pmapi

import (
	"context"
	"io"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/go-resty/resty/v2"
)

// Client defines the interface of a PMAPI client.
type Client interface {
	Auth2FA(context.Context, string) error
	AuthSalt(ctx context.Context) (string, error)
	AuthDelete(context.Context) error
	AddAuthRefreshHandler(AuthRefreshHandler)

	GetUser(ctx context.Context) (*User, error)
	CurrentUser(ctx context.Context) (*User, error)
	UpdateUser(ctx context.Context) (*User, error)
	Unlock(ctx context.Context, passphrase []byte) (err error)
	ReloadKeys(ctx context.Context, passphrase []byte) (err error)
	IsUnlocked() bool

	Addresses() AddressList
	GetAddresses(context.Context) (addresses AddressList, err error)
	ReorderAddresses(ctx context.Context, addressIDs []string) error

	GetEvent(ctx context.Context, eventID string) (*Event, error)

	SendMessage(context.Context, string, *SendMessageReq) (sent, parent *Message, err error)
	CreateDraft(ctx context.Context, m *Message, parent string, action int) (created *Message, err error)
	Import(context.Context, ImportMsgReqs) ([]*ImportMsgRes, error)

	CountMessages(ctx context.Context, addressID string) ([]*MessagesCount, error)
	ListMessages(ctx context.Context, filter *MessagesFilter) ([]*Message, int, error)
	GetMessage(ctx context.Context, apiID string) (*Message, error)
	DeleteMessages(ctx context.Context, apiIDs []string) error
	LabelMessages(ctx context.Context, apiIDs []string, labelID string) error
	UnlabelMessages(ctx context.Context, apiIDs []string, labelID string) error
	MarkMessagesRead(ctx context.Context, apiIDs []string) error
	MarkMessagesUnread(ctx context.Context, apiIDs []string) error

	ListLabels(ctx context.Context) ([]*Label, error)
	CreateLabel(ctx context.Context, label *Label) (*Label, error)
	UpdateLabel(ctx context.Context, label *Label) (*Label, error)
	DeleteLabel(ctx context.Context, labelID string) error
	EmptyFolder(ctx context.Context, labelID string, addressID string) error

	// /core/V4/labels routes
	ListLabelsOnly(ctx context.Context) ([]*Label, error)
	ListFoldersOnly(ctx context.Context) ([]*Label, error)
	CreateLabelV4(ctx context.Context, label *Label) (*Label, error)
	UpdateLabelV4(ctx context.Context, label *Label) (*Label, error)
	DeleteLabelV4(ctx context.Context, labelID string) error

	GetMailSettings(ctx context.Context) (MailSettings, error)
	GetContactEmailByEmail(context.Context, string, int, int) ([]ContactEmail, error)
	GetContactByID(context.Context, string) (Contact, error)
	DecryptAndVerifyCards([]Card) ([]Card, error)

	GetAttachment(ctx context.Context, id string) (att io.ReadCloser, err error)
	CreateAttachment(ctx context.Context, att *Attachment, r io.Reader, sig io.Reader) (created *Attachment, err error)

	GetUserKeyRing() (*crypto.KeyRing, error)
	KeyRingForAddressID(string) (kr *crypto.KeyRing, err error)
	GetPublicKeysForEmail(context.Context, string) ([]PublicKey, bool, error)
}

type AuthRefreshHandler func(*AuthRefresh)

type clientManager interface {
	r(context.Context) *resty.Request
	authRefresh(context.Context, string, string) (*AuthRefresh, error)
	setSentryUserID(userID string)
}
