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

package liveapi

import (
	"net/http"
	"sync"

	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
)

type Controller struct {
	// Internal states.
	lock                 *sync.RWMutex
	calls                []*fakeCall
	pmapiByUsername      map[string]pmapi.Client
	messageIDsByUsername map[string][]string
	clientManager        pmapi.Manager

	// State controlled by test.
	noInternetConnection bool
}

func NewController() (*Controller, pmapi.Manager) {
	controller := &Controller{
		lock:                 &sync.RWMutex{},
		calls:                []*fakeCall{},
		pmapiByUsername:      map[string]pmapi.Client{},
		messageIDsByUsername: map[string][]string{},

		noInternetConnection: false,
	}

	// FIXME(conman): Set connect values here?
	cm := pmapi.New(pmapi.DefaultConfig)

	cm.SetTransport(&fakeTransport{
		ctl:       controller,
		transport: http.DefaultTransport,
	})

	return controller, cm
}
