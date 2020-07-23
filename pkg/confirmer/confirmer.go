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
)

// NOTE: For now, Confirmer only supports bool values but it could easily be made generic.

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

func (c *Confirmer) NewRequest(timeout time.Duration) *Request {
	c.locker.Lock()
	defer c.locker.Unlock()

	req := newRequest(timeout)

	c.requests[req.ID()] = req

	return req
}

func (c *Confirmer) SetResponse(uuid string, value bool) error {
	c.locker.Lock()
	defer c.locker.Unlock()

	req, ok := c.requests[uuid]
	if !ok {
		return errors.New("no such request")
	}

	req.value <- value

	return nil
}
