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

package fakeapi

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
	"github.com/sirupsen/logrus"
)

type fakePMAPIManager struct {
	controller          *Controller
	connectionObservers []pmapi.ConnectionObserver
}

func (m *fakePMAPIManager) NewClient(uid string, acc string, ref string, _ time.Time) pmapi.Client {
	if uid == "" {
		return &FakePMAPI{
			controller:  m.controller,
			log:         logrus.WithField("pkg", "fakeapi"),
			addrKeyRing: make(map[string]*crypto.KeyRing),
		}
	}

	session, ok := m.controller.sessionsByUID[uid]
	if !ok {
		panic("session " + uid + " is missing")
	}

	user, ok := m.controller.usersByUsername[session.username]
	if !ok {
		panic("user " + session.username + " from session " + uid + " is missing")
	}

	client, err := newFakePMAPI(m.controller, session.username, user.user.ID, session.uid, session.acc, session.ref)
	if err != nil {
		panic(err)
	}

	m.controller.fakeAPIs = append(m.controller.fakeAPIs, client)

	return client
}

func (m *fakePMAPIManager) NewClientWithRefresh(_ context.Context, uid, ref string) (pmapi.Client, *pmapi.AuthRefresh, error) {
	if err := m.controller.checkAndRecordCall(POST, "/auth/refresh", []string{uid, ref}); err != nil {
		return nil, nil, err
	}

	session, err := m.controller.refreshSessionIfAuthorized(uid, ref)
	if err != nil {
		return nil, nil, err
	}

	user, ok := m.controller.usersByUsername[session.username]
	if !ok {
		return nil, nil, errWrongNameOrPassword
	}

	client, err := newFakePMAPI(m.controller, session.username, user.user.ID, session.uid, session.acc, session.ref)
	if err != nil {
		return nil, nil, err
	}

	m.controller.fakeAPIs = append(m.controller.fakeAPIs, client)

	auth := &pmapi.AuthRefresh{
		UID:          session.uid,
		AccessToken:  session.acc,
		RefreshToken: session.ref,
		ExpiresIn:    86400, // seconds,
	}

	return client, auth, nil
}

func (m *fakePMAPIManager) NewClientWithLogin(_ context.Context, username string, password []byte) (pmapi.Client, *pmapi.Auth, error) {
	if err := m.controller.checkAndRecordCall(POST, "/auth/info", &pmapi.GetAuthInfoReq{Username: username}); err != nil {
		return nil, nil, err
	}

	// If username is wrong, API server will return empty but positive response.
	// However, we will fail to create a client, so we return error here.
	user, ok := m.controller.usersByUsername[username]
	if !ok {
		return nil, nil, errWrongNameOrPassword
	}

	if err := m.controller.checkAndRecordCall(POST, "/auth", &pmapi.AuthReq{Username: username}); err != nil {
		return nil, nil, err
	}

	session, err := m.controller.createSessionIfAuthorized(username, password)
	if err != nil {
		return nil, nil, err
	}

	client, err := newFakePMAPI(m.controller, username, user.user.ID, session.uid, session.acc, session.ref)
	if err != nil {
		return nil, nil, err
	}

	m.controller.fakeAPIs = append(m.controller.fakeAPIs, client)

	auth := &pmapi.Auth{
		UserID: user.user.ID,
		AuthRefresh: pmapi.AuthRefresh{
			UID:          session.uid,
			AccessToken:  session.acc,
			RefreshToken: session.ref,
			ExpiresIn:    86400, // seconds,
		},
	}

	if user.has2FA {
		auth.TwoFA = &pmapi.TwoFAInfo{
			Enabled: pmapi.TOTPEnabled,
		}
	}

	return client, auth, nil
}

func (m *fakePMAPIManager) DownloadAndVerify(kr *crypto.KeyRing, url, sig string) ([]byte, error) {
	panic("Not implemented: not used by tests")
}

func (m *fakePMAPIManager) ReportBug(_ context.Context, bugReport pmapi.ReportBugReq) error {
	return m.controller.checkAndRecordCall(POST, "/reports/bug", bugReport)
}

func (m *fakePMAPIManager) SendSimpleMetric(_ context.Context, cat string, act string, lab string) error {
	v := url.Values{}
	v.Set("Category", cat)
	v.Set("Action", act)
	v.Set("Label", lab)
	return m.controller.checkAndRecordCall(GET, "/metrics?"+v.Encode(), nil)
}

func (m *fakePMAPIManager) SetLogging(*logrus.Entry, bool) {
	// NOOP
}

func (m *fakePMAPIManager) SetTransport(http.RoundTripper) {
	// NOOP
}

func (m *fakePMAPIManager) SetCookieJar(http.CookieJar) {
	// NOOP
}

func (m *fakePMAPIManager) SetRetryCount(int) {
	// NOOP
}

func (m *fakePMAPIManager) AddConnectionObserver(connectionObserver pmapi.ConnectionObserver) {
	m.connectionObservers = append(m.connectionObservers, connectionObserver)
}

func (m *fakePMAPIManager) AllowProxy() {
	// NOOP
}

func (m *fakePMAPIManager) DisallowProxy() {
	// NOOP
}
