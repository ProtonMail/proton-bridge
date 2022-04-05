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

package semaphore

import "sync"

type Semaphore struct {
	ch chan struct{}
	wg sync.WaitGroup
}

func New(max int) Semaphore {
	return Semaphore{ch: make(chan struct{}, max)}
}

func (sem *Semaphore) Lock() {
	sem.ch <- struct{}{}
}

func (sem *Semaphore) Unlock() {
	<-sem.ch
}

func (sem *Semaphore) Go(fn func()) {
	sem.Lock()
	sem.wg.Add(1)

	go func() {
		defer sem.Unlock()
		defer sem.wg.Done()

		fn()
	}()
}

func (sem *Semaphore) Wait() {
	sem.wg.Wait()
}
