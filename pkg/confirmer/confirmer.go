package confirmer

import (
	"errors"
	"sync"
	"time"
)

type Confirmer struct {
	requests map[string]*Request
	locker   sync.Locker
}

func New() *Confirmer {
	return &Confirmer{
		requests: make(map[string]*Request),
		locker:   &sync.Mutex{},
	}
}

func (c *Confirmer) NewRequest(timeout time.Duration) *Request {
	c.locker.Lock()
	defer c.locker.Unlock()

	req := newRequest(timeout)

	c.requests[req.ID()] = req

	return req
}

func (c *Confirmer) SetResponse(uuid string, value bool) error {
	c.locker.Lock()
	defer c.locker.Unlock()

	req, ok := c.requests[uuid]
	if !ok {
		return errors.New("no such request")
	}

	req.value <- value

	return nil
}
