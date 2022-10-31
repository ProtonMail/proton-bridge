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

package pmapi

import (
	"context"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	errCodeUpgradeApplication   = 5003
	errCodePasswordWrong        = 8002
	errCodeAuthPaidPlanRequired = 10004
)

type Error struct {
	Code    int
	Message string `json:"Error"`
}

func (err Error) Error() string {
	return err.Message
}

func (m *manager) catchAPIError(_ *resty.Client, res *resty.Response) error {
	if !res.IsError() {
		return nil
	}

	if res.StatusCode() == http.StatusUnauthorized {
		return ErrUnauthorized
	}

	var err error

	if apiErr, ok := res.Error().(*Error); ok {
		switch {
		case apiErr.Code == errCodeUpgradeApplication:
			if m.cfg.UpgradeApplicationHandler != nil {
				m.cfg.UpgradeApplicationHandler()
			}
			return ErrUpgradeApplication
		case apiErr.Code == errCodePasswordWrong:
			return ErrPasswordWrong
		case apiErr.Code == errCodeAuthPaidPlanRequired:
			return ErrPaidPlanRequired
		default:
			err = apiErr
		}
	} else {
		err = errors.New(res.Status())
	}

	switch res.StatusCode() {
	case http.StatusUnprocessableEntity:
		err = ErrUnprocessableEntity{err}
	case http.StatusBadRequest:
		err = ErrBadRequest{err}
	}

	return err
}

func updateTime(_ *resty.Client, res *resty.Response) error {
	if date, err := time.Parse(time.RFC1123, res.Header().Get("Date")); err != nil {
		log.WithError(err).Warning("Cannot parse header date")
	} else {
		crypto.UpdateTime(date.Unix())
	}

	return nil
}

func logConnReuse(_ *resty.Client, res *resty.Response) error {
	if !res.Request.TraceInfo().IsConnReused {
		logrus.WithField("host", res.Request.URL).Trace("Connection was NOT reused")
	}

	return nil
}

func catchRetryAfter(_ *resty.Client, res *resty.Response) (time.Duration, error) {
	if res.StatusCode() == http.StatusTooManyRequests {
		if after := res.Header().Get("Retry-After"); after != "" {
			seconds, err := strconv.Atoi(after)
			if err != nil {
				log.WithError(err).Warning("Cannot convert Retry-After to number")
				seconds = 10
			}

			// To avoid spikes when all clients retry at the same time, we add some random wait.
			seconds += rand.Intn(10) //nolint:gosec // It is OK to use weak random number generator here.

			log.Warningf("Retrying %s after %ds induced by http code %d", res.Request.URL, seconds, res.StatusCode())
			return time.Duration(seconds) * time.Second, nil
		}
	}

	// 0 and no error means default behaviour which is exponential backoff with jitter.
	return 0, nil
}

func (m *manager) shouldRetry(res *resty.Response, err error) bool {
	if isRetryDisabled(res.Request.Context()) {
		return false
	}
	if isTooManyRequest(res) {
		return true
	}
	if isNoResponse(res, err) {
		// Even if the context of request allows to retry we should check
		// whether the server is reachable or not. In some cases the we can
		// keep retrying but also report that connection is lost.
		go m.pingUntilSuccess()
		return true
	}
	return false
}

func isTooManyRequest(res *resty.Response) bool {
	return res.StatusCode() == http.StatusTooManyRequests
}

func isNoResponse(res *resty.Response, err error) bool {
	// Do not retry TLS failures
	if errors.Is(err, ErrTLSMismatch) {
		return false
	}
	return res.RawResponse == nil && err != nil
}

func wrapNoConnection(res *resty.Response, err error) (*resty.Response, error) {
	if err, ok := err.(*resty.ResponseError); ok {
		return res, err
	}

	if errors.Is(err, context.Canceled) {
		return res, err
	}

	if res.RawResponse != nil {
		return res, err
	}

	// Log useful information and return back nicer and clear error message.
	logrus.WithError(err).WithField("url", res.Request.URL).Warn("No internet connection")
	return res, ErrNoConnection
}
