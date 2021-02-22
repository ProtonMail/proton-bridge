// Copyright (c) 2021 Proton Technologies AG
//
// This file is part of ProtonMail Bridge.Bridge.
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

package fakeapi

import (
	"errors"
	"fmt"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
	"github.com/sirupsen/logrus"
)

var errBadRequest = errors.New("NOT OK: 400 Bad Request")

type FakePMAPI struct {
	username         string
	userID           string
	controller       *Controller
	eventIDGenerator idGenerator

	authHandlers []pmapi.AuthHandler
	user         *pmapi.User
	userKeyRing  *crypto.KeyRing
	addresses    *pmapi.AddressList
	addrKeyRing  map[string]*crypto.KeyRing
	labels       []*pmapi.Label
	messages     []*pmapi.Message
	events       []*pmapi.Event

	// uid represents the API UID. It is the unique session ID.
	uid string
	acc string // FIXME(conman): Check this is correct!
	ref string // FIXME(conman): Check this is correct!

	log *logrus.Entry
}

func newFakePMAPI(controller *Controller, userID, uid, acc, ref string) *FakePMAPI {
	return &FakePMAPI{
		controller:  controller,
		log:         logrus.WithField("pkg", "fakeapi").WithField("uid", uid),
		uid:         uid,
		acc:         acc, // FIXME(conman): This should be checked!
		ref:         ref, // FIXME(conman): This should be checked!
		userID:      userID,
		addrKeyRing: make(map[string]*crypto.KeyRing),
	}
}

func NewFakePMAPI(controller *Controller, username, userID, uid, acc, ref string) (*FakePMAPI, error) {
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

	fakePMAPI := newFakePMAPI(controller, userID, uid, acc, ref)

	fakePMAPI.log = fakePMAPI.log.WithField("username", username)
	fakePMAPI.username = username
	fakePMAPI.user = user.user
	fakePMAPI.addresses = addresses
	fakePMAPI.labels = labels
	fakePMAPI.messages = messages

	fakePMAPI.addEvent(&pmapi.Event{
		EventID: fakePMAPI.eventIDGenerator.last("event"),
		Refresh: 0,
		More:    0,
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

	if err := api.controller.recordCall(method, path, request); err != nil {
		return err
	}

	// FIXME(conman): This needs to match conman behaviour. Should try auth refresh somehow.
	if !api.controller.checkAccessToken(api.uid, api.acc) {
		return pmapi.ErrUnauthorized
	}

	if path != "/auth/2fa" && !api.controller.checkScope(api.uid) {
		return errors.New("Access token does not have sufficient scope") //nolint[stylecheck]
	}

	return nil
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

func (api *FakePMAPI) unsetUser() {
	api.uid = ""
	api.acc = "" // FIXME(conman): This should be checked!
	api.user = nil
	api.labels = nil
	api.messages = nil
	api.events = nil
}
