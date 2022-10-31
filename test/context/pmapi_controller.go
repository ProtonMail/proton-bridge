// Copyright (c) 2022 Proton AG
//
// This file is part of Proton Mail Bridge.Bridge.
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

package context

import (
	"os"

	"github.com/ProtonMail/proton-bridge/v2/internal/events"
	"github.com/ProtonMail/proton-bridge/v2/pkg/listener"
	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
	"github.com/ProtonMail/proton-bridge/v2/test/accounts"
	"github.com/ProtonMail/proton-bridge/v2/test/fakeapi"
	"github.com/ProtonMail/proton-bridge/v2/test/liveapi"
)

type PMAPIController interface {
	TurnInternetConnectionOff()
	TurnInternetConnectionOn()
	GetAuthClient(username string) pmapi.Client
	AddUser(account *accounts.TestAccount) error
	AddUserLabel(username string, label *pmapi.Label) error
	GetLabelIDs(username string, labelNames []string) ([]string, error)
	AddUserMessage(username string, message *pmapi.Message) (string, error)
	SetDraftBody(username string, messageID string, body string) error
	GetMessages(username, labelID string) ([]*pmapi.Message, error)
	ReorderAddresses(user *pmapi.User, addressIDs []string) error
	PrintCalls()
	WasCalled(method, path string, expectedRequest []byte) bool
	WasCalledRegex(methodRegex, pathRegex string, expectedRequest []byte) (bool, error)
	GetCalls(method, path string) [][]byte
	LockEvents(username string)
	UnlockEvents(username string)
	RemoveUserMessageWithoutEvent(username, messageID string) error
	RevokeSession(username string) error
}

func newPMAPIController(listener listener.Listener) (PMAPIController, pmapi.Manager) {
	switch os.Getenv(EnvName) {
	case EnvFake:
		cntl, cm := fakeapi.NewController()
		addConnectionObserver(cm, listener)
		return cntl, cm

	case EnvLive:
		cntl, cm := liveapi.NewController()
		addConnectionObserver(cm, listener)
		return cntl, cm

	default:
		panic("unknown env")
	}
}

func addConnectionObserver(cm pmapi.Manager, listener listener.Listener) {
	cm.AddConnectionObserver(pmapi.NewConnectionObserver(
		func() { listener.Emit(events.InternetConnChangedEvent, events.InternetOff) },
		func() { listener.Emit(events.InternetConnChangedEvent, events.InternetOn) },
	))
}
