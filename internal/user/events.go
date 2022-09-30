package user

import (
	"context"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/proton-bridge/v2/internal/events"
	"github.com/bradenaw/juniper/xslices"
	"gitlab.protontech.ch/go/liteapi"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

// handleAPIEvent handles the given liteapi.Event.
func (user *User) handleAPIEvent(event liteapi.Event) error {
	if event.User != nil {
		if err := user.handleUserEvent(*event.User); err != nil {
			return err
		}
	}

	if len(event.Addresses) > 0 {
		if err := user.handleAddressEvents(event.Addresses); err != nil {
			return err
		}
	}

	if event.MailSettings != nil {
		if err := user.handleMailSettingsEvent(*event.MailSettings); err != nil {
			return err
		}
	}

	if len(event.Labels) > 0 {
		if err := user.handleLabelEvents(event.Labels); err != nil {
			return err
		}
	}

	if len(event.Messages) > 0 {
		if err := user.handleMessageEvents(event.Messages); err != nil {
			return err
		}
	}

	return nil
}

// handleUserEvent handles the given user event.
func (user *User) handleUserEvent(userEvent liteapi.User) error {
	userKR, err := userEvent.Keys.Unlock(user.vault.KeyPass(), nil)
	if err != nil {
		return err
	}

	user.apiUser = userEvent

	user.userKR = userKR

	user.notifyCh <- events.UserChanged{
		UserID: user.ID(),
	}

	return nil
}

// handleAddressEvents handles the given address events.
// TODO: If split address mode, need to signal back to bridge to update the addresses!
func (user *User) handleAddressEvents(addressEvents []liteapi.AddressEvent) error {
	for _, event := range addressEvents {
		switch event.Action {
		case liteapi.EventDelete:
			address, err := user.deleteAddress(event.ID)
			if err != nil {
				return err
			}

			// TODO: This is not the same as addressChangedLogout event!
			// That was only relevant in split mode. This is used differently now.
			user.notifyCh <- events.UserAddressDeleted{
				UserID:  user.ID(),
				Address: address.Email,
			}

		case liteapi.EventCreate:
			if err := user.createAddress(event.Address); err != nil {
				return err
			}

			user.notifyCh <- events.UserAddressCreated{
				UserID:  user.ID(),
				Address: event.Address.Email,
			}

		case liteapi.EventUpdate:
			if err := user.updateAddress(event.Address); err != nil {
				return err
			}

			user.notifyCh <- events.UserAddressChanged{
				UserID:  user.ID(),
				Address: event.Address.Email,
			}
		}
	}

	return nil
}

// createAddress creates the given address.
func (user *User) createAddress(address liteapi.Address) error {
	addrKR, err := address.Keys.Unlock(user.vault.KeyPass(), user.userKR)
	if err != nil {
		return err
	}

	if user.imapConn != nil {
		user.imapConn.addAddress(address.Email)
	}

	user.addresses = append(user.addresses, address)

	user.addrKRs[address.ID] = addrKR

	return nil
}

// updateAddress updates the given address.
func (user *User) updateAddress(address liteapi.Address) error {
	if _, err := user.deleteAddress(address.ID); err != nil {
		return err
	}

	return user.createAddress(address)
}

// deleteAddress deletes the given address.
func (user *User) deleteAddress(addressID string) (liteapi.Address, error) {
	idx := xslices.IndexFunc(user.addresses, func(address liteapi.Address) bool {
		return address.ID == addressID
	})

	if idx < 0 {
		return liteapi.Address{}, ErrNoSuchAddress
	}

	if user.imapConn != nil {
		user.imapConn.remAddress(user.addresses[idx].Email)
	}

	var address liteapi.Address

	address, user.addresses = user.addresses[idx], append(user.addresses[:idx], user.addresses[idx+1:]...)

	delete(user.addrKRs, addressID)

	return address, nil
}

// handleMailSettingsEvent handles the given mail settings event.
func (user *User) handleMailSettingsEvent(mailSettingsEvent liteapi.MailSettings) error {
	user.settings = mailSettingsEvent
	return nil
}

// handleLabelEvents handles the given label events.
func (user *User) handleLabelEvents(labelEvents []liteapi.LabelEvent) error {
	for _, event := range labelEvents {
		switch event.Action {
		case liteapi.EventDelete:
			user.updateCh <- imap.NewMailboxDeleted(imap.LabelID(event.ID))

		case liteapi.EventCreate:
			user.updateCh <- newMailboxCreatedUpdate(imap.LabelID(event.ID), getMailboxName(event.Label))

		case liteapi.EventUpdate, liteapi.EventUpdateFlags:
			user.updateCh <- imap.NewMailboxUpdated(imap.LabelID(event.ID), getMailboxName(event.Label))
		}
	}

	return nil
}

// handleMessageEvents handles the given message events.
func (user *User) handleMessageEvents(messageEvents []liteapi.MessageEvent) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for _, event := range messageEvents {
		switch event.Action {
		case liteapi.EventDelete:
			return ErrNotImplemented

		case liteapi.EventCreate:
			messages, err := user.builder.ProcessAll(ctx, []request{{event.ID, user.addrKRs[event.Message.AddressID]}})
			if err != nil {
				return err
			}

			user.updateCh <- imap.NewMessagesCreated(maps.Values(messages)...)

		case liteapi.EventUpdate, liteapi.EventUpdateFlags:
			user.updateCh <- imap.NewMessageLabelsUpdated(
				imap.MessageID(event.ID),
				imapLabelIDs(filterLabelIDs(event.Message.LabelIDs)),
				bool(!event.Message.Unread),
				slices.Contains(event.Message.LabelIDs, liteapi.StarredLabel),
			)
		}
	}

	return nil
}

func getMailboxName(label liteapi.Label) []string {
	var name []string

	switch label.Type {
	case liteapi.LabelTypeFolder:
		name = []string{folderPrefix, label.Name}

	case liteapi.LabelTypeLabel:
		name = []string{labelPrefix, label.Name}

	default:
		name = []string{label.Name}
	}

	return name
}
