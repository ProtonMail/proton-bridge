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

package fakeapi

import (
	"sync"

	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
	"github.com/ProtonMail/proton-bridge/v2/test/context/calls"
	"github.com/sirupsen/logrus"
)

// Controller implements dummy PMAPIController interface without actual
// endpoint.
type Controller struct {
	// Internal states.
	lock               *sync.RWMutex
	fakeAPIs           []*FakePMAPI
	calls              calls.Calls
	labelIDGenerator   idGenerator
	messageIDGenerator idGenerator
	tokenGenerator     idGenerator
	clientManager      *fakePMAPIManager

	// State controlled by test.
	noInternetConnection bool
	usersByUsername      map[string]*fakeUser
	sessionsByUID        map[string]*fakeSession
	addressesByUsername  map[string]*pmapi.AddressList
	labelsByUsername     map[string][]*pmapi.Label
	messagesByUsername   map[string][]*pmapi.Message

	locker sync.Locker
	log    *logrus.Entry
}

func NewController() (*Controller, pmapi.Manager) {
	controller := &Controller{
		lock:               &sync.RWMutex{},
		fakeAPIs:           []*FakePMAPI{},
		calls:              calls.Calls{},
		labelIDGenerator:   100, // We cannot use system label IDs.
		messageIDGenerator: 0,
		tokenGenerator:     1000, // No specific reason; 1000 simply feels right.

		noInternetConnection: false,
		usersByUsername:      map[string]*fakeUser{},
		sessionsByUID:        map[string]*fakeSession{},
		addressesByUsername:  map[string]*pmapi.AddressList{},
		labelsByUsername:     map[string][]*pmapi.Label{},
		messagesByUsername:   map[string][]*pmapi.Message{},

		locker: &sync.Mutex{},
		log:    logrus.WithField("pkg", "fakeapi-controller"),
	}

	cm := &fakePMAPIManager{
		controller: controller,
	}

	controller.clientManager = cm

	return controller, cm
}
