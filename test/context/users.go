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

package context

import (
	"context"
	"fmt"
	"math/rand"
	"path/filepath"
	"time"

	"github.com/ProtonMail/go-srp"
	"github.com/ProtonMail/proton-bridge/v2/internal/store"
	"github.com/ProtonMail/proton-bridge/v2/internal/users"
	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

// GetUsers returns users instance.
func (ctx *TestContext) GetUsers() *users.Users {
	return ctx.users
}

// LoginUser logs in the user with the given username, password, and mailbox password.
func (ctx *TestContext) LoginUser(username string, password, mailboxPassword []byte) error {
	srp.RandReader = rand.New(rand.NewSource(42)) //nolint:gosec // It is OK to use weaker random number generator here

	client, auth, err := ctx.users.Login(username, password)
	if err != nil {
		return errors.Wrap(err, "failed to login")
	}

	if auth.HasTwoFactor() {
		if err := client.Auth2FA(context.Background(), "2fa code"); err != nil {
			return errors.Wrap(err, "failed to login with 2FA")
		}
	}

	user, err := ctx.users.FinishLogin(client, auth, mailboxPassword)
	if err != nil {
		return errors.Wrap(err, "failed to finish login")
	}

	ctx.addCleanupChecked(user.Logout, "Logging out user")

	return nil
}

// FinishLogin prevents authentication if not necessary.
func (ctx *TestContext) FinishLogin(client pmapi.Client, mailboxPassword []byte) error {
	type currentAuthGetter interface {
		GetCurrentAuth() *pmapi.Auth
	}

	c, ok := client.(currentAuthGetter)
	if c == nil || !ok {
		return errors.New("cannot get current auth tokens from client")
	}

	user, err := ctx.users.FinishLogin(client, c.GetCurrentAuth(), mailboxPassword)
	if err != nil {
		return errors.Wrap(err, "failed to finish login")
	}

	ctx.addCleanupChecked(user.Logout, "Logging out user")

	return nil
}

// GetUser retrieves the bridge user matching the given query string.
func (ctx *TestContext) GetUser(username string) (*users.User, error) {
	return ctx.users.GetUser(username)
}

// GetStore retrieves the store for given username.
func (ctx *TestContext) GetStore(username string) (*store.Store, error) {
	user, err := ctx.GetUser(username)
	if err != nil {
		return nil, err
	}
	return user.GetStore(), nil
}

// GetStoreAddress retrieves the store address for given username and addressID.
func (ctx *TestContext) GetStoreAddress(username, addressID string) (*store.Address, error) {
	store, err := ctx.GetStore(username)
	if err != nil {
		return nil, err
	}
	return store.GetAddress(addressID)
}

// GetStoreMailbox retrieves the store mailbox for given username, address ID and mailbox name.
func (ctx *TestContext) GetStoreMailbox(username, addressID, mailboxName string) (*store.Mailbox, error) {
	address, err := ctx.GetStoreAddress(username, addressID)
	if err != nil {
		return nil, err
	}
	return address.GetMailbox(mailboxName)
}

// GetDatabaseFilePath returns the file path of the user's store file.
func (ctx *TestContext) GetDatabaseFilePath(userID string) string {
	// We cannot use store to get information because we need to check db file also when user is deleted from bridge.
	fileName := fmt.Sprintf("mailbox-%v.db", userID)
	return filepath.Join(ctx.cache.GetDBDir(), fileName)
}

// WaitForSync waits for sync to be done.
func (ctx *TestContext) WaitForSync(username string) error {
	store, err := ctx.GetStore(username)
	if err != nil {
		return err
	}
	if store == nil {
		return nil
	}
	// First wait for ongoing sync to be done before starting and waiting for new one.
	ctx.eventuallySyncIsFinished(store)
	store.TestSync()
	ctx.eventuallySyncIsFinished(store)
	return nil
}

func (ctx *TestContext) eventuallySyncIsFinished(store *store.Store) {
	assert.Eventually(ctx.t, func() bool { return !store.TestIsSyncRunning() }, 30*time.Second, 10*time.Millisecond)
}

// EventuallySyncIsFinishedForUsername will wait until sync is finished or
// deadline is reached see eventuallySyncIsFinished for timing.
func (ctx *TestContext) EventuallySyncIsFinishedForUsername(username string) {
	store, err := ctx.GetStore(username)
	assert.Nil(ctx.t, err)
	ctx.eventuallySyncIsFinished(store)
}

// LogoutUser logs out the given user.
func (ctx *TestContext) LogoutUser(query string) (err error) {
	user, err := ctx.users.GetUser(query)
	if err != nil {
		return errors.Wrap(err, "failed to get user")
	}

	if err = user.Logout(); err != nil {
		return errors.Wrap(err, "failed to logout user")
	}

	return
}

// DeleteUser deletes the given user.
func (ctx *TestContext) DeleteUser(query string, deleteStore bool) (err error) {
	user, err := ctx.users.GetUser(query)
	if err != nil {
		return errors.Wrap(err, "failed to get user")
	}

	if err = ctx.users.DeleteUser(user.ID(), deleteStore); err != nil {
		err = errors.Wrap(err, "failed to delete user")
	}

	return
}
