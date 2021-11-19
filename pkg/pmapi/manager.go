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

package pmapi

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
)

type manager struct {
	cfg Config
	rc  *resty.Client

	isDown              bool
	locker              sync.Locker
	connectionObservers []ConnectionObserver
	proxyDialer         *ProxyTLSDialer
}

func New(cfg Config) Manager {
	return newManager(cfg)
}

func newManager(cfg Config) *manager {
	m := &manager{
		cfg:    cfg,
		rc:     resty.New().EnableTrace(),
		locker: &sync.Mutex{},
	}

	proxyDialer, transport := newProxyDialerAndTransport(cfg)
	m.proxyDialer = proxyDialer
	m.rc.SetTransport(transport)

	m.rc.SetHostURL(cfg.HostURL)
	m.rc.OnBeforeRequest(m.setHeaderValues)

	// Any HTTP status code higher than 399 with JSON inside (and proper header)
	// is converted to Error. `catchAPIError` then processes API custom errors
	// wrapped in JSON. If error is returned, `handleRequestFailure` is called,
	// otherwise `handleRequestSuccess` is called.
	m.rc.SetError(&Error{})
	m.rc.OnAfterResponse(logConnReuse)
	m.rc.OnAfterResponse(updateTime)
	m.rc.OnAfterResponse(m.catchAPIError)
	m.rc.OnAfterResponse(m.handleRequestSuccess)
	m.rc.OnError(m.handleRequestFailure)

	// Configure retry mechanism.
	m.rc.SetRetryMaxWaitTime(time.Minute)
	m.rc.SetRetryAfter(catchRetryAfter)
	m.rc.AddRetryCondition(shouldRetry)

	return m
}

func (m *manager) SetTransport(transport http.RoundTripper) {
	m.rc.SetTransport(transport)
	m.proxyDialer = nil
}

func (m *manager) SetCookieJar(jar http.CookieJar) {
	m.rc.SetCookieJar(jar)
}

func (m *manager) SetRetryCount(count int) {
	m.rc.SetRetryCount(count)
}

func (m *manager) AddConnectionObserver(observer ConnectionObserver) {
	m.connectionObservers = append(m.connectionObservers, observer)
}

func (m *manager) setHeaderValues(_ *resty.Client, req *resty.Request) error {
	req.SetHeaders(map[string]string{
		"x-pm-appversion": m.cfg.AppVersion,
		"User-Agent":      m.cfg.getUserAgent(),
	})
	return nil
}

func (m *manager) r(ctx context.Context) *resty.Request {
	return m.rc.R().SetContext(ctx)
}

func (m *manager) handleRequestSuccess(_ *resty.Client, res *resty.Response) error {
	m.locker.Lock()
	defer m.locker.Unlock()

	if !m.isDown {
		return nil
	}

	// We successfully got a response; connection must be up.

	m.isDown = false

	for _, observer := range m.connectionObservers {
		observer.OnUp()
	}

	return nil
}

func (m *manager) handleRequestFailure(req *resty.Request, err error) {
	m.locker.Lock()
	defer m.locker.Unlock()

	if m.isDown {
		return
	}

	if res, ok := err.(*resty.ResponseError); ok && res.Response.RawResponse != nil {
		return
	}

	// We didn't get any response; connection must be down.

	m.isDown = true

	for _, observer := range m.connectionObservers {
		observer.OnDown()
	}

	go m.pingUntilSuccess()
}
