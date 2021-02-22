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

package transfer

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
	"github.com/pkg/errors"
)

const (
	pmapiRetries          = 10
	pmapiReconnectTimeout = 30 * time.Minute
	pmapiReconnectSleep   = time.Minute
)

func (p *PMAPIProvider) ensureConnection(callback func() error) error {
	var callErr error
	for i := 1; i <= pmapiRetries; i++ {
		callErr = callback()
		if callErr == nil {
			return nil
		}

		log.WithField("attempt", i).WithError(callErr).Warning("API call failed, trying reconnect")
		err := p.tryReconnect()
		if err != nil {
			return err
		}
	}
	return errors.Wrap(callErr, "too many retries")
}

func (p *PMAPIProvider) tryReconnect() error {
	start := time.Now()
	var previousErr error
	for {
		if time.Since(start) > pmapiReconnectTimeout {
			return previousErr
		}

		// FIXME(conman): This should register as a connection observer somehow...
		// Maybe the entire "provider" could register as an observer and pause if it is notified of dropped connection?

		/*
			err := p.clientManager.CheckConnection()
			log.WithError(err).Debug("Connection check")
			if err != nil {
				time.Sleep(pmapiReconnectSleep)
				previousErr = err
				continue
			}
		*/

		break
	}
	return nil
}

func (p *PMAPIProvider) listMessages(filter *pmapi.MessagesFilter) (messages []*pmapi.Message, count int, err error) {
	err = p.ensureConnection(func() error {
		// Sort is used in the key so the filter is different for estimating and real fetching.
		key := fmt.Sprintf("%s_%s_%d", filter.LabelID, filter.Sort, filter.Page)
		p.timeIt.start("listing", key)
		defer p.timeIt.stop("listing", key)

		messages, count, err = p.client.ListMessages(context.TODO(), filter)
		return err
	})
	return
}

func (p *PMAPIProvider) getMessage(msgID string) (message *pmapi.Message, err error) {
	err = p.ensureConnection(func() error {
		p.timeIt.start("download", msgID)
		defer p.timeIt.stop("download", msgID)

		message, err = p.client.GetMessage(context.TODO(), msgID)
		return err
	})
	return
}

func (p *PMAPIProvider) importRequest(msgSourceID string, req pmapi.ImportMsgReqs) (res []*pmapi.ImportMsgRes, err error) {
	err = p.ensureConnection(func() error {
		p.timeIt.start("upload", msgSourceID)
		defer p.timeIt.stop("upload", msgSourceID)

		res, err = p.client.Import(context.TODO(), req)
		return err
	})
	return
}

func (p *PMAPIProvider) createDraft(msgSourceID string, message *pmapi.Message, parent string, action int) (draft *pmapi.Message, err error) {
	err = p.ensureConnection(func() error {
		p.timeIt.start("upload", msgSourceID)
		defer p.timeIt.stop("upload", msgSourceID)

		draft, err = p.client.CreateDraft(context.TODO(), message, parent, action)
		return err
	})
	return
}

func (p *PMAPIProvider) createAttachment(msgSourceID string, att *pmapi.Attachment, r io.Reader, sig io.Reader) (created *pmapi.Attachment, err error) {
	err = p.ensureConnection(func() error {
		// Use some attributes from attachment to have unique key for each call.
		key := fmt.Sprintf("%s_%s_%d", msgSourceID, att.Name, att.Size)
		p.timeIt.start("upload", key)
		defer p.timeIt.stop("upload", key)

		created, err = p.client.CreateAttachment(context.TODO(), att, r, sig)
		return err
	})
	return
}
