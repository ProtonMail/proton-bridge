// Copyright (c) 2021 Proton Technologies AG
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

	"github.com/ProtonMail/proton-bridge/internal/events"
	"github.com/ProtonMail/proton-bridge/pkg/listener"
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
	AddUserMessage(username string, message *pmapi.Message) (string, error)
	GetMessages(username, labelID string) ([]*pmapi.Message, error)
	ReorderAddresses(user *pmapi.User, addressIDs []string) error
	PrintCalls()
	WasCalled(method, path string, expectedRequest []byte) bool
	GetCalls(method, path string) [][]byte
}

func newPMAPIController(app string, listener listener.Listener) (PMAPIController, pmapi.Manager) {
	switch os.Getenv(EnvName) {
	case EnvFake:
		cntl, cm := fakeapi.NewController()
		addConnectionObserver(cm, listener)
		return cntl, cm

	case EnvLive:
		cntl, cm := liveapi.NewController(app)
		addConnectionObserver(cm, listener)
		return cntl, cm

	default:
		panic("unknown env")
	}
}

func addConnectionObserver(cm pmapi.Manager, listener listener.Listener) {
	cm.AddConnectionObserver(pmapi.NewConnectionObserver(
		func() { listener.Emit(events.InternetOffEvent, "") },
		func() { listener.Emit(events.InternetOnEvent, "") },
	))
}
