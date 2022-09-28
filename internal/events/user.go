package events

import "github.com/ProtonMail/proton-bridge/v2/internal/vault"

type UserLoggedIn struct {
	eventBase

	UserID string
}

type UserLoggedOut struct {
	eventBase

	UserID string
}

type UserDeauth struct {
	eventBase

	UserID string
}

type UserDeleted struct {
	eventBase

	UserID string
}

type UserChanged struct {
	eventBase

	UserID string
}

type UserAddressCreated struct {
	eventBase

	UserID    string
	AddressID string
	Email     string
}

type UserAddressUpdated struct {
	eventBase

	UserID    string
	AddressID string
	Email     string
}

type UserAddressDeleted struct {
	eventBase

	UserID    string
	AddressID string
	Email     string
}

type AddressModeChanged struct {
	eventBase

	UserID      string
	AddressMode vault.AddressMode
}
