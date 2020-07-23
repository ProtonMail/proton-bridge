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

package confirmer

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Request struct {
	uuid    string
	value   chan bool
	timeout time.Duration
}

func newRequest(timeout time.Duration) *Request {
	return &Request{
		uuid:    uuid.New().String(),
		value:   make(chan bool),
		timeout: timeout,
	}
}

func (r *Request) ID() string {
	return r.uuid
}

func (r *Request) Result() (bool, error) {
	select {
	case res := <-r.value:
		return res, nil

	case <-time.After(r.timeout):
		return false, errors.New("timed out waiting for result")
	}
}
