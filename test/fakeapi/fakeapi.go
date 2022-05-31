// Copyright (c) 2022 Proton AG
//
// This file is part of Proton Mail Bridge.Bridge.
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
	"errors"
	"fmt"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
	"github.com/sirupsen/logrus"
)

var errBadRequest = errors.New("NOT OK: 400 Bad Request")

type FakePMAPI struct {
	username         string
	userID           string
	controller       *Controller
	eventIDGenerator idGenerator

	authHandlers []pmapi.AuthRefreshHandler
	user         *pmapi.User
	userKeyRing  *crypto.KeyRing
	addresses    *pmapi.AddressList
	addrKeyRing  map[string]*crypto.KeyRing
	labels       []*pmapi.Label
	messages     []*pmapi.Message
	events       []*pmapi.Event

	// uid represents the API UID. It is the unique session ID.
	uid string
	acc string
	ref string

	log *logrus.Entry
}

func newFakePMAPI(controller *Controller, username, userID, uid, acc, ref string) (*FakePMAPI, error) {
	user, ok := controller.usersByUsername[username]
	if !ok {
		return nil, fmt.Errorf("user %s does not exist", username)
	}

	addresses, ok := controller.addressesByUsername[username]
	if !ok {
		addresses = &pmapi.AddressList{}
	}

	labels, ok := controller.labelsByUsername[username]
	if !ok {
		labels = []*pmapi.Label{}
	}

	messages, ok := controller.messagesByUsername[username]
	if !ok {
		messages = []*pmapi.Message{}
	}

	fakePMAPI := &FakePMAPI{
		username:   username,
		userID:     userID,
		controller: controller,

		user:      user.user,
		addresses: addresses,
		labels:    labels,
		messages:  messages,

		uid:         uid,
		acc:         acc,
		ref:         ref,
		addrKeyRing: make(map[string]*crypto.KeyRing),

		log: logrus.WithField("pkg", "fakeapi").WithField("uid", uid).WithField("username", username),
	}

	fakePMAPI.addEvent(&pmapi.Event{
		EventID: fakePMAPI.eventIDGenerator.last("event"),
		Refresh: 0,
		More:    false,
	})

	return fakePMAPI, nil
}

func (api *FakePMAPI) CloseConnections() {
	// NOOP
}

func (api *FakePMAPI) checkAndRecordCall(method method, path string, request interface{}) error {
	api.controller.locker.Lock()
	defer api.controller.locker.Unlock()

	api.log.WithField(string(method), path).Trace("CALL")

	if err := api.controller.checkAndRecordCall(method, path, request); err != nil {
		return err
	}

	if !api.controller.checkAccessToken(api.uid, api.acc) {
		if err := api.authRefresh(); err != nil {
			return err
		}
	}

	if path != "/auth/2fa" && !api.controller.checkScope(api.uid) {
		return errors.New("Access token does not have sufficient scope") //nolint:stylecheck
	}

	return nil
}

func (api *FakePMAPI) authRefresh() error {
	if err := api.controller.checkAndRecordCall(POST, "/auth/refresh", []string{api.uid, api.ref}); err != nil {
		return err
	}

	session, err := api.controller.refreshSessionIfAuthorized(api.uid, api.ref)
	if err != nil {
		if pmapi.IsFailedAuth(err) {
			go api.handleAuth(nil)
		}
		return err
	}

	api.ref = session.ref
	api.acc = session.acc

	go api.handleAuth(&pmapi.AuthRefresh{
		UID:          api.uid,
		AccessToken:  api.acc,
		RefreshToken: api.ref,
		ExpiresIn:    7200,
		Scopes:       []string{"full", "self", "user", "mail"},
	})

	return nil
}

func (api *FakePMAPI) handleAuth(auth *pmapi.AuthRefresh) {
	for _, handle := range api.authHandlers {
		handle(auth)
	}
}

func (api *FakePMAPI) setUser(username string) error {
	api.username = username
	api.log = api.log.WithField("username", username)

	user, ok := api.controller.usersByUsername[username]
	if !ok {
		return fmt.Errorf("user %s does not exist", username)
	}
	api.user = user.user

	addresses, ok := api.controller.addressesByUsername[username]
	if !ok {
		addresses = &pmapi.AddressList{}
	}
	api.addresses = addresses

	labels, ok := api.controller.labelsByUsername[username]
	if !ok {
		labels = []*pmapi.Label{}
	}
	api.labels = labels

	messages, ok := api.controller.messagesByUsername[username]
	if !ok {
		messages = []*pmapi.Message{}
	}
	api.messages = messages

	return nil
}
