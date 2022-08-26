package events

type Event interface {
	_isEvent()
}

type eventBase struct{}

func (eventBase) _isEvent() {}
