package user

import (
	"context"
	"fmt"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/queue"
	"github.com/ProtonMail/proton-bridge/v2/internal/events"
	"github.com/ProtonMail/proton-bridge/v2/internal/safe"
	"github.com/ProtonMail/proton-bridge/v2/internal/vault"
	"github.com/bradenaw/juniper/xslices"
	"gitlab.protontech.ch/go/liteapi"
)

// handleAPIEvent handles the given liteapi.Event.
func (user *User) handleAPIEvent(ctx context.Context, event liteapi.Event) error {
	if event.User != nil {
		if err := user.handleUserEvent(ctx, *event.User); err != nil {
			return err
		}
	}

	if len(event.Addresses) > 0 {
		if err := user.handleAddressEvents(ctx, event.Addresses); err != nil {
			return err
		}
	}

	if event.MailSettings != nil {
		if err := user.handleMailSettingsEvent(ctx, *event.MailSettings); err != nil {
			return err
		}
	}

	if len(event.Labels) > 0 {
		if err := user.handleLabelEvents(ctx, event.Labels); err != nil {
			return err
		}
	}

	if len(event.Messages) > 0 {
		if err := user.handleMessageEvents(ctx, event.Messages); err != nil {
			return err
		}
	}

	return nil
}

// handleUserEvent handles the given user event.
func (user *User) handleUserEvent(ctx context.Context, userEvent liteapi.User) error {
	userKR, err := userEvent.Keys.Unlock(user.vault.KeyPass(), nil)
	if err != nil {
		return err
	}

	user.apiUser.Set(userEvent)

	user.userKR = userKR

	user.eventCh.Enqueue(events.UserChanged{
		UserID: user.ID(),
	})

	return nil
}

// handleAddressEvents handles the given address events.
// TODO: If split address mode, need to signal back to bridge to update the addresses!
func (user *User) handleAddressEvents(ctx context.Context, addressEvents []liteapi.AddressEvent) error {
	for _, event := range addressEvents {
		switch event.Action {
		case liteapi.EventCreate:
			if err := user.handleCreateAddressEvent(ctx, event); err != nil {
				return fmt.Errorf("failed to handle create address event: %w", err)
			}

		case liteapi.EventUpdate:
			if err := user.handleUpdateAddressEvent(ctx, event); err != nil {
				return fmt.Errorf("failed to handle update address event: %w", err)
			}

		case liteapi.EventDelete:
			if err := user.handleDeleteAddressEvent(ctx, event); err != nil {
				return fmt.Errorf("failed to delete address: %w", err)
			}
		}
	}

	return nil
}

func (user *User) handleCreateAddressEvent(ctx context.Context, event liteapi.AddressEvent) error {
	addrKR, err := event.Address.Keys.Unlock(user.vault.KeyPass(), user.userKR)
	if err != nil {
		return fmt.Errorf("failed to unlock address keys: %w", err)
	}

	apiAddrs, err := user.client.GetAddresses(ctx)
	if err != nil {
		return fmt.Errorf("failed to get addresses: %w", err)
	}

	user.apiAddrs.Set(apiAddrs)

	user.addrKRs[event.Address.ID] = addrKR

	user.eventCh.Enqueue(events.UserAddressCreated{
		UserID:    user.ID(),
		AddressID: event.Address.ID,
		Email:     event.Address.Email,
	})

	if user.vault.AddressMode() == vault.SplitMode {
		user.updateCh[event.Address.ID] = queue.NewQueuedChannel[imap.Update](0, 0)

		if err := syncLabels(ctx, user.client, user.updateCh[event.Address.ID]); err != nil {
			return fmt.Errorf("failed to sync labels to new address: %w", err)
		}
	}

	return nil
}

func (user *User) handleUpdateAddressEvent(ctx context.Context, event liteapi.AddressEvent) error {
	addrKR, err := event.Address.Keys.Unlock(user.vault.KeyPass(), user.userKR)
	if err != nil {
		return fmt.Errorf("failed to unlock address keys: %w", err)
	}

	apiAddrs, err := user.client.GetAddresses(ctx)
	if err != nil {
		return fmt.Errorf("failed to get addresses: %w", err)
	}

	user.apiAddrs.Set(apiAddrs)

	user.addrKRs[event.Address.ID] = addrKR

	user.eventCh.Enqueue(events.UserAddressUpdated{
		UserID:    user.ID(),
		AddressID: event.Address.ID,
		Email:     event.Address.Email,
	})

	return nil
}

func (user *User) handleDeleteAddressEvent(ctx context.Context, event liteapi.AddressEvent) error {
	email, err := safe.GetSliceErr(user.apiAddrs, func(apiAddrs []liteapi.Address) (string, error) {
		return getAddrEmail(apiAddrs, event.ID)
	})
	if err != nil {
		return fmt.Errorf("failed to get address email: %w", err)
	}

	apiAddrs, err := user.client.GetAddresses(ctx)
	if err != nil {
		return fmt.Errorf("failed to get addresses: %w", err)
	}

	user.apiAddrs.Set(apiAddrs)

	delete(user.addrKRs, event.ID)

	if len(user.updateCh) > 1 {
		user.updateCh[event.ID].Close()
		delete(user.updateCh, event.ID)
	}

	user.eventCh.Enqueue(events.UserAddressDeleted{
		UserID:    user.ID(),
		AddressID: event.ID,
		Email:     email,
	})

	return nil
}

