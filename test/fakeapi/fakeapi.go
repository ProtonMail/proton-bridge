// Copyright (c) 2020 Proton Technologies AG
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

	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
	"github.com/sirupsen/logrus"
)

var errBadRequest = errors.New("NOT OK: 400 Bad Request")

type FakePMAPI struct {
	username         string
	controller       *Controller
	eventIDGenerator idGenerator

	auths     chan<- *pmapi.Auth
	user      *pmapi.User
	addresses *pmapi.AddressList
	labels    []*pmapi.Label
	messages  []*pmapi.Message
	events    []*pmapi.Event

	// uid represents the API UID. It is the unique session ID.
	uid, lastToken string

	log *logrus.Entry
}

func New(controller *Controller) *FakePMAPI {
	fakePMAPI := &FakePMAPI{
		controller: controller,
		log:        logrus.WithField("pkg", "fakeapi"),
	}
	fakePMAPI.addEvent(&pmapi.Event{
		EventID: fakePMAPI.eventIDGenerator.last("event"),
		Refresh: 0,
		More:    0,
	})
	return fakePMAPI
}

func (api *FakePMAPI) checkAndRecordCall(method method, path string, request interface{}) error {
	if err := api.checkInternetAndRecordCall(method, path, request); err != nil {
		return err
	}

	// Try re-auth
	if api.uid == "" && api.lastToken != "" {
		api.log.WithField("lastToken", api.lastToken).Warn("Handling unauthorized status")
		if _, err := api.AuthRefresh(api.lastToken); err != nil {
			return err
		}
	}

	// Check client is authenticated. There is difference between
	//    * invalid token
	//    * and missing token
	// but API treats it the same
	if api.uid == "" {
		return pmapi.ErrInvalidToken
	}

	// Any route (except Auth and AuthRefresh) can end with wrong
	// token and it should be translated into logout
	session, ok := api.controller.sessionsByUID[api.uid]
	if !ok {
		api.setUID("") // all consecutive requests will not send auth nil
		api.sendAuth(nil)
		return pmapi.ErrInvalidToken
	} else if !session.hasFullScope {
		// This is exact error string from the server (at least from documentation).
		return errors.New("Access token does not have sufficient scope") //nolint[stylecheck]
	}

	return nil
}

func (api *FakePMAPI) checkInternetAndRecordCall(method method, path string, request interface{}) error {
	api.log.WithField(string(method), path).Trace("CALL")
	api.controller.recordCall(method, path, request)
	if api.controller.noInternetConnection {
		return pmapi.ErrAPINotReachable
	}
	return nil
}

func (api *FakePMAPI) sendAuth(auth *pmapi.Auth) {
	go func() {
		api.controller.clientManager.GetClientAuthChannel() <- pmapi.ClientAuth{
			UserID: api.user.ID,
			Auth:   auth,
		}
	}()
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

func (api *FakePMAPI) setUID(uid string) {
	api.uid = uid
	api.log = api.log.WithField("uid", api.uid)
	api.log.Info("UID updated")
}

func (api *FakePMAPI) unsetUser() {
	api.setUID("")
	api.user = nil
	api.labels = nil
	api.messages = nil
	api.events = nil
}
