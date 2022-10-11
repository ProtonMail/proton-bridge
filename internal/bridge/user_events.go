package bridge

import (
	"context"
	"fmt"

	"github.com/ProtonMail/proton-bridge/v2/internal/events"
	"github.com/ProtonMail/proton-bridge/v2/internal/user"
	"github.com/ProtonMail/proton-bridge/v2/internal/vault"
)

func (bridge *Bridge) handleUserEvent(ctx context.Context, user *user.User, event events.Event) error {
	switch event := event.(type) {
	case events.UserAddressCreated:
		if err := bridge.handleUserAddressCreated(ctx, user, event); err != nil {
			return fmt.Errorf("failed to handle user address created event: %w", err)
		}

	case events.UserAddressUpdated:
		if err := bridge.handleUserAddressUpdated(ctx, user, event); err != nil {
			return fmt.Errorf("failed to handle user address updated event: %w", err)
		}

	case events.UserAddressDeleted:
		if err := bridge.handleUserAddressDeleted(ctx, user, event); err != nil {
			return fmt.Errorf("failed to handle user address deleted event: %w", err)
		}

	case events.UserDeauth:
		if err := bridge.logoutUser(context.Background(), event.UserID); err != nil {
			return fmt.Errorf("failed to logout user: %w", err)
		}
	}

	return nil
}

func (bridge *Bridge) handleUserAddressCreated(ctx context.Context, user *user.User, event events.UserAddressCreated) error {
	switch user.GetAddressMode() {
	case vault.CombinedMode:
		for addrID, gluonID := range user.GetGluonIDs() {
			if err := bridge.imapServer.RemoveUser(ctx, gluonID, false); err != nil {
				return fmt.Errorf("failed to remove user from IMAP server: %w", err)
			}

			imapConn, err := user.NewIMAPConnector(addrID)
			if err != nil {
				return fmt.Errorf("failed to create IMAP connector: %w", err)
			}

			if err := bridge.imapServer.LoadUser(ctx, imapConn, gluonID, user.GluonKey()); err != nil {
				return fmt.Errorf("failed to add user to IMAP server: %w", err)
			}
		}

	case vault.SplitMode:
		imapConn, err := user.NewIMAPConnector(event.AddressID)
		if err != nil {
			return fmt.Errorf("failed to create IMAP connector: %w", err)
		}

		gluonID, err := bridge.imapServer.AddUser(ctx, imapConn, user.GluonKey())
		if err != nil {
			return fmt.Errorf("failed to add user to IMAP server: %w", err)
		}

		if err := user.SetGluonID(event.AddressID, gluonID); err != nil {
			return fmt.Errorf("failed to set gluon ID: %w", err)
		}
	}

	return nil
}

// TODO: Handle addresses that have been disabled!
func (bridge *Bridge) handleUserAddressUpdated(_ context.Context, user *user.User, _ events.UserAddressUpdated) error {
	switch user.GetAddressMode() {
	case vault.CombinedMode:
		return fmt.Errorf("not implemented")

	case vault.SplitMode:
		return fmt.Errorf("not implemented")
	}

	return nil
}

func (bridge *Bridge) handleUserAddressDeleted(ctx context.Context, user *user.User, event events.UserAddressDeleted) error {
	switch user.GetAddressMode() {
	case vault.CombinedMode:
		for addrID, gluonID := range user.GetGluonIDs() {
			if err := bridge.imapServer.RemoveUser(ctx, gluonID, false); err != nil {
				return fmt.Errorf("failed to remove user from IMAP server: %w", err)
			}

			imapConn, err := user.NewIMAPConnector(addrID)
			if err != nil {
				return fmt.Errorf("failed to create IMAP connector: %w", err)
			}

			if err := bridge.imapServer.LoadUser(ctx, imapConn, gluonID, user.GluonKey()); err != nil {
				return fmt.Errorf("failed to add user to IMAP server: %w", err)
			}
		}

	case vault.SplitMode:
		gluonID, ok := user.GetGluonID(event.AddressID)
		if !ok {
			return fmt.Errorf("gluon ID not found for address %s", event.AddressID)
		}

		if err := bridge.imapServer.RemoveUser(ctx, gluonID, true); err != nil {
			return fmt.Errorf("failed to remove user from IMAP server: %w", err)
		}
	}

	return nil
}
