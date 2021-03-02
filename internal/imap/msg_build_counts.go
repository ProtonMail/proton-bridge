// Copyright (c) 2021 Proton Technologies AG
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

package imap

import (
	"fmt"
	"sync"
)

// msgBuildCountHistogram is used to analyse and log the number of repetitive
// downloads of requested messages per one fetch. The number of builds per each
// messageID is stored in persistent database. The msgBuildCountHistogram will
// take this number for each message in ongoing fetch and create histogram of
// repeats.
//
// Example: During `fetch 1:300` there were
// - 100 messages were downloaded first time
// - 100 messages were downloaded second time
// - 99 messages were downloaded 10th times
// - 1 messages were downloaded 100th times
type msgBuildCountHistogram struct {
	// Key represents how many times message was build.
	// Value stores how many messages are build X times based on the key.
	counts map[uint32]uint32
	lock   sync.Locker
}

func newMsgBuildCountHistogram() *msgBuildCountHistogram {
	return &msgBuildCountHistogram{
		counts: map[uint32]uint32{},
		lock:   &sync.Mutex{},
	}
}

func (c *msgBuildCountHistogram) String() string {
	res := ""
	for nRebuild, counts := range c.counts {
		if res != "" {
			res += ", "
		}
		res += fmt.Sprintf("[%d]:%d", nRebuild, counts)
	}
	return res
}

func (c *msgBuildCountHistogram) add(nRebuild uint32) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.counts[nRebuild]++
}