// handleMailSettingsEvent handles the given mail settings event.
func (user *User) handleMailSettingsEvent(ctx context.Context, mailSettingsEvent liteapi.MailSettings) error {
	user.settings.Set(mailSettingsEvent)

	return nil
}

// handleLabelEvents handles the given label events.
func (user *User) handleLabelEvents(ctx context.Context, labelEvents []liteapi.LabelEvent) error {
	for _, event := range labelEvents {
		switch event.Action {
		case liteapi.EventCreate:
			if err := user.handleCreateLabelEvent(ctx, event); err != nil {
				return fmt.Errorf("failed to handle create label event: %w", err)
			}

		case liteapi.EventUpdate, liteapi.EventUpdateFlags:
			if err := user.handleUpdateLabelEvent(ctx, event); err != nil {
				return fmt.Errorf("failed to handle update label event: %w", err)
			}

		case liteapi.EventDelete:
			if err := user.handleDeleteLabelEvent(ctx, event); err != nil {
				return fmt.Errorf("failed to handle delete label event: %w", err)
			}
		}
	}

	return nil
}

func (user *User) handleCreateLabelEvent(ctx context.Context, event liteapi.LabelEvent) error {
	for _, updateCh := range user.updateCh {
		updateCh.Enqueue(newMailboxCreatedUpdate(imap.LabelID(event.ID), getMailboxName(event.Label)))
	}

	return nil
}

func (user *User) handleUpdateLabelEvent(ctx context.Context, event liteapi.LabelEvent) error {
	for _, updateCh := range user.updateCh {
		updateCh.Enqueue(imap.NewMailboxUpdated(imap.LabelID(event.ID), getMailboxName(event.Label)))
	}

	return nil
}

func (user *User) handleDeleteLabelEvent(ctx context.Context, event liteapi.LabelEvent) error {
	for _, updateCh := range user.updateCh {
		updateCh.Enqueue(imap.NewMailboxDeleted(imap.LabelID(event.ID)))
	}

	return nil
}

// handleMessageEvents handles the given message events.
func (user *User) handleMessageEvents(ctx context.Context, messageEvents []liteapi.MessageEvent) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for _, event := range messageEvents {
		switch event.Action {
		case liteapi.EventCreate:
			if err := user.handleCreateMessageEvent(ctx, event); err != nil {
				return fmt.Errorf("failed to handle create message event: %w", err)
			}

		case liteapi.EventUpdate, liteapi.EventUpdateFlags:
			if err := user.handleUpdateMessageEvent(ctx, event); err != nil {
				return fmt.Errorf("failed to handle update message event: %w", err)
			}

		case liteapi.EventDelete:
			return ErrNotImplemented
		}
	}

	return nil
}

func (user *User) handleCreateMessageEvent(ctx context.Context, event liteapi.MessageEvent) error {
	full, err := user.client.GetFullMessage(ctx, event.Message.ID)
	if err != nil {
		return fmt.Errorf("failed to get full message: %w", err)
	}

	buildRes, err := buildRFC822(ctx, full, user.addrKRs)
	if err != nil {
		return fmt.Errorf("failed to build RFC822: %w", err)
	}

	if len(user.updateCh) > 1 {
		user.updateCh[buildRes.addressID].Enqueue(imap.NewMessagesCreated(buildRes.update))
	} else {
		user.apiAddrs.Get(func(apiAddrs []liteapi.Address) {
			user.updateCh[apiAddrs[0].ID].Enqueue(imap.NewMessagesCreated(buildRes.update))
		})
	}

	return nil
}

func (user *User) handleUpdateMessageEvent(ctx context.Context, event liteapi.MessageEvent) error {
	update := imap.NewMessageLabelsUpdated(
		imap.MessageID(event.ID),
		mapTo[string, imap.LabelID](xslices.Filter(event.Message.LabelIDs, wantLabelID)),
		event.Message.Seen(),
		event.Message.Starred(),
	)

	if len(user.updateCh) > 1 {
		user.updateCh[event.Message.AddressID].Enqueue(update)
	} else {
		user.apiAddrs.Get(func(apiAddrs []liteapi.Address) {
			user.updateCh[apiAddrs[0].ID].Enqueue(update)
		})
	}

	return nil
}

func getMailboxName(label liteapi.Label) []string {
	var name []string

	switch label.Type {
	case liteapi.LabelTypeFolder:
		name = append([]string{folderPrefix}, label.Path...)

	case liteapi.LabelTypeLabel:
		name = append([]string{labelPrefix}, label.Path...)

	default:
		name = label.Path
	}

	return name
}
