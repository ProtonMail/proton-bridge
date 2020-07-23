package confirmer

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Request struct {
	uuid    string
	value   chan bool
	timeout time.Duration
}

func newRequest(timeout time.Duration) *Request {
	return &Request{
		uuid:    uuid.New().String(),
		value:   make(chan bool),
		timeout: timeout,
	}
}

func (r *Request) ID() string {
	return r.uuid
}

func (r *Request) Result() (bool, error) {
	select {
	case res := <-r.value:
		return res, nil

	case <-time.After(r.timeout):
		return false, errors.New("timed out waiting for result")
	}
}
