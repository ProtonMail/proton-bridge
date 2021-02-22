package pmapi

import "errors"

var (
	ErrNoConnection = errors.New("no internet connection")
	ErrAPIFailure   = errors.New("API returned an error")
	ErrUnauthorized = errors.New("API client is unauthorized")
)
