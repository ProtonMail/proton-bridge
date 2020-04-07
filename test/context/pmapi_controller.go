// Copyright (c) 2020 Proton Technologies AG
//
// This file is part of ProtonMail Bridge.Bridge.
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

package context

import (
	"os"

	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
	"github.com/ProtonMail/proton-bridge/test/fakeapi"
	"github.com/ProtonMail/proton-bridge/test/liveapi"
)

type PMAPIController interface {
	TurnInternetConnectionOff()
	TurnInternetConnectionOn()
	AddUser(user *pmapi.User, addresses *pmapi.AddressList, password string, twoFAEnabled bool) error
	AddUserLabel(username string, label *pmapi.Label) error
	GetLabelIDs(username string, labelNames []string) ([]string, error)
	AddUserMessage(username string, message *pmapi.Message) error
	GetMessageID(username, messageIndex string) string
	PrintCalls()
	WasCalled(method, path string, expectedRequest []byte) bool
	GetCalls(method, path string) [][]byte
}

func newPMAPIController(cm *pmapi.ClientManager) PMAPIController {
	switch os.Getenv(EnvName) {
	case EnvFake:
		return newFakePMAPIController(cm)
	case EnvLive:
		return newLivePMAPIController(cm)
	default:
		panic("unknown env")
	}
}

func newFakePMAPIController(cm *pmapi.ClientManager) PMAPIController {
	return newFakePMAPIControllerWrap(fakeapi.NewController(cm))
}

type fakePMAPIControllerWrap struct {
	*fakeapi.Controller
}

func newFakePMAPIControllerWrap(controller *fakeapi.Controller) PMAPIController {
	return &fakePMAPIControllerWrap{Controller: controller}
}

func newLivePMAPIController(cm *pmapi.ClientManager) PMAPIController {
	return newLiveAPIControllerWrap(liveapi.NewController(cm))
}

type liveAPIControllerWrap struct {
	*liveapi.Controller
}

func newLiveAPIControllerWrap(controller *liveapi.Controller) PMAPIController {
	return &liveAPIControllerWrap{Controller: controller}
}
