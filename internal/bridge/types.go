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

package bridge

import (
	"io"

	pmcrypto "github.com/ProtonMail/gopenpgp/crypto"
	"github.com/ProtonMail/proton-bridge/internal/bridge/credentials"
	pmapi "github.com/ProtonMail/proton-bridge/pkg/pmapi" // mockgen needs this to be given an explicit import name
)

type Configer interface {
	ClearData() error
	GetDBDir() string
	GetIMAPCachePath() string
	GetAPIConfig() *pmapi.ClientConfig
}

type PreferenceProvider interface {
	Get(key string) string
	GetBool(key string) bool
	GetInt(key string) int
	Set(key string, value string)
}

type PanicHandler interface {
	HandlePanic()
}

type Clientman interface {
}

type PMAPIProvider interface {
	Auth(username, password string, info *pmapi.AuthInfo) (*pmapi.Auth, error)
	AuthInfo(username string) (*pmapi.AuthInfo, error)
	AuthRefresh(token string) (*pmapi.Auth, error)
	Unlock(mailboxPassword string) (kr *pmcrypto.KeyRing, err error)
	UnlockAddresses(passphrase []byte) error
	CurrentUser() (*pmapi.User, error)
	UpdateUser() (*pmapi.User, error)
	Addresses() pmapi.AddressList

	Logout()

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

	ListLabels() ([]*pmapi.Label, error)
	CreateLabel(label *pmapi.Label) (*pmapi.Label, error)
	UpdateLabel(label *pmapi.Label) (*pmapi.Label, error)
	DeleteLabel(labelID string) error
	EmptyFolder(labelID string, addressID string) error

	ReportBugWithEmailClient(os, osVersion, title, description, username, email, emailClient string) error
	SendSimpleMetric(category, action, label string) error

	Auth2FA(twoFactorCode string, auth *pmapi.Auth) (*pmapi.Auth2FA, error)

	GetMailSettings() (pmapi.MailSettings, error)
	GetContactEmailByEmail(string, int, int) ([]pmapi.ContactEmail, error)
	GetContactByID(string) (pmapi.Contact, error)
	DecryptAndVerifyCards([]pmapi.Card) ([]pmapi.Card, error)
	GetPublicKeysForEmail(string) ([]pmapi.PublicKey, bool, error)
	SendMessage(string, *pmapi.SendMessageReq) (sent, parent *pmapi.Message, err error)
	CreateDraft(m *pmapi.Message, parent string, action int) (created *pmapi.Message, err error)
	CreateAttachment(att *pmapi.Attachment, r io.Reader, sig io.Reader) (created *pmapi.Attachment, err error)
	KeyRingForAddressID(string) (kr *pmcrypto.KeyRing)

	GetAttachment(id string) (att io.ReadCloser, err error)
}

type CredentialsStorer interface {
	List() (userIDs []string, err error)
	Add(userID, userName, apiToken, mailboxPassword string, emails []string) (*credentials.Credentials, error)
	Get(userID string) (*credentials.Credentials, error)
	SwitchAddressMode(userID string) error
	UpdateEmails(userID string, emails []string) error
	UpdateToken(userID, apiToken string) error
	Logout(userID string) error
	Delete(userID string) error
}
