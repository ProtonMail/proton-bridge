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

package transfer

// imapError is base for all IMAP errors.
type imapError struct {
	Message string
	Err     error
}

func (e imapError) Error() string {
	return e.Message + ": " + e.Err.Error()
}

func (e imapError) Unwrap() error {
	return e.Err
}

func (e imapError) Cause() error {
	return e.Err
}

// ErrIMAPConnection is error representing connection issues.
type ErrIMAPConnection struct {
	imapError
}

func (e ErrIMAPConnection) Is(target error) bool {
	_, ok := target.(*ErrIMAPConnection)
	return ok
}

// ErrIMAPAuth is error representing authentication issues.
type ErrIMAPAuth struct {
	imapError
}

func (e ErrIMAPAuth) Is(target error) bool {
	_, ok := target.(*ErrIMAPAuth)
	return ok
}

// ErrIMAPAuthMethod is error representing wrong auth method.
type ErrIMAPAuthMethod struct {
	imapError
}

func (e ErrIMAPAuthMethod) Is(target error) bool {
	_, ok := target.(*ErrIMAPAuthMethod)
	return ok
}
