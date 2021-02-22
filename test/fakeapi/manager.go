package fakeapi

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
	"github.com/go-resty/resty/v2"
)

type fakePMAPIManager struct {
	controller *Controller
}

func (m *fakePMAPIManager) NewClient(uid string, acc string, ref string, _ time.Time) pmapi.Client {
	session, ok := m.controller.sessionsByUID[uid]
	if !ok {
		return newFakePMAPI(m.controller, "", "", "", "")
	}

	user, ok := m.controller.usersByUsername[session.username]
	if !ok {
		return newFakePMAPI(m.controller, "", "", "", "")
	}

	client, err := NewFakePMAPI(m.controller, session.username, user.user.ID, session.uid, session.acc, session.ref)
	if err != nil {
		return newFakePMAPI(m.controller, "", "", "", "")
	}

	m.controller.fakeAPIs = append(m.controller.fakeAPIs, client)

	return client
}

func (m *fakePMAPIManager) NewClientWithRefresh(_ context.Context, uid, ref string) (pmapi.Client, *pmapi.Auth, error) {
	if err := m.controller.recordCall(POST, "/auth/refresh", &pmapi.AuthRefreshReq{
		UID:          uid,
		RefreshToken: ref,
		ResponseType: "token",
		GrantType:    "refresh_token",
		RedirectURI:  "https://protonmail.ch",
		State:        "random_string",
	}); err != nil {
		return nil, nil, err
	}

	session, err := m.controller.refreshSessionIfAuthorized(uid, ref)
	if err != nil {
		return nil, nil, pmapi.ErrUnauthorized
	}

	user, ok := m.controller.usersByUsername[session.username]
	if !ok {
		return nil, nil, errWrongNameOrPassword
	}

	client, err := NewFakePMAPI(m.controller, session.username, user.user.ID, session.uid, session.acc, session.ref)
	if err != nil {
		return nil, nil, err
	}

	m.controller.fakeAPIs = append(m.controller.fakeAPIs, client)

	auth := &pmapi.Auth{
		UID:          session.uid,
		AccessToken:  session.acc,
		RefreshToken: session.ref,
		ExpiresIn:    86400, // seconds,
	}

	if user.has2FA {
		auth.TwoFA = pmapi.TwoFAInfo{
			Enabled: pmapi.TOTPEnabled,
		}
	}

	return client, auth, nil
}

func (m *fakePMAPIManager) NewClientWithLogin(_ context.Context, username string, password string) (pmapi.Client, *pmapi.Auth, error) {
	if err := m.controller.recordCall(POST, "/auth/info", &pmapi.GetAuthInfoReq{Username: username}); err != nil {
		return nil, nil, err
	}

	// If username is wrong, API server will return empty but positive response.
	// However, we will fail to create a client, so we return error here.
	user, ok := m.controller.usersByUsername[username]
	if !ok {
		return nil, nil, errWrongNameOrPassword
	}

	if err := m.controller.recordCall(POST, "/auth", &pmapi.AuthReq{Username: username}); err != nil {
		return nil, nil, err
	}

	session, err := m.controller.createSessionIfAuthorized(username, password)
	if err != nil {
		return nil, nil, err
	}

	client, err := NewFakePMAPI(m.controller, username, user.user.ID, session.uid, session.acc, session.ref)
	if err != nil {
		return nil, nil, err
	}

	m.controller.fakeAPIs = append(m.controller.fakeAPIs, client)

	auth := &pmapi.Auth{
		UID:          session.uid,
		AccessToken:  session.acc,
		RefreshToken: session.ref,
		ExpiresIn:    86400, // seconds,
	}

	if user.has2FA {
		auth.TwoFA = pmapi.TwoFAInfo{
			Enabled: pmapi.TOTPEnabled,
		}
	}

	return client, auth, nil
}

func (*fakePMAPIManager) DownloadAndVerify(kr *crypto.KeyRing, url, sig string) ([]byte, error) {
	panic("TODO")
}

func (*fakePMAPIManager) ReportBug(context.Context, pmapi.ReportBugReq) error {
	panic("TODO")
}

func (m *fakePMAPIManager) SendSimpleMetric(_ context.Context, cat string, act string, lab string) error {
	v := url.Values{}

	v.Set("Category", cat)
	v.Set("Action", act)
	v.Set("Label", lab)

	return m.controller.recordCall(GET, "/metrics?"+v.Encode(), nil)
}

func (*fakePMAPIManager) SetLogger(resty.Logger) {
	panic("TODO")
}

func (*fakePMAPIManager) SetTransport(http.RoundTripper) {
	panic("TODO")
}

func (*fakePMAPIManager) SetCookieJar(http.CookieJar) {
	panic("TODO")
}

func (*fakePMAPIManager) SetRetryCount(int) {
	panic("TODO")
}

func (*fakePMAPIManager) AddConnectionObserver(pmapi.ConnectionObserver) {
	panic("TODO")
}
