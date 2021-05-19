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

package liveapi

import (
	"context"
	"fmt"
	"math/rand"
	"os"

	"github.com/ProtonMail/proton-bridge/internal/constants"
	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
	"github.com/ProtonMail/proton-bridge/pkg/srp"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// persistentClients keeps authenticated clients for tests.
//
// We need to reduce the number of authentication done by live tests.
// Before every *scenario* we are creating and authenticating new client.
// This is not necessary for controller purposes. We can reuse the same clients
// for all tests.
//
//nolint[gochecknoglobals]
var persistentClients = struct {
	manager    pmapi.Manager
	byName     map[string]pmapi.Client
	saltByName map[string]string
}{}

type persistentClient struct {
	pmapi.Client
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

func SetupPersistentClients() {
	app := os.Getenv("TEST_APP")

	persistentClients.manager = pmapi.New(pmapi.NewConfig(getAppVersionName(app), constants.Version))
	persistentClients.manager.SetLogging(logrus.WithField("pkg", "liveapi"), logrus.GetLevel() == logrus.TraceLevel)

	persistentClients.byName = map[string]pmapi.Client{}
	persistentClients.saltByName = map[string]string{}
}

func getAppVersionName(app string) string {
	if app == "ie" {
		return "importExport"
	}
	return app
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

	srp.RandReader = rand.New(rand.NewSource(42)) //nolint[gosec] It is OK to use weaker random number generator here

	client, _, err := persistentClients.manager.NewClientWithLogin(context.Background(), username, password)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create new persistent client")
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
