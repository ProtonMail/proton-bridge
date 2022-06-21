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

package cache

import "sync"

type pending struct {
	lock sync.Mutex
	path map[string]chan struct{}
}

func newPending() *pending {
	return &pending{path: make(map[string]chan struct{})}
}

func (p *pending) add(path string) bool {
	p.lock.Lock()
	defer p.lock.Unlock()

	if _, ok := p.path[path]; ok {
		return false
	}

	p.path[path] = make(chan struct{})

	return true
}

func (p *pending) wait(path string) {
	p.lock.Lock()
	ch, ok := p.path[path]
	p.lock.Unlock()

	if ok {
		<-ch
	}
}

func (p *pending) done(path string) {
	p.lock.Lock()
	defer p.lock.Unlock()

	defer close(p.path[path])

	delete(p.path, path)
}
