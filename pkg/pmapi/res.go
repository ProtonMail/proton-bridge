// Copyright (c) 2020 Proton Technologies AG
//
// This file is part of ProtonMail Bridge.
//
// ProtonMail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ProtonMail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ProtonMail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package pmapi

import (
	"net/http"

	"github.com/pkg/errors"
)

// Common response codes.
const (
	CodeOk = 1000
)

// Res is an API response.
type Res struct {
	// The response code is the code from the body JSON. It's still used,
	// but preference is to use HTTP status code instead for new changes.
	Code       int
	StatusCode int

	// The error, if there is any.
	*ResError
}

// Err returns error if the response is an error. Otherwise, returns nil.
func (res Res) Err() error {
	if res.StatusCode == http.StatusUnprocessableEntity {
		return &ErrUnprocessableEntity{errors.New(res.Error)}
	}

	if res.ResError == nil {
		return nil
	}

	if res.Code == ForceUpgradeBadAPIVersion ||
		res.Code == ForceUpgradeInvalidAPI ||
		res.Code == ForceUpgradeBadAppVersion {
		return ErrUpgradeApplication
	}

	if res.Code == APIOffline {
		return ErrAPINotReachable
	}

	return &Error{
		Code:         res.Code,
		ErrorMessage: res.ResError.Error,
	}
}

type ResError struct {
	Error string
}

// Error is an API error.
type Error struct {
	// The error code.
	Code int
	// The error message.
	ErrorMessage string `json:"Error"`
}

func (err Error) Error() string {
	return err.ErrorMessage
}
