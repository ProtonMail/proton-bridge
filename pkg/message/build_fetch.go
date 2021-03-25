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
	"context"
	"io/ioutil"
	"sync"

	"github.com/ProtonMail/proton-bridge/pkg/parallel"
	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
)

type fetchReq struct {
	ctx       context.Context
	api       Fetcher
	messageID string
	opts      JobOptions
}

type fetchRes struct {
	fetchReq

	msg  *pmapi.Message
	atts [][]byte
	err  error
}

func newFetchResSuccess(req fetchReq, msg *pmapi.Message, atts [][]byte) fetchRes {
	return fetchRes{
		fetchReq: req,
		msg:      msg,
		atts:     atts,
	}
}

func newFetchResFailure(req fetchReq, err error) fetchRes {
	return fetchRes{
		fetchReq: req,
		err:      err,
	}
}

// startFetchWorkers starts the given number of fetch workers.
// These workers download message and attachment data from API.
// Each fetch worker will use up to the given number of attachment workers to download attachments.
// Two channels are returned:
//  - fetchReqCh: used to send work items to the worker pool
//  - fetchResCh: used to receive work results from the worker pool
func startFetchWorkers(fetchWorkers, attachWorkers int) (chan fetchReq, chan fetchRes) {
	fetchReqCh := make(chan fetchReq)
	fetchResCh := make(chan fetchRes)

	go func() {
		defer close(fetchResCh)

		var wg sync.WaitGroup

		wg.Add(fetchWorkers)

		for workerID := 0; workerID < fetchWorkers; workerID++ {
			go fetchWorker(fetchReqCh, fetchResCh, attachWorkers, &wg)
		}

		wg.Wait()
	}()

	return fetchReqCh, fetchResCh
}

func fetchWorker(fetchReqCh <-chan fetchReq, fetchResCh chan<- fetchRes, attachWorkers int, wg *sync.WaitGroup) {
	defer wg.Done()

	for req := range fetchReqCh {
		msg, atts, err := fetchMessage(req, attachWorkers)
		if err != nil {
			fetchResCh <- newFetchResFailure(req, err)
		} else {
			fetchResCh <- newFetchResSuccess(req, msg, atts)
		}
	}
}

func fetchMessage(req fetchReq, attachWorkers int) (*pmapi.Message, [][]byte, error) {
	msg, err := req.api.GetMessage(req.messageID)
	if err != nil {
		return nil, nil, err
	}

	attList := make([]interface{}, len(msg.Attachments))

	for i, att := range msg.Attachments {
		attList[i] = att.ID
	}

	process := func(value interface{}) (interface{}, error) {
		rc, err := req.api.GetAttachment(value.(string))
		if err != nil {
			return nil, err
		}

		b, err := ioutil.ReadAll(rc)
		if err != nil {
			return nil, err
		}

		if err := rc.Close(); err != nil {
			return nil, err
		}

		return b, nil
	}

	attData := make([][]byte, len(msg.Attachments))

	collect := func(idx int, value interface{}) error {
		attData[idx] = value.([]byte)
		return nil
	}

	if err := parallel.RunParallel(attachWorkers, attList, process, collect); err != nil {
		return nil, nil, err
	}

	return msg, attData, nil
}
