package events

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

	UserID  string
	Address string
}

type UserAddressChanged struct {
	eventBase

	UserID  string
	Address string
}

type UserAddressDeleted struct {
	eventBase

	UserID  string
	Address string
}
