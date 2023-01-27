// Copyright (c) 2023 Proton AG
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
// along with Proton Mail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package bridge

import (
	"context"
	"errors"
	"fmt"
	"runtime"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/proton-bridge/v3/internal/async"
	"github.com/ProtonMail/proton-bridge/v3/internal/events"
	"github.com/ProtonMail/proton-bridge/v3/internal/logging"
	"github.com/ProtonMail/proton-bridge/v3/internal/safe"
	"github.com/ProtonMail/proton-bridge/v3/internal/try"
	"github.com/ProtonMail/proton-bridge/v3/internal/user"
	"github.com/ProtonMail/proton-bridge/v3/internal/vault"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

type UserState int

const (
	SignedOut UserState = iota
	Locked
	Connected
)

type UserInfo struct {
	// UserID is the user's API ID.
	UserID string

	// Username is the user's API username.
	Username string

	// Signed Out is true if the user is signed out (no AuthUID, user will need to provide credentials to log in again)
	State UserState

	// Addresses holds the user's email addresses. The first address is the primary address.
	Addresses []string

	// AddressMode is the user's address mode.
	AddressMode vault.AddressMode

	// BridgePass is the user's bridge password.
	BridgePass []byte

	// UsedSpace is the amount of space used by the user.
	UsedSpace int

	// MaxSpace is the total amount of space available to the user.
	MaxSpace int
}

// GetUserIDs returns the IDs of all known users (authorized or not).
func (bridge *Bridge) GetUserIDs() []string {
	return bridge.vault.GetUserIDs()
}

// HasUser returns true iff the given user is known (authorized or not).
func (bridge *Bridge) HasUser(userID string) bool {
	return bridge.vault.HasUser(userID)
}

// GetUserInfo returns info about the given user.
func (bridge *Bridge) GetUserInfo(userID string) (UserInfo, error) {
	return safe.RLockRetErr(func() (UserInfo, error) {
		if user, ok := bridge.users[userID]; ok {
			return getConnUserInfo(user), nil
		}

		var info UserInfo

		if err := bridge.vault.GetUser(userID, func(user *vault.User) {
			state := Locked
			if len(user.AuthUID()) == 0 {
				state = SignedOut
			}
			info = getUserInfo(user.UserID(), user.Username(), state, user.AddressMode())
		}); err != nil {
			return UserInfo{}, fmt.Errorf("failed to get user info: %w", err)
		}

		return info, nil
	}, bridge.usersLock)
}

// QueryUserInfo queries the user info by username or address.
func (bridge *Bridge) QueryUserInfo(query string) (UserInfo, error) {
	return safe.RLockRetErr(func() (UserInfo, error) {
		for _, user := range bridge.users {
			if user.Match(query) {
				return getConnUserInfo(user), nil
			}
		}

		return UserInfo{}, ErrNoSuchUser
	}, bridge.usersLock)
}

// LoginAuth begins the login process. It returns an authorized client that might need 2FA.
func (bridge *Bridge) LoginAuth(ctx context.Context, username string, password []byte) (*proton.Client, proton.Auth, error) {
	logrus.WithField("username", logging.Sensitive(username)).Info("Authorizing user for login")

	if username == "crash@bandicoot" {
		panic("Your wish is my command.. I crash!")
	}

	client, auth, err := bridge.api.NewClientWithLogin(ctx, username, password)
	if err != nil {
		return nil, proton.Auth{}, fmt.Errorf("failed to create new API client: %w", err)
	}

	if ok := safe.RLockRet(func() bool { return mapHas(bridge.users, auth.UserID) }, bridge.usersLock); ok {
		logrus.WithField("userID", auth.UserID).Warn("User already logged in")

		if err := client.AuthDelete(ctx); err != nil {
			logrus.WithError(err).Warn("Failed to delete auth")
		}

		return nil, proton.Auth{}, ErrUserAlreadyLoggedIn
	}

	return client, auth, nil
}

// LoginUser finishes the user login process using the client and auth received from LoginAuth.
func (bridge *Bridge) LoginUser(
	ctx context.Context,
	client *proton.Client,
	auth proton.Auth,
	keyPass []byte,
) (string, error) {
	logrus.WithField("userID", auth.UserID).Info("Logging in authorized user")

	userID, err := try.CatchVal(
		func() (string, error) {
			return bridge.loginUser(ctx, client, auth.UID, auth.RefreshToken, keyPass)
		},
		func() error {
			return client.AuthDelete(ctx)
		},
	)
	if err != nil {
		return "", fmt.Errorf("failed to login user: %w", err)
	}

	bridge.publish(events.UserLoggedIn{
		UserID: userID,
	})

	return userID, nil
}

// LoginFull authorizes a new bridge user with the given username and password.
// If necessary, a TOTP and mailbox password are requested via the callbacks.
// This is equivalent to doing LoginAuth and LoginUser separately.
func (bridge *Bridge) LoginFull(
	ctx context.Context,
	username string,
	password []byte,
	getTOTP func() (string, error),
	getKeyPass func() ([]byte, error),
) (string, error) {
	logrus.WithField("username", logging.Sensitive(username)).Info("Performing full user login")

	client, auth, err := bridge.LoginAuth(ctx, username, password)
	if err != nil {
		return "", fmt.Errorf("failed to begin login process: %w", err)
	}

	if auth.TwoFA.Enabled&proton.HasTOTP != 0 {
		logrus.WithField("userID", auth.UserID).Info("Requesting TOTP")

		totp, err := getTOTP()
		if err != nil {
			return "", fmt.Errorf("failed to get TOTP: %w", err)
		}

		if err := client.Auth2FA(ctx, proton.Auth2FAReq{TwoFactorCode: totp}); err != nil {
			return "", fmt.Errorf("failed to authorize 2FA: %w", err)
		}
	}

	var keyPass []byte

	if auth.PasswordMode == proton.TwoPasswordMode {
		logrus.WithField("userID", auth.UserID).Info("Requesting mailbox password")

		userKeyPass, err := getKeyPass()
		if err != nil {
			return "", fmt.Errorf("failed to get key password: %w", err)
		}

		keyPass = userKeyPass
	} else {
		keyPass = password
	}

	return bridge.LoginUser(ctx, client, auth, keyPass)
}

// LogoutUser logs out the given user.
func (bridge *Bridge) LogoutUser(ctx context.Context, userID string) error {
	logrus.WithField("userID", userID).Info("Logging out user")

	return safe.LockRet(func() error {
		user, ok := bridge.users[userID]
		if !ok {
			return ErrNoSuchUser
		}

		bridge.logoutUser(ctx, user, true, false)

		bridge.publish(events.UserLoggedOut{
			UserID: userID,
		})

		return nil
	}, bridge.usersLock)
}

// DeleteUser deletes the given user.
func (bridge *Bridge) DeleteUser(ctx context.Context, userID string) error {
	logrus.WithField("userID", userID).Info("Deleting user")

	return safe.LockRet(func() error {
		if !bridge.vault.HasUser(userID) {
			return ErrNoSuchUser
		}

		if user, ok := bridge.users[userID]; ok {
			bridge.logoutUser(ctx, user, true, true)
		}

		if err := bridge.vault.DeleteUser(userID); err != nil {
			logrus.WithError(err).Error("Failed to delete vault user")
		}

		bridge.publish(events.UserDeleted{
			UserID: userID,
		})

		return nil
	}, bridge.usersLock)
}

// SetAddressMode sets the address mode for the given user.
func (bridge *Bridge) SetAddressMode(ctx context.Context, userID string, mode vault.AddressMode) error {
	logrus.WithField("userID", userID).WithField("mode", mode).Info("Setting address mode")

	return safe.RLockRet(func() error {
		user, ok := bridge.users[userID]
		if !ok {
			return ErrNoSuchUser
		}

		if user.GetAddressMode() == mode {
			return fmt.Errorf("address mode is already %q", mode)
		}

		if err := bridge.removeIMAPUser(ctx, user, true); err != nil {
			return fmt.Errorf("failed to remove IMAP user: %w", err)
		}

		if err := user.SetAddressMode(ctx, mode); err != nil {
			return fmt.Errorf("failed to set address mode: %w", err)
		}

		if err := bridge.addIMAPUser(ctx, user); err != nil {
			return fmt.Errorf("failed to add IMAP user: %w", err)
		}

		bridge.publish(events.AddressModeChanged{
			UserID:      userID,
			AddressMode: mode,
		})

		return nil
	}, bridge.usersLock)
}

func (bridge *Bridge) loginUser(ctx context.Context, client *proton.Client, authUID, authRef string, keyPass []byte) (string, error) {
	apiUser, err := client.GetUser(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get API user: %w", err)
	}

	salts, err := client.GetSalts(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get key salts: %w", err)
	}

	saltedKeyPass, err := salts.SaltForKey(keyPass, apiUser.Keys.Primary().ID)
	if err != nil {
		return "", fmt.Errorf("failed to salt key password: %w", err)
	}

	if userKR, err := apiUser.Keys.Unlock(saltedKeyPass, nil); err != nil {
		return "", fmt.Errorf("failed to unlock user keys: %w", err)
	} else if userKR.CountDecryptionEntities() == 0 {
		return "", fmt.Errorf("failed to unlock user keys")
	}

	if err := bridge.addUser(ctx, client, apiUser, authUID, authRef, saltedKeyPass, true); err != nil {
		return "", fmt.Errorf("failed to add bridge user: %w", err)
	}

	return apiUser.ID, nil
}

// loadUsers tries to load each user in the vault that isn't already loaded.
func (bridge *Bridge) loadUsers(ctx context.Context) error {
	return bridge.vault.ForUser(runtime.NumCPU(), func(user *vault.User) error {
		if user.AuthUID() == "" {
			return nil
		}

		if safe.RLockRet(func() bool { return mapHas(bridge.users, user.UserID()) }, bridge.usersLock) {
			return nil
		}

		logrus.WithField("userID", user.UserID()).Info("Loading connected user")

		bridge.publish(events.UserLoading{
			UserID: user.UserID(),
		})

		if err := bridge.loadUser(ctx, user); err != nil {
			logrus.WithError(err).Error("Failed to load connected user")

			bridge.publish(events.UserLoadFail{
				UserID: user.UserID(),
				Error:  err,
			})
		} else {
			logrus.WithField("userID", user.UserID()).Info("Successfully loaded user")

			bridge.publish(events.UserLoadSuccess{
				UserID: user.UserID(),
			})
		}

		return nil
	})
}

// loadUser loads an existing user from the vault.
func (bridge *Bridge) loadUser(ctx context.Context, user *vault.User) error {
	client, auth, err := bridge.api.NewClientWithRefresh(ctx, user.AuthUID(), user.AuthRef())
	if err != nil {
		if apiErr := new(proton.Error); errors.As(err, &apiErr) && (apiErr.Code == proton.AuthRefreshTokenInvalid) {
			// The session cannot be refreshed, we sign out the user by clearing his auth secrets.
			if err := user.Clear(); err != nil {
				logrus.WithError(err).Warn("Failed to clear user secrets")
			}
		}
		return fmt.Errorf("failed to create API client: %w", err)
	}

	if err := user.SetAuth(auth.UID, auth.RefreshToken); err != nil {
		return fmt.Errorf("failed to set auth: %w", err)
	}

	apiUser, err := client.GetUser(ctx)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if err := bridge.addUser(ctx, client, apiUser, auth.UID, auth.RefreshToken, user.KeyPass(), false); err != nil {
		return fmt.Errorf("failed to add user: %w", err)
	}

	return nil
}

// addUser adds a new user with an already salted mailbox password.
func (bridge *Bridge) addUser(
	ctx context.Context,
	client *proton.Client,
	apiUser proton.User,
	authUID, authRef string,
	saltedKeyPass []byte,
	isLogin bool,
) error {
	vaultUser, isNew, err := bridge.newVaultUser(apiUser, authUID, authRef, saltedKeyPass)
	if err != nil {
		return fmt.Errorf("failed to add vault user: %w", err)
	}

	if err := bridge.addUserWithVault(ctx, client, apiUser, vaultUser); err != nil {
		if _, ok := err.(*resty.ResponseError); ok || isLogin {
			logrus.WithError(err).Error("Failed to add user, clearing its secrets from vault")

			if err := vaultUser.Clear(); err != nil {
				logrus.WithError(err).Error("Failed to clear user secrets")
			}
		} else {
			logrus.WithError(err).Error("Failed to add user")
		}

		if err := vaultUser.Close(); err != nil {
			logrus.WithError(err).Error("Failed to close vault user")
		}

		if isNew {
			logrus.Warn("Deleting newly added vault user")

			if err := bridge.vault.DeleteUser(apiUser.ID); err != nil {
				logrus.WithError(err).Error("Failed to delete vault user")
			}
		}

		return fmt.Errorf("failed to add user with vault: %w", err)
	}

	return nil
}

// addUserWithVault adds a new user to bridge with the given vault.
func (bridge *Bridge) addUserWithVault(
	ctx context.Context,
	client *proton.Client,
	apiUser proton.User,
	vault *vault.User,
) error {
	user, err := user.New(
		ctx,
		vault,
		client,
		bridge.reporter,
		apiUser,
		bridge.crashHandler,
		bridge.vault.SyncWorkers(),
		bridge.vault.GetShowAllMail(),
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	// Connect the user's address(es) to gluon.
	if err := bridge.addIMAPUser(ctx, user); err != nil {
		return fmt.Errorf("failed to add IMAP user: %w", err)
	}

	// Handle events coming from the user before forwarding them to the bridge.
	// For example, if the user's addresses change, we need to update them in gluon.
	bridge.tasks.Once(func(ctx context.Context) {
		async.RangeContext(ctx, user.GetEventCh(), func(event events.Event) {
			logrus.WithFields(logrus.Fields{
				"userID": apiUser.ID,
				"event":  event,
			}).Debug("Received user event")

			if err := bridge.handleUserEvent(ctx, user, event); err != nil {
				logrus.WithError(err).Error("Failed to handle user event")
			} else {
				bridge.publish(event)
			}
		})
	})

	// Gluon will set the IMAP ID in the context, if known, before making requests on behalf of this user.
	// As such, if we find this ID in the context, we should use it to update our user agent.
	client.AddPreRequestHook(func(_ *resty.Client, r *resty.Request) error {
		if imapID, ok := imap.GetIMAPIDFromContext(r.Context()); ok {
			bridge.identifier.SetClient(imapID.Name, imapID.Version)
		}

		return nil
	})

	// Finally, save the user in the bridge.
	safe.Lock(func() {
		bridge.users[apiUser.ID] = user
	}, bridge.usersLock)

	return nil
}

// newVaultUser creates a new vault user from the given auth information.
// If one already exists in the vault, its data will be updated.
func (bridge *Bridge) newVaultUser(
	apiUser proton.User,
	authUID, authRef string,
	saltedKeyPass []byte,
) (*vault.User, bool, error) {
	if !bridge.vault.HasUser(apiUser.ID) {
		user, err := bridge.vault.AddUser(apiUser.ID, apiUser.Name, authUID, authRef, saltedKeyPass)
		if err != nil {
			return nil, false, fmt.Errorf("failed to add user to vault: %w", err)
		}

		return user, true, nil
	}

	user, err := bridge.vault.NewUser(apiUser.ID)
	if err != nil {
		return nil, false, err
	}

	if err := user.SetAuth(authUID, authRef); err != nil {
		return nil, false, err
	}

	if err := user.SetKeyPass(saltedKeyPass); err != nil {
		return nil, false, err
	}

	return user, false, nil
}

// logout logs out the given user, optionally logging them out from the API too.
func (bridge *Bridge) logoutUser(ctx context.Context, user *user.User, withAPI, withData bool) {
	defer delete(bridge.users, user.ID())

	logrus.WithFields(logrus.Fields{
		"userID":   user.ID(),
		"withAPI":  withAPI,
		"withData": withData,
	}).Debug("Logging out user")

	if err := bridge.removeIMAPUser(ctx, user, withData); err != nil {
		logrus.WithError(err).Error("Failed to remove IMAP user")
	}

	if err := user.Logout(ctx, withAPI); err != nil {
		logrus.WithError(err).Error("Failed to logout user")
	}

	user.Close()
}

// getUserInfo returns information about a disconnected user.
func getUserInfo(userID, username string, state UserState, addressMode vault.AddressMode) UserInfo {
	return UserInfo{
		State:       state,
		UserID:      userID,
		Username:    username,
		AddressMode: addressMode,
	}
}

// getConnUserInfo returns information about a connected user.
func getConnUserInfo(user *user.User) UserInfo {
	return UserInfo{
		State:       Connected,
		UserID:      user.ID(),
		Username:    user.Name(),
		Addresses:   user.Emails(),
		AddressMode: user.GetAddressMode(),
		BridgePass:  user.BridgePass(),
		UsedSpace:   user.UsedSpace(),
		MaxSpace:    user.MaxSpace(),
	}
}

func mapHas[Key comparable, Val any](m map[Key]Val, key Key) bool {
	_, ok := m[key]
	return ok
}
