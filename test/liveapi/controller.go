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

package liveapi

import (
	"net/http"
	"sync"

	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
	"github.com/ProtonMail/proton-bridge/v2/test/context/calls"
	"github.com/sirupsen/logrus"
)

// Controller implements PMAPIController interface for specified endpoint.
type Controller struct {
	log *logrus.Entry
	// Internal states.
	lock                 *sync.RWMutex
	calls                calls.Calls
	messageIDsByUsername map[string][]string

	// State controlled by test.
	noInternetConnection bool

	lastEventByUsername map[string]string
}

func NewController() (*Controller, pmapi.Manager) {
	controller := &Controller{
		log:                  logrus.WithField("pkg", "live-controller"),
		lock:                 &sync.RWMutex{},
		calls:                calls.Calls{},
		messageIDsByUsername: map[string][]string{},

		noInternetConnection: false,
		lastEventByUsername:  map[string]string{},
	}

	persistentClients.manager.SetTransport(&fakeTransport{
		ctl:       controller,
		transport: http.DefaultTransport,
	})

	return controller, persistentClients.manager
}
