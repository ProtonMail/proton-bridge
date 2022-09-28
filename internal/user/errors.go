package user

import "errors"

var (
	ErrNoSuchAddress     = errors.New("no such address")
	ErrNotImplemented    = errors.New("not implemented")
	ErrNotSupported      = errors.New("not supported")
	ErrInvalidReturnPath = errors.New("invalid return path")
	ErrInvalidRecipient  = errors.New("invalid recipient")
	ErrMissingAddrKey    = errors.New("missing address key")
)
