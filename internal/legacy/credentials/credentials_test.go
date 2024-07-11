// Copyright (c) 2024 Proton AG
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

package credentials

import (
	"encoding/base64"
	"fmt"
	"strings"
	"testing"
	"time"

	r "github.com/stretchr/testify/require"
)

var wantCredentials = Credentials{
	UserID:                "1",
	Name:                  "name",
	Emails:                "email1;email2",
	APIToken:              "token",
	MailboxPassword:       []byte("mailbox pass"),
	BridgePassword:        "bridge pass",
	Version:               "k11",
	Timestamp:             time.Now().Unix(),
	IsHidden:              false,
	IsCombinedAddressMode: false,
}

func TestUnmarshallBridge(t *testing.T) {
	encoded := wantCredentials.Marshal()
	haveCredentials := Credentials{UserID: "1"}
	r.NoError(t, haveCredentials.Unmarshal(encoded))
	r.Equal(t, wantCredentials, haveCredentials)
}

func TestUnmarshallImportExport(t *testing.T) {
	items := []string{
		wantCredentials.Name,
		wantCredentials.Emails,
		wantCredentials.APIToken,
		string(wantCredentials.MailboxPassword),
		"k11",
		fmt.Sprint(wantCredentials.Timestamp),
	}

	str := strings.Join(items, sep)
	encoded := base64.StdEncoding.EncodeToString([]byte(str))

	haveCredentials := Credentials{UserID: "1"}
	haveCredentials.BridgePassword = wantCredentials.BridgePassword // This one is not used.
	r.NoError(t, haveCredentials.Unmarshal(encoded))
	r.Equal(t, wantCredentials, haveCredentials)
}

func FuzzUnmarshal(f *testing.F) {
	items := []string{
		wantCredentials.Name,
		wantCredentials.Emails,
		wantCredentials.APIToken,
		string(wantCredentials.MailboxPassword),
		"k11",
		fmt.Sprint(wantCredentials.Timestamp),
	}
	str := strings.Join(items, sep)
	f.Add([]byte(str))

	f.Fuzz(func(_ *testing.T, secret []byte) {
		encodedSecret := base64.StdEncoding.EncodeToString(secret)

		creds := &Credentials{}
		_ = creds.Unmarshal(encodedSecret)
	})
}
