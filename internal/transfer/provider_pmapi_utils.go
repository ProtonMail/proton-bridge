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

package transfer

import (
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

		err := p.clientManager.CheckConnection()
		log.WithError(err).Debug("Connection check")
		if err != nil {
			time.Sleep(pmapiReconnectSleep)
			previousErr = err
			continue
		}

		break
	}
	return nil
}

func (p *PMAPIProvider) listMessages(filter *pmapi.MessagesFilter) (messages []*pmapi.Message, count int, err error) {
	err = p.ensureConnection(func() error {
		messages, count, err = p.client().ListMessages(filter)
		return err
	})
	return
}

func (p *PMAPIProvider) getMessage(msgID string) (message *pmapi.Message, err error) {
	err = p.ensureConnection(func() error {
		message, err = p.client().GetMessage(msgID)
		return err
	})
	return
}

func (p *PMAPIProvider) importRequest(req []*pmapi.ImportMsgReq) (res []*pmapi.ImportMsgRes, err error) {
	err = p.ensureConnection(func() error {
		res, err = p.client().Import(req)
		return err
	})
	return
}

func (p *PMAPIProvider) createDraft(message *pmapi.Message, parent string, action int) (draft *pmapi.Message, err error) {
	err = p.ensureConnection(func() error {
		draft, err = p.client().CreateDraft(message, parent, action)
		return err
	})
	return
}

func (p *PMAPIProvider) createAttachment(att *pmapi.Attachment, r io.Reader, sig io.Reader) (created *pmapi.Attachment, err error) {
	err = p.ensureConnection(func() error {
		created, err = p.client().CreateAttachment(att, r, sig)
		return err
	})
	return
}
