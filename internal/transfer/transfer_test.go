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

package transfer

import (
	"io/ioutil"
	"testing"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	transfermocks "github.com/ProtonMail/proton-bridge/internal/transfer/mocks"
	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
	pmapimocks "github.com/ProtonMail/proton-bridge/pkg/pmapi/mocks"
	gomock "github.com/golang/mock/gomock"
)

type mocks struct {
	t *testing.T

	ctrl               *gomock.Controller
	panicHandler       *transfermocks.MockPanicHandler
	clientManager      *transfermocks.MockClientManager
	imapClientProvider *transfermocks.MockIMAPClientProvider
	pmapiClient        *pmapimocks.MockClient
	pmapiConfig        *pmapi.ClientConfig

	keyring *crypto.KeyRing
}

func initMocks(t *testing.T) mocks {
	mockCtrl := gomock.NewController(t)

	m := mocks{
		t: t,

		ctrl:               mockCtrl,
		panicHandler:       transfermocks.NewMockPanicHandler(mockCtrl),
		clientManager:      transfermocks.NewMockClientManager(mockCtrl),
		imapClientProvider: transfermocks.NewMockIMAPClientProvider(mockCtrl),
		pmapiClient:        pmapimocks.NewMockClient(mockCtrl),
		pmapiConfig:        &pmapi.ClientConfig{},
		keyring:            newTestKeyring(),
	}

	m.clientManager.EXPECT().GetClient("user").Return(m.pmapiClient).AnyTimes()

	return m
}

func newTestKeyring() *crypto.KeyRing {
	data, err := ioutil.ReadFile("testdata/keyring_userKey")
	if err != nil {
		panic(err)
	}
	key, err := crypto.NewKeyFromArmored(string(data))
	if err != nil {
		panic(err)
	}
	userKey, err := crypto.NewKeyRing(key)
	if err != nil {
		panic(err)
	}
	return userKey
}
