package events

type MessageSent struct {
	eventBase

	UserID    string
	AddressID string
	MessageID string
}
