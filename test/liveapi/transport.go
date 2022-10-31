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
	"io"
	"net/http"

	"github.com/pkg/errors"
)

func (ctl *Controller) TurnInternetConnectionOff() {
	ctl.noInternetConnection = true
}

func (ctl *Controller) TurnInternetConnectionOn() {
	ctl.noInternetConnection = false
}

type fakeTransport struct {
	ctl       *Controller
	transport http.RoundTripper
}

func (t *fakeTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.ctl.noInternetConnection {
		return nil, errors.New("no route to host")
	}

	body := []byte{}
	if req.GetBody != nil {
		bodyReader, err := req.GetBody()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get body")
		}
		if bodyReader != nil {
			body, err = io.ReadAll(bodyReader)
			if err != nil {
				return nil, errors.Wrap(err, "failed to read body")
			}
		}
	}
	t.ctl.recordCall(req.Method, req.URL.Path, body)

	return t.transport.RoundTrip(req)
}
