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
// along with Proton Mail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package bridge

import (
	"context"
	"fmt"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/proton-bridge/v2/internal/async"
	"github.com/ProtonMail/proton-bridge/v2/internal/events"
	"github.com/ProtonMail/proton-bridge/v2/internal/safe"
	"github.com/ProtonMail/proton-bridge/v2/internal/try"
	"github.com/ProtonMail/proton-bridge/v2/internal/user"
	"github.com/ProtonMail/proton-bridge/v2/internal/vault"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
	"gitlab.protontech.ch/go/liteapi"
)

type UserInfo struct {
	// UserID is the user's API ID.
	UserID string

	// Username is the user's API username.
	Username string

	// Connected is true if the user is logged in (has API auth).
	Connected bool

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

// GetUserInfo returns info about the given user.
func (bridge *Bridge) GetUserInfo(userID string) (UserInfo, error) {
	if info, ok := safe.MapGetRet(bridge.users, userID, getConnUserInfo); ok {
		return info, nil
	}

	var info UserInfo

	if err := bridge.vault.GetUser(userID, func(user *vault.User) {
		info = getUserInfo(user.UserID(), user.Username(), user.AddressMode())
	}); err != nil {
		return UserInfo{}, fmt.Errorf("failed to get user info: %w", err)
	}

	return info, nil
}

// QueryUserInfo queries the user info by username or address.
func (bridge *Bridge) QueryUserInfo(query string) (UserInfo, error) {
	return safe.MapValuesRetErr(bridge.users, func(users []*user.User) (UserInfo, error) {
		for _, user := range users {
			if user.Match(query) {
				return getConnUserInfo(user), nil
			}
		}

		return UserInfo{}, ErrNoSuchUser
	})
}

// LoginAuth begins the login process. It returns an authorized client that might need 2FA.
func (bridge *Bridge) LoginAuth(ctx context.Context, username string, password []byte) (*liteapi.Client, liteapi.Auth, error) {
	client, auth, err := bridge.api.NewClientWithLogin(ctx, username, password)
	if err != nil {
		return nil, liteapi.Auth{}, fmt.Errorf("failed to create new API client: %w", err)
	}

	if bridge.users.Has(auth.UserID) {
		if err := client.AuthDelete(ctx); err != nil {
			logrus.WithError(err).Warn("Failed to delete auth")
		}

		return nil, liteapi.Auth{}, ErrUserAlreadyLoggedIn
	}

	return client, auth, nil
}

// LoginUser finishes the user login process using the client and auth received from LoginAuth.
func (bridge *Bridge) LoginUser(
	ctx context.Context,
	client *liteapi.Client,
	auth liteapi.Auth,
	keyPass []byte,
) (string, error) {
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
	client, auth, err := bridge.LoginAuth(ctx, username, password)
	if err != nil {
		return "", fmt.Errorf("failed to begin login process: %w", err)
	}

	if auth.TwoFA.Enabled == liteapi.TOTPEnabled {
		totp, err := getTOTP()
		if err != nil {
			return "", fmt.Errorf("failed to get TOTP: %w", err)
		}

		if err := client.Auth2FA(ctx, liteapi.Auth2FAReq{TwoFactorCode: totp}); err != nil {
			return "", fmt.Errorf("failed to authorize 2FA: %w", err)
		}
	}

	var keyPass []byte

	if auth.PasswordMode == liteapi.TwoPasswordMode {
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
	if err := bridge.logoutUser(ctx, userID); err != nil {
		return fmt.Errorf("failed to logout user: %w", err)
	}

	bridge.publish(events.UserLoggedOut{
		UserID: userID,
	})

	return nil
}

// DeleteUser deletes the given user.
func (bridge *Bridge) DeleteUser(ctx context.Context, userID string) error {
	bridge.deleteUser(ctx, userID)

	bridge.publish(events.UserDeleted{
		UserID: userID,
	})

	return nil
}

// SetAddressMode sets the address mode for the given user.
func (bridge *Bridge) SetAddressMode(ctx context.Context, userID string, mode vault.AddressMode) error {
	if ok, err := bridge.users.GetErr(userID, func(user *user.User) error {
		if user.GetAddressMode() == mode {
			return fmt.Errorf("address mode is already %q", mode)
		}

		for _, gluonID := range user.GetGluonIDs() {
			if err := bridge.imapServer.RemoveUser(ctx, gluonID, true); err != nil {
				return fmt.Errorf("failed to remove user from IMAP server: %w", err)
			}
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
	}); !ok {
		return ErrNoSuchUser
	} else if err != nil {
		return fmt.Errorf("failed to set address mode: %w", err)
	}

	return nil
}

func (bridge *Bridge) loginUser(ctx context.Context, client *liteapi.Client, authUID, authRef string, keyPass []byte) (string, error) {
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

	if err := bridge.addUser(ctx, client, apiUser, authUID, authRef, saltedKeyPass, true); err != nil {
		return "", fmt.Errorf("failed to add bridge user: %w", err)
	}

	return apiUser.ID, nil
}

// loadUsers tries to load each user in the vault that isn't already loaded.
func (bridge *Bridge) loadUsers(ctx context.Context) error {
	return bridge.vault.ForUser(func(user *vault.User) error {
		if bridge.users.Has(user.UserID()) || user.AuthUID() == "" {
			return nil
		}

		if err := bridge.loadUser(ctx, user); err != nil {
			logrus.WithError(err).Error("Failed to load connected user")
		} else {
			bridge.publish(events.UserLoaded{
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
		return fmt.Errorf("failed to create API client: %w", err)
	}

	if err := user.SetAuth(auth.UID, auth.RefreshToken); err != nil {
		return fmt.Errorf("failed to set auth: %w", err)
	}

	return try.Catch(
		func() error {
			apiUser, err := client.GetUser(ctx)
			if err != nil {
				return fmt.Errorf("failed to get user: %w", err)
			}

			if err := bridge.addUser(ctx, client, apiUser, auth.UID, auth.RefreshToken, user.KeyPass(), false); err != nil {
				return fmt.Errorf("failed to add user: %w", err)
			}

			return nil
		},
	)
}

// addUser adds a new user with an already salted mailbox password.
func (bridge *Bridge) addUser(
	ctx context.Context,
	client *liteapi.Client,
	apiUser liteapi.User,
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
	client *liteapi.Client,
	apiUser liteapi.User,
	vaultUser *vault.User,
) error {
	user, err := user.New(ctx, vaultUser, client, apiUser, bridge.GetShowAllMail())
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	if had := bridge.users.Set(apiUser.ID, user); had {
		panic("double add")
	}

	// Connect the user's address(es) to gluon.
	if err := bridge.addIMAPUser(ctx, user); err != nil {
		return fmt.Errorf("failed to add IMAP user: %w", err)
	}

	// Handle events coming from the user before forwarding them to the bridge.
	// For example, if the user's addresses change, we need to update them in gluon.
	bridge.tasks.Once(func(ctx context.Context) {
		async.RangeContext(ctx, user.GetEventCh(), func(event events.Event) {
			if err := bridge.handleUserEvent(ctx, user, event); err != nil {
				logrus.WithError(err).Error("Failed to handle user event")
			} else {
				bridge.publish(event)
			}
		})
	})

	// Gluon will set the IMAP ID in the context, if known, before making requests on behalf of this user.
	// As such, if we find this ID in the context, we should use it to update our user agent.
	client.AddPreRequestHook(func(ctx context.Context, req *resty.Request) error {
		if imapID, ok := imap.GetIMAPIDFromContext(ctx); ok {
			bridge.identifier.SetClient(imapID.Name, imapID.Version)
		}

		return nil
	})

	return nil
}

// newVaultUser creates a new vault user from the given auth information.
// If one already exists in the vault, its data will be updated.
func (bridge *Bridge) newVaultUser(
	apiUser liteapi.User,
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

// addIMAPUser connects the given user to gluon.
func (bridge *Bridge) addIMAPUser(ctx context.Context, user *user.User) error {
	imapConn, err := user.NewIMAPConnectors()
	if err != nil {
		return fmt.Errorf("failed to create IMAP connectors: %w", err)
	}

	for addrID, imapConn := range imapConn {
		if gluonID, ok := user.GetGluonID(addrID); ok {
			if err := bridge.imapServer.LoadUser(ctx, imapConn, gluonID, user.GluonKey()); err != nil {
				return fmt.Errorf("failed to load IMAP user: %w", err)
			}
		} else {
			gluonID, err := bridge.imapServer.AddUser(ctx, imapConn, user.GluonKey())
			if err != nil {
				return fmt.Errorf("failed to add IMAP user: %w", err)
			}

			if err := user.SetGluonID(addrID, gluonID); err != nil {
				return fmt.Errorf("failed to set IMAP user ID: %w", err)
			}
		}
	}

	return nil
}

// logoutUser logs the given user out from bridge.
func (bridge *Bridge) logoutUser(ctx context.Context, userID string) error {
	if ok := bridge.users.GetDelete(userID, func(user *user.User) {
		for _, gluonID := range user.GetGluonIDs() {
			if err := bridge.imapServer.RemoveUser(ctx, gluonID, false); err != nil {
				logrus.WithError(err).Error("Failed to remove IMAP user")
			}
		}

		if err := user.Logout(ctx); err != nil {
			logrus.WithError(err).Error("Failed to logout user")
		}

		user.Close()
	}); !ok {
		return ErrNoSuchUser
	}

	return nil
}

// deleteUser deletes the given user from bridge.
func (bridge *Bridge) deleteUser(ctx context.Context, userID string) {
	if ok := bridge.users.GetDelete(userID, func(user *user.User) {
		for _, gluonID := range user.GetGluonIDs() {
			if err := bridge.imapServer.RemoveUser(ctx, gluonID, true); err != nil {
				logrus.WithError(err).Error("Failed to remove IMAP user")
			}
		}

		if err := user.Logout(ctx); err != nil {
			logrus.WithError(err).Error("Failed to logout user")
		}

		user.Close()
	}); !ok {
		logrus.Debug("The bridge user was not connected")
	}

	if err := bridge.vault.DeleteUser(userID); err != nil {
		logrus.WithError(err).Error("Failed to delete user from vault")
	}
}

// getUserInfo returns information about a disconnected user.
func getUserInfo(userID, username string, addressMode vault.AddressMode) UserInfo {
	return UserInfo{
		UserID:      userID,
		Username:    username,
		AddressMode: addressMode,
	}
}

// getConnUserInfo returns information about a connected user.
func getConnUserInfo(user *user.User) UserInfo {
	return UserInfo{
		Connected:   true,
		UserID:      user.ID(),
		Username:    user.Name(),
		Addresses:   user.Emails(),
		AddressMode: user.GetAddressMode(),
		BridgePass:  user.BridgePass(),
		UsedSpace:   user.UsedSpace(),
		MaxSpace:    user.MaxSpace(),
	}
}
