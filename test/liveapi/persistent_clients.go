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

package liveapi

import (
	"context"
	"fmt"
	"math/rand"
	"sync"

	"github.com/ProtonMail/go-srp"
	"github.com/ProtonMail/proton-bridge/v2/internal/constants"
	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type clientAuthGetter interface {
	pmapi.Client
	GetCurrentAuth() *pmapi.Auth
}

// persistentClients keeps authenticated clients for tests.
//
// We need to reduce the number of authentication done by live tests.
// Before every *scenario* we are creating and authenticating new client.
// This is not necessary for controller purposes. We can reuse the same clients
// for all tests.
//
//nolint:gochecknoglobals // This is necessary for testing
var persistentClients = struct {
	manager    pmapi.Manager
	byName     map[string]clientAuthGetter
	saltByName map[string]string

	eventsPaused         sync.WaitGroup
	skipDeletedMessageID map[string]struct{}
}{}

type persistentClient struct {
	clientAuthGetter
	username string
}

// AuthDelete is noop. All sessions will be closed in CleanupPersistentClients.
func (pc *persistentClient) AuthDelete(_ context.Context) error {
	return nil
}

// AuthSalt returns cached string. Otherwise after some time there is an error:
//
//     Access token does not have sufficient scope
//
// while all other routes works normally. Need to confirm with Aron that this
// is expected behaviour.
func (pc *persistentClient) AuthSalt(_ context.Context) (string, error) {
	return persistentClients.saltByName[pc.username], nil
}

// GetEvent needs to wait for preparation to finish. Otherwise messages will be
// in wrong order and test will fail.
func (pc *persistentClient) GetEvent(ctx context.Context, eventID string) (*pmapi.Event, error) {
	persistentClients.eventsPaused.Wait()
	normalClient, ok := persistentClients.byName[pc.username].(pmapi.Client)
	if !ok {
		return nil, errors.New("cannot convert to normal client")
	}

	event, err := normalClient.GetEvent(ctx, eventID)
	if err != nil {
		return event, err
	}

	return skipDeletedMessageIDs(event), nil
}

func addMessageIDToSkipEventOnceDeleted(msgID string) {
	if persistentClients.skipDeletedMessageID == nil {
		persistentClients.skipDeletedMessageID = map[string]struct{}{}
	}
	persistentClients.skipDeletedMessageID[msgID] = struct{}{}
}

func skipDeletedMessageIDs(event *pmapi.Event) *pmapi.Event {
	if len(event.Messages) == 0 {
		return event
	}

	n := 0
	for i, m := range event.Messages {
		if _, ok := persistentClients.skipDeletedMessageID[m.ID]; ok && m.Action == pmapi.EventDelete {
			delete(persistentClients.skipDeletedMessageID, m.ID)
			continue
		}

		event.Messages[i] = m
		n++
	}
	event.Messages = event.Messages[:n]

	return event
}

func SetupPersistentClients() {
	app := "bridge"

	persistentClients.manager = pmapi.New(pmapi.NewConfig(app, constants.Version))
	persistentClients.manager.SetLogging(logrus.WithField("pkg", "liveapi"), logrus.GetLevel() == logrus.TraceLevel)

	persistentClients.byName = map[string]clientAuthGetter{}
	persistentClients.saltByName = map[string]string{}
}

func CleanupPersistentClients() {
	for username, client := range persistentClients.byName {
		if err := client.AuthDelete(context.Background()); err != nil {
			logrus.WithError(err).
				WithField("username", username).
				Error("Failed to logout persistent client")
		}
	}
}

func addPersistentClient(username string, password, mailboxPassword []byte) (pmapi.Client, error) {
	if cl, ok := persistentClients.byName[username]; ok {
		return cl, nil
	}

	srp.RandReader = rand.New(rand.NewSource(42)) //nolint:gosec // It is OK to use weaker random number generator here

	normalClient, _, err := persistentClients.manager.NewClientWithLogin(context.Background(), username, password)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create new persistent client")
	}

	client, ok := normalClient.(clientAuthGetter)
	if !ok {
		return nil, errors.New("cannot make clientAuthGetter")
	}

	salt, err := client.AuthSalt(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, "persistent client: failed to get salt")
	}

	hashedMboxPass, err := pmapi.HashMailboxPassword(mailboxPassword, salt)
	if err != nil {
		return nil, errors.Wrap(err, "persistent client: failed to hash mailbox password")
	}

	if err := client.Unlock(context.Background(), hashedMboxPass); err != nil {
		return nil, errors.Wrap(err, "persistent client: failed to unlock user")
	}

	persistentClients.byName[username] = client
	persistentClients.saltByName[username] = salt
	return client, nil
}

func getPersistentClient(username string) (pmapi.Client, error) {
	v, ok := persistentClients.byName[username]
	if !ok {
		return nil, fmt.Errorf("user %s does not exist", username)
	}
	return &persistentClient{v, username}, nil
}
