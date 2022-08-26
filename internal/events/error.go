package events

type Error struct {
	eventBase

	Error error
}
