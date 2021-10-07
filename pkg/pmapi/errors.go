// Copyright (c) 2021 Proton Technologies AG
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

import "errors"

var (
	ErrNoConnection       = errors.New("no internet connection")
	ErrUnauthorized       = errors.New("API client is unauthorized")
	ErrUpgradeApplication = errors.New("application upgrade required")

	ErrBad2FACode         = errors.New("incorrect 2FA code")
	ErrBad2FACodeTryAgain = errors.New("incorrect 2FA code: please try again")

	ErrPaidPlanRequired = errors.New("paid subscription plan is required")
	ErrPasswordWrong    = errors.New("wrong password")
)

type ErrUnprocessableEntity struct {
	originalError error
}

func (err ErrUnprocessableEntity) Error() string {
	return err.originalError.Error()
}
