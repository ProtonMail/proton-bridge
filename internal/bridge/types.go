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
	"github.com/ProtonMail/proton-bridge/internal/bridge/credentials"
	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
)

type Configer interface {
	ClearData() error
	GetDBDir() string
	GetVersion() string
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

type CredentialsStorer interface {
	List() (userIDs []string, err error)
	Add(userID, userName, apiToken, mailboxPassword string, emails []string) (*credentials.Credentials, error)
	Get(userID string) (*credentials.Credentials, error)
	SwitchAddressMode(userID string) error
	UpdateEmails(userID string, emails []string) error
	UpdatePassword(userID, password string) error
	UpdateToken(userID, apiToken string) error
	Logout(userID string) error
	Delete(userID string) error
}

type ClientManager interface {
	GetClient(userID string) pmapi.Client
	GetAnonymousClient() pmapi.Client
	AllowProxy()
	DisallowProxy()
	GetAuthUpdateChannel() chan pmapi.ClientAuth
	CheckConnection() error
	SetUserAgent(clientName, clientVersion, os string)
}
