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

type mailboxList struct {
	addresses []*mail.Address
}

func (ml *mailboxList) withMailbox(mailbox *mailbox) {
	ml.addresses = append(ml.addresses, &mail.Address{
		Name:    mailbox.name,
		Address: mailbox.address,
	})
}

func (ml *mailboxList) withObsMboxList(obsMboxList *obsMboxList) {
	ml.addresses = append(ml.addresses, obsMboxList.addresses...)
}

func (w *walker) EnterMailboxList(ctx *parser.MailboxListContext) {
	logrus.WithField("text", ctx.GetText()).Trace("Entering mailboxList")
	w.enter(&mailboxList{})
}

func (w *walker) ExitMailboxList(ctx *parser.MailboxListContext) {
	logrus.WithField("text", ctx.GetText()).Trace("Exiting mailboxList")

	type withMailboxList interface {
		withMailboxList(*mailboxList)
	}

	res := w.exit().(*mailboxList)

	if parent, ok := w.parent().(withMailboxList); ok {
		parent.withMailboxList(res)
	}
}
