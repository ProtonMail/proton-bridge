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

type address struct {
	addresses []*mail.Address
}

func (a *address) withMailbox(mailbox *mailbox) {
	a.addresses = append(a.addresses, &mail.Address{
		Name:    mailbox.name,
		Address: mailbox.address,
	})
}

func (a *address) withGroup(group *group) {
	a.addresses = append(a.addresses, group.addresses...)
}

func (w *walker) EnterAddress(ctx *parser.AddressContext) {
	logrus.WithField("text", ctx.GetText()).Trace("Entering address")
	w.enter(&address{})
}

func (w *walker) ExitAddress(ctx *parser.AddressContext) {
	logrus.WithField("text", ctx.GetText()).Trace("Exiting address")

	type withAddress interface {
		withAddress(*address)
	}

	res := w.exit().(*address)

	if parent, ok := w.parent().(withAddress); ok {
		parent.withAddress(res)
	}
}
