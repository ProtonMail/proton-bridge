package pmapi

type ConnectionObserver interface {
	OnDown()
	OnUp()
}

type observer struct {
	onDown, onUp func()
}

// NewConnectionObserver is a helper function to create a new connection observer from two callbacks.
// It doesn't need to be used; anything which implements the ConnectionObserver interface can be an observer.
func NewConnectionObserver(onDown, onUp func()) ConnectionObserver {
	return &observer{
		onDown: onDown,
		onUp:   onUp,
	}
}

func (o observer) OnDown() { o.onDown() }

func (o observer) OnUp() { o.onUp() }
