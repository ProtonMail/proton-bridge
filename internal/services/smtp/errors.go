// Copyright (c) 2023 Proton AG
//
// This file is part of Proton Mail Bridge.
//
// Proton Mail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Proton Mail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Proton Mail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package smtp

import (
	"errors"
	"fmt"
)

var ErrInvalidRecipient = errors.New("invalid recipient")
var ErrInvalidReturnPath = errors.New("invalid return path")
var ErrNoSuchUser = errors.New("no such user")
var ErrTooManyErrors = errors.New("too many failed requests, please try again later")

type ErrCanNotSendOnAddress struct {
	address string
}

func NewErrCanNotSendOnAddress(address string) *ErrCanNotSendOnAddress {
	return &ErrCanNotSendOnAddress{address: address}
}

func (e ErrCanNotSendOnAddress) Error() string {
	return fmt.Sprintf("can't send on address: %v", e.address)
}
