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

package message

import (
	"sync"

	"github.com/pkg/errors"
)

type buildRes struct {
	messageID string
	literal   []byte
	err       error
}

func newBuildResSuccess(messageID string, literal []byte) buildRes {
	return buildRes{
		messageID: messageID,
		literal:   literal,
	}
}

func newBuildResFailure(messageID string, err error) buildRes {
	return buildRes{
		messageID: messageID,
		err:       err,
	}
}

// startBuildWorkers starts the given number of build workers.
// These workers decrypt and build messages into RFC822 literals.
// Two channels are returned:
//  - buildReqCh: used to send work items to the worker pool
//  - buildResCh: used to receive work results from the worker pool
func startBuildWorkers(buildWorkers int) (chan fetchRes, chan buildRes) {
	buildReqCh := make(chan fetchRes)
	buildResCh := make(chan buildRes)

	go func() {
		defer close(buildResCh)

		var wg sync.WaitGroup

		wg.Add(buildWorkers)

		for workerID := 0; workerID < buildWorkers; workerID++ {
			go buildWorker(buildReqCh, buildResCh, &wg)
		}

		wg.Wait()
	}()

	return buildReqCh, buildResCh
}

func buildWorker(buildReqCh <-chan fetchRes, buildResCh chan<- buildRes, wg *sync.WaitGroup) {
	defer wg.Done()

	for req := range buildReqCh {
		if kr, err := req.api.KeyRingForAddressID(req.msg.AddressID); err != nil {
			buildResCh <- newBuildResFailure(req.msg.ID, errors.Wrap(ErrNoSuchKeyRing, err.Error()))
		} else if literal, err := buildRFC822(kr, req.msg, req.atts, req.opts); err != nil {
			buildResCh <- newBuildResFailure(req.msg.ID, err)
		} else {
			buildResCh <- newBuildResSuccess(req.msg.ID, literal)
		}
	}
}
