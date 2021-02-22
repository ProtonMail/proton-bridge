package pmapi

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
)

type Error struct {
	Code    int
	Message string `json:"Error"`
}

func (err Error) Error() string {
	return err.Message
}

func catchAPIError(_ *resty.Client, res *resty.Response) error {
	if !res.IsError() {
		return nil
	}

	var err error

	if apiErr, ok := res.Error().(*Error); ok {
		err = apiErr
	} else {
		err = errors.New(res.Status())
	}

	switch res.StatusCode() {
	case http.StatusUnauthorized:
		return errors.Wrap(ErrUnauthorized, err.Error())

	default:
		return errors.Wrap(ErrAPIFailure, err.Error())
	}
}

func catchRetryAfter(_ *resty.Client, res *resty.Response) (time.Duration, error) {
	if res.StatusCode() == http.StatusTooManyRequests {
		if after := res.Header().Get("Retry-After"); after != "" {
			seconds, err := strconv.Atoi(after)
			if err != nil {
				return 0, err
			}

			return time.Duration(seconds) * time.Second, nil
		}
	}

	return 0, nil
}

func catchTooManyRequests(res *resty.Response, _ error) bool {
	return res.StatusCode() == http.StatusTooManyRequests
}

func catchNoResponse(res *resty.Response, err error) bool {
	return res.RawResponse == nil && err != nil
}

func catchProxyAvailable(res *resty.Response, err error) bool {
	/*
		if res.Request.Attempt < ... {
			return false
		}

		if response is not empty {
			return false
		}

		if proxy is available {
			return true
		}
	*/

	return false
}
