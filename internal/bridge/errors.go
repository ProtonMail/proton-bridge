package bridge

import "errors"

var (
	ErrServeIMAP    = errors.New("failed to serve IMAP")
	ErrServeSMTP    = errors.New("failed to serve SMTP")
	ErrWatchUpdates = errors.New("failed to watch for updates")

	ErrNoSuchUser          = errors.New("no such user")
	ErrUserAlreadyExists   = errors.New("user already exists")
	ErrUserAlreadyLoggedIn = errors.New("user already logged in")
	ErrNotImplemented      = errors.New("not implemented")

	ErrSizeTooLarge = errors.New("file is too big")
)
