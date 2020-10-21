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

package rfc5322

import (
	"net/mail"

	"github.com/ProtonMail/proton-bridge/pkg/message/rfc5322/parser"
	"github.com/sirupsen/logrus"
)

type addressList struct {
	addresses []*mail.Address
}

func (a *addressList) withAddress(address *address) {
	a.addresses = append(a.addresses, address.addresses...)
}

func (w *walker) EnterAddressList(ctx *parser.AddressListContext) {
	logrus.WithField("text", ctx.GetText()).Trace("Entering addressList")
	w.enter(&addressList{})
}

func (w *walker) ExitAddressList(ctx *parser.AddressListContext) {
	logrus.WithField("text", ctx.GetText()).Trace("Exiting addressList")
	w.res = w.exit().(*addressList).addresses
}
