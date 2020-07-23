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
	"sync"
	"time"

	"github.com/google/uuid"
)

// Request provides a result when it becomes available.
type Request struct {
	uuid    string
	ch      chan bool
	timeout time.Duration

	expired bool
	locker  sync.Locker
}

func newRequest(timeout time.Duration) *Request {
	return &Request{
		uuid:    uuid.New().String(),
		ch:      make(chan bool),
		timeout: timeout,
		locker:  &sync.Mutex{},
	}
}

// ID returns the request's ID, used to set the request's value.
func (r *Request) ID() string {
	return r.uuid
}

// Result returns the result or an error if it is not available within the request timeout.
func (r *Request) Result() (bool, error) {
	if r.hasExpired() {
		return false, errors.New("this result has expired")
	}

	defer r.done()

	select {
	case res := <-r.ch:
		return res, nil

	case <-time.After(r.timeout):
		return false, errors.New("timed out waiting for result")
	}
}

func (r *Request) hasExpired() bool {
	r.locker.Lock()
	defer r.locker.Unlock()

	return r.expired
}

func (r *Request) done() {
	r.locker.Lock()
	defer r.locker.Unlock()

	r.expired = true
}
