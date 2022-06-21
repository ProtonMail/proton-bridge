// Copyright (c) 2022 Proton AG
//
// This file is part of Proton Mail Bridge.
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

package confirmer

import (
	"errors"
	"sync"
	"time"
)

// NOTE: For now, Confirmer only supports bool values but it could easily be made generic.

// Confirmer is used to ask for some value (e.g. a confirmation from a GUI element)
// in a threadsafe manner and retrieve that value later.
type Confirmer struct {
	requests map[string]*Request
	locker   sync.Locker
}

func New() *Confirmer {
	return &Confirmer{
		requests: make(map[string]*Request),
		locker:   &sync.Mutex{},
	}
}

// NewRequest creates a new request object that waits up to the given amount of time for the result.
func (c *Confirmer) NewRequest(timeout time.Duration) *Request {
	c.locker.Lock()
	defer c.locker.Unlock()

	req := newRequest(timeout)

	c.requests[req.ID()] = req

	return req
}

// SetResult sets the result value of the request with the given ID.
func (c *Confirmer) SetResult(id string, value bool) error {
	c.locker.Lock()
	defer c.locker.Unlock()

	req, ok := c.requests[id]
	if !ok {
		return errors.New("no such request")
	}

	req.ch <- value

	close(req.ch)
	delete(c.requests, id)

	return nil
}
