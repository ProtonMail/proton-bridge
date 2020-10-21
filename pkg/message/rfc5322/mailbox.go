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
	"fmt"

	"github.com/ProtonMail/proton-bridge/pkg/message/rfc5322/parser"
	"github.com/sirupsen/logrus"
)

type mailbox struct {
	name, address string
}

func (m *mailbox) withNameAddr(nameAddr *nameAddr) {
	m.name = nameAddr.name
	m.address = nameAddr.address
}

func (m *mailbox) withAddrSpec(addrSpec *addrSpec) {
	m.address = fmt.Sprintf("%v@%v", addrSpec.localPart, addrSpec.domain)
}

func (w *walker) EnterMailbox(ctx *parser.MailboxContext) {
	logrus.WithField("text", ctx.GetText()).Trace("Entering mailbox")
	w.enter(&mailbox{})
}

func (w *walker) ExitMailbox(ctx *parser.MailboxContext) {
	logrus.WithField("text", ctx.GetText()).Trace("Exiting mailbox")

	type withMailbox interface {
		withMailbox(*mailbox)
	}

	res := w.exit().(*mailbox)

	if parent, ok := w.parent().(withMailbox); ok {
		parent.withMailbox(res)
	}
}
